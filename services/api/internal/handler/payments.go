package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/payment"
)

// orderLocks serializes settlement per order so the webhook and the status
// poller can never activate the same order twice in this process.
var orderLocks sync.Map

func lockOrder(id uuid.UUID) func() {
	v, _ := orderLocks.LoadOrStore(id, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

type CheckoutRequest struct {
	PlanID         string `json:"plan_id"`
	BillingCycle   string `json:"billing_cycle"` // 'monthly' or 'annual'
	CryptoCurrency string `json:"currency"`      // coin id from GET /payments/currencies
}

// GetCryptoCurrencies GET /payments/currencies — live list of payable coins,
// fetched server-side so the gateway key/host never reach the browser.
func (h *Handler) GetCryptoCurrencies(w http.ResponseWriter, r *http.Request) {
	var coins []payment.CryptoCurrency
	var err error
	if h.gateway != nil {
		coins, err = h.gateway.FetchCurrencies(r.Context())
	}
	if err != nil || len(coins) == 0 {
		coins = payment.GetDefaultCurrencies()
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"currencies": coins,
	})
}

// CheckoutSubscription POST /subscriptions/checkout — creates a pending order
// and a white-label crypto invoice for the chosen role plan + coin.
// Enforces strict anti-abuse, active order reuse, concurrency locking, and 5x/day limit.
func (h *Handler) CheckoutSubscription(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 1. IP-level rate limiting (max 20 checkout attempts/min per IP)
	if h.rdb != nil {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
		ipKey := fmt.Sprintf("ratelimit:checkout:ip:%s", ip)
		cnt, err := h.rdb.Incr(r.Context(), ipKey).Result()
		if err == nil {
			if cnt == 1 {
				h.rdb.Expire(r.Context(), ipKey, time.Minute)
			}
			if cnt > 20 {
				writeError(w, http.StatusTooManyRequests, "Too many requests from this network. Please wait a moment.")
				return
			}
		}
	}

	// 2. Concurrency lock per user (prevents race condition & duplicate generation)
	if h.rdb != nil {
		lockKey := fmt.Sprintf("lock:checkout:usr:%s", user.ID.String())
		acquired, err := h.rdb.SetNX(r.Context(), lockKey, "1", 10*time.Second).Result()
		if err == nil && !acquired {
			writeError(w, http.StatusTooManyRequests, "Another checkout request is currently being processed. Please wait.")
			return
		}
		defer h.rdb.Del(r.Context(), lockKey)
	}

	// 3. User-level cooldown (minimum 3 seconds between new invoice generation attempts)
	if h.rdb != nil {
		cooldownKey := fmt.Sprintf("cooldown:checkout:usr:%s", user.ID.String())
		if ex, _ := h.rdb.Exists(r.Context(), cooldownKey).Result(); ex > 0 {
			writeError(w, http.StatusTooManyRequests, "Please wait a few seconds before generating another invoice.")
			return
		}
	}

	// 4. Existing Purchase Detection (Reuse active transaction if pending)
	activeOrder, err := h.repo.GetActivePaymentOrder(r.Context(), user.ID)
	if err == nil && activeOrder != nil {
		inv := invoiceFromMetadata(activeOrder.Metadata)
		if time.Since(activeOrder.CreatedAt) > 2*time.Hour || (inv != nil && expired(inv.ExpireUTC)) {
			h.closeOrder(r, activeOrder.ID, "expired")
		} else {
			var op *payment.CryptoOperation
			if h.gateway != nil && h.gateway.IsConfigured() && activeOrder.ProviderOrderID != "" {
				if live, err := h.gateway.GetOperation(r.Context(), activeOrder.ProviderOrderID); err == nil {
					op = live
					switch op.Status {
					case "completed", "mismatch":
						h.settleOrder(r, activeOrder.ID)
						writeError(w, http.StatusBadRequest, "Your existing order has already been paid and activated.")
						return
					case "expired", "cancelled", "error":
						h.closeOrder(r, activeOrder.ID, op.Status)
						activeOrder = nil
					}
				}
			}
			if activeOrder != nil {
				res := h.buildOrderStatus(activeOrder, h.planDisplayName(r, activeOrder.PlanID), inv, op)
				res["is_existing"] = true
				res["message"] = "An active pending order already exists. Please complete payment or cancel it to select a different plan."
				writeJSON(w, http.StatusOK, res)
				return
			}
		}
	}

	// 5. Strict Daily Generation Limit (Max 5 transactions per calendar day in Asia/Jakarta)
	dailyCount, err := h.repo.CountDailyPaymentOrders(r.Context(), user.ID)
	if err == nil && dailyCount >= 5 {
		writeError(w, http.StatusTooManyRequests, "Daily purchase limit reached (5/5). Quota resets at 00:00 WIB.")
		return
	}

	// 6. Safe body decoding with 8KB payload cap
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.CryptoCurrency = strings.TrimSpace(req.CryptoCurrency)
	if req.PlanID == "" {
		req.PlanID = "vip"
	}
	if req.CryptoCurrency == "" {
		req.CryptoCurrency = "USDT_TRX"
	}
	if req.BillingCycle != "annual" {
		req.BillingCycle = "monthly"
	}

	var planID, planName, planCurrency string
	var amountCents int

	if memPlan, err := h.repo.GetMembershipPlanByID(r.Context(), req.PlanID); err == nil && memPlan != nil {
		if !memPlan.IsActive || memPlan.Key == "free" || memPlan.PriceAmountCents <= 0 {
			writeError(w, http.StatusBadRequest, "selected plan is not available for purchase")
			return
		}
		planID = memPlan.ID
		planName = memPlan.Name
		planCurrency = memPlan.Currency
		amountCents = memPlan.PriceAmountCents
		if req.BillingCycle == "annual" {
			amountCents = memPlan.PriceAmountCents * 12
		}
	} else {
		plan, err := h.repo.GetSubscriptionPlanByID(r.Context(), req.PlanID)
		if err != nil {
			writeError(w, http.StatusNotFound, "subscription plan not found")
			return
		}
		planID = plan.ID
		planName = plan.Name
		planCurrency = plan.Currency
		amountCents = plan.MonthlyPriceCents
		if req.BillingCycle == "annual" {
			amountCents = plan.AnnualPriceCents
		}
	}
	if planCurrency == "" {
		planCurrency = "USD"
	}

	order := &domain.PaymentOrder{
		ID:           uuid.New(),
		UserID:       user.ID,
		PlanID:       planID,
		BillingCycle: req.BillingCycle,
		AmountCents:  amountCents,
		Currency:     planCurrency,
		Provider:     "crypto",
		Status:       "pending",
	}

	var inv *payment.CryptoInvoice
	switch {
	case h.gateway != nil && h.gateway.IsConfigured():
		if !h.gateway.IsSupportedCurrency(r.Context(), req.CryptoCurrency) {
			writeError(w, http.StatusBadRequest, "selected coin is not supported")
			return
		}
		created, err := h.gateway.CreateInvoice(r.Context(), order.ID, planName, amountCents, req.CryptoCurrency, user.Email)
		if err != nil {
			slog.Error("payments.invoice_failed", "order_id", order.ID, "coin", req.CryptoCurrency, "error", err.Error())
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		inv = created
	default:
		if h.gateway != nil {
			inv = h.gateway.CreateSimulatedInvoice(amountCents, req.CryptoCurrency)
		} else {
			inv = (&payment.CryptoGateway{}).CreateSimulatedInvoice(amountCents, req.CryptoCurrency)
		}
	}

	order.ProviderOrderID = inv.TxnID
	order.Metadata = map[string]interface{}{"payment": inv}

	if err := h.repo.CreatePaymentOrder(r.Context(), order); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment order")
		return
	}

	// Set 3-second cooldown
	if h.rdb != nil {
		cooldownKey := fmt.Sprintf("cooldown:checkout:usr:%s", user.ID.String())
		h.rdb.Set(r.Context(), cooldownKey, "1", 3*time.Second)
	}

	slog.Info("payments.order_created", "order_id", order.ID, "plan", planID, "coin", inv.Currency, "daily_count", dailyCount+1)
	writeJSON(w, http.StatusCreated, h.buildOrderStatus(order, planName, inv, nil))
}

// GetActiveSubscriptionOrder GET /subscriptions/orders/active — returns user's active pending order if any.
func (h *Handler) GetActiveSubscriptionOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	order, err := h.repo.GetActivePaymentOrder(r.Context(), user.ID)
	if err != nil || order == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"order": nil})
		return
	}

	inv := invoiceFromMetadata(order.Metadata)
	if time.Since(order.CreatedAt) > 2*time.Hour || (inv != nil && expired(inv.ExpireUTC)) {
		h.closeOrder(r, order.ID, "expired")
		writeJSON(w, http.StatusOK, map[string]interface{}{"order": nil})
		return
	}

	var op *payment.CryptoOperation
	if h.gateway != nil && h.gateway.IsConfigured() && order.ProviderOrderID != "" {
		if live, err := h.gateway.GetOperation(r.Context(), order.ProviderOrderID); err == nil {
			op = live
			switch op.Status {
			case "completed", "mismatch":
				h.settleOrder(r, order.ID)
				writeJSON(w, http.StatusOK, map[string]interface{}{"order": nil})
				return
			case "expired", "cancelled", "error":
				h.closeOrder(r, order.ID, op.Status)
				writeJSON(w, http.StatusOK, map[string]interface{}{"order": nil})
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order": h.buildOrderStatus(order, h.planDisplayName(r, order.PlanID), inv, op),
	})
}

// CancelSubscriptionOrder POST /subscriptions/orders/{orderId}/cancel — securely cancels an active pending order.
func (h *Handler) CancelSubscriptionOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := uuid.Parse(chi.URLParam(r, "orderId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if order.UserID != user.ID {
		writeError(w, http.StatusForbidden, "forbidden: cannot cancel an order belonging to another user")
		return
	}
	if order.Status != "pending" {
		writeError(w, http.StatusBadRequest, "only pending orders can be cancelled; completed or settled orders cannot be cancelled")
		return
	}

	unlock := lockOrder(order.ID)
	defer unlock()

	if err := h.repo.CancelPaymentOrder(r.Context(), order.ID, user.ID); err != nil {
		writeError(w, http.StatusBadRequest, "order cannot be cancelled")
		return
	}

	slog.Info("payments.order_cancelled_by_user", "order_id", order.ID, "user_id", user.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"message":  "Order successfully cancelled",
		"order_id": order.ID.String(),
	})
}

// GetSubscriptionOrder GET /subscriptions/orders/{orderId} — payment status
// for the checkout popup (pending / partial / processing / completed / expired).
// While pending, the live gateway operation is reconciled so the UI updates
// even when the webhook has not arrived yet.
func (h *Handler) GetSubscriptionOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID, err := uuid.Parse(chi.URLParam(r, "orderId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if order.UserID != user.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	inv := invoiceFromMetadata(order.Metadata)
	var op *payment.CryptoOperation

	if order.Status == "pending" {
		if h.gateway.IsConfigured() && order.ProviderOrderID != "" {
			live, err := h.gateway.GetOperation(r.Context(), order.ProviderOrderID)
			if err != nil {
				slog.Warn("payments.reconcile_error", "order_id", order.ID, "error", err.Error())
			} else {
				op = live
				switch op.Status {
				case "completed", "mismatch":
					// "mismatch" = funds received but over-paid; still credited.
					h.settleOrder(r, order.ID)
				case "expired":
					h.closeOrder(r, order.ID, "expired")
				case "cancelled", "error":
					h.closeOrder(r, order.ID, "cancelled")
				}
			}
		} else if inv != nil && expired(inv.ExpireUTC) {
			h.closeOrder(r, order.ID, "expired")
		}
		if refreshed, err := h.repo.GetPaymentOrderByID(r.Context(), order.ID); err == nil {
			order = refreshed
		}
	}

	writeJSON(w, http.StatusOK, h.buildOrderStatus(order, h.planDisplayName(r, order.PlanID), inv, op))
}

// HandleCryptoWebhook POST /webhooks/payments/crypto — signed gateway callback.
func (h *Handler) HandleCryptoWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.gateway.IsConfigured() {
		writeError(w, http.StatusServiceUnavailable, "payments are not configured")
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var data map[string]interface{}
	var payload payment.WebhookPayload
	if json.Unmarshal(bodyBytes, &data) == nil {
		_ = json.Unmarshal(bodyBytes, &payload)
	}
	if payload.OrderNumber == "" {
		// Form-encoded callback fallback.
		if vals, err := url.ParseQuery(string(bodyBytes)); err == nil {
			data = make(map[string]interface{}, len(vals))
			for k, v := range vals {
				if len(v) > 0 {
					data[k] = v[0]
				}
			}
			payload.Status = vals.Get("status")
			payload.OrderNumber = vals.Get("order_number")
			payload.TxnID = vals.Get("txn_id")
			payload.SourceAmount = vals.Get("source_amount")
			payload.VerifyHash = vals.Get("verify_hash")
		}
	}

	if payload.VerifyHash == "" {
		slog.Warn("payments.webhook_unsigned")
		writeError(w, http.StatusUnauthorized, "verify_hash signature is required")
		return
	}
	if !h.gateway.VerifyWebhook(data, payload.VerifyHash, bodyBytes) {
		slog.Warn("payments.webhook_bad_signature", "order_number", payload.OrderNumber)
		writeError(w, http.StatusUnprocessableEntity, "signature verification failed")
		return
	}

	orderID, err := uuid.Parse(payload.OrderNumber)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order_number format")
		return
	}
	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	if order.ProviderOrderID != "" && payload.TxnID != "" && payload.TxnID != order.ProviderOrderID {
		slog.Warn("payments.webhook_txn_mismatch", "order_id", order.ID)
		writeError(w, http.StatusBadRequest, "transaction mismatch")
		return
	}
	if payload.SourceAmount != "" {
		if f, err := strconv.ParseFloat(payload.SourceAmount, 64); err == nil {
			if int(math.Round(f*100)) != order.AmountCents {
				slog.Warn("payments.webhook_amount_mismatch", "order_id", order.ID)
				writeError(w, http.StatusBadRequest, "amount mismatch")
				return
			}
		}
	}

	switch strings.ToLower(payload.Status) {
	case "completed", "mismatch":
		h.settleOrder(r, order.ID)
	case "expired":
		h.closeOrder(r, order.ID, "expired")
	case "cancelled", "error":
		h.closeOrder(r, order.ID, "cancelled")
	default:
		slog.Info("payments.webhook_pending", "order_id", order.ID, "status", payload.Status)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// settleOrder marks an order completed and activates the role exactly once.
func (h *Handler) settleOrder(r *http.Request, orderID uuid.UUID) {
	unlock := lockOrder(orderID)
	defer unlock()

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil || order.Status == "completed" {
		return
	}
	if err := h.repo.UpdatePaymentOrderStatus(r.Context(), order.ID, "completed", ""); err != nil {
		slog.Error("payments.settle_failed", "order_id", order.ID, "error", err.Error())
		return
	}
	h.activateOrderMembershipAndRole(r.Context(), order)
	slog.Info("payments.settled", "order_id", order.ID, "plan", order.PlanID)
}

// closeOrder moves a pending order to expired/cancelled (never downgrades paid).
func (h *Handler) closeOrder(r *http.Request, orderID uuid.UUID, status string) {
	unlock := lockOrder(orderID)
	defer unlock()

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil || order.Status != "pending" {
		return
	}
	_ = h.repo.UpdatePaymentOrderStatus(r.Context(), order.ID, status, "")
}

func (h *Handler) planDisplayName(r *http.Request, planID string) string {
	if memPlan, err := h.repo.GetMembershipPlanByID(r.Context(), planID); err == nil && memPlan != nil {
		return memPlan.Name
	}
	if plan, err := h.repo.GetSubscriptionPlanByID(r.Context(), planID); err == nil && plan != nil {
		return plan.Name
	}
	return planID
}

// buildOrderStatus returns the sanitized status payload consumed by the UI.
func (h *Handler) buildOrderStatus(order *domain.PaymentOrder, planName string, inv *payment.CryptoInvoice, op *payment.CryptoOperation) map[string]interface{} {
	out := map[string]interface{}{
		"order_id":      order.ID,
		"status":        order.Status,
		"plan_id":       order.PlanID,
		"plan_name":     planName,
		"billing_cycle": order.BillingCycle,
		"amount_cents":  order.AmountCents,
		"fiat_currency": order.Currency,
		"created_at":    order.CreatedAt,
	}
	if inv != nil {
		out["coin"] = inv.Currency
		out["crypto_amount"] = inv.Amount
		out["pending_amount"] = inv.PendingAmount
		out["wallet_hash"] = inv.WalletHash
		out["qr_code"] = inv.QRCode
		out["expire_at"] = toUnix(inv.ExpireUTC)
		out["gateway_status"] = inv.Status
		out["is_simulated"] = inv.IsSimulated
	}
	if op != nil {
		out["gateway_status"] = op.Status
		if op.Amount != "" {
			out["crypto_amount"] = op.Amount
		}
		// Only override pending_amount if non-zero, preserving the total required amount
		if op.PendingAmount != "" && op.PendingAmount != "0" && op.PendingAmount != "0.000000" && op.PendingAmount != "0.00000000" {
			out["pending_amount"] = op.PendingAmount
		}
		out["received_amount"] = op.ReceivedAmount
		if op.Currency != "" {
			out["coin"] = op.Currency
		}
	}
	if out["pending_amount"] == nil || out["pending_amount"] == "" || out["pending_amount"] == "0.000000" || out["pending_amount"] == "0.00000000" {
		out["pending_amount"] = out["crypto_amount"]
	}
	return out
}

// invoiceFromMetadata decodes the stored invoice regardless of whether the
// repository returned a struct pointer (memory) or a JSON map (postgres).
func invoiceFromMetadata(meta map[string]interface{}) *payment.CryptoInvoice {
	if meta == nil {
		return nil
	}
	raw, ok := meta["payment"]
	if !ok {
		raw, ok = meta["white_label"] // legacy orders
	}
	if !ok || raw == nil {
		return nil
	}
	if inv, ok := raw.(*payment.CryptoInvoice); ok {
		return inv
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var inv payment.CryptoInvoice
	if err := json.Unmarshal(b, &inv); err != nil {
		return nil
	}
	return &inv
}

func toUnix(v interface{}) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	case json.Number:
		n, _ := t.Int64()
		return n
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err == nil {
			return n
		}
		if ts, err := time.Parse(time.RFC3339, t); err == nil {
			return ts.Unix()
		}
	}
	return 0
}

func expired(v interface{}) bool {
	ts := toUnix(v)
	return ts > 0 && time.Now().Unix() > ts
}

