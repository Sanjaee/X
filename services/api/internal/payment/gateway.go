package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/config"
)

// CryptoGateway is the server-side client for the upstream crypto payment
// gateway. The API key and base URL are read from the environment only
// (PAYMENT_GATEWAY_API_KEY / PAYMENT_GATEWAY_BASE_URL) and are never
// returned to clients, written to logs, or embedded in error messages.
type CryptoGateway struct {
	apiKey     string
	baseURL    string
	appURL     string
	httpClient *http.Client

	cacheMu        sync.Mutex
	cachedCoins    []CryptoCurrency
	cachedCoinsAt  time.Time
	currencyTTL    time.Duration
}

// NewCryptoGateway builds the gateway client from config.
func NewCryptoGateway(cfg *config.Config) *CryptoGateway {
	return &CryptoGateway{
		apiKey:      cfg.PaymentGatewayAPIKey,
		baseURL:     strings.TrimRight(cfg.PaymentGatewayBaseURL, "/"),
		appURL:      strings.TrimRight(cfg.AppURL, "/"),
		httpClient:  &http.Client{Timeout: 20 * time.Second},
		currencyTTL: 5 * time.Minute,
	}
}

// IsConfigured reports whether a real key + base URL are present.
func (g *CryptoGateway) IsConfigured() bool {
	return g != nil && g.apiKey != "" && g.apiKey != "CHANGE_ME" && g.baseURL != ""
}

// CryptoCurrency is the sanitized coin shape exposed to the frontend.
type CryptoCurrency struct {
	CID      string `json:"cid"`
	Currency string `json:"currency"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	PriceUsd string `json:"price_usd,omitempty"`
}

// CryptoInvoice is the white-label payment data needed to render the
// payment screen (deposit address, QR, crypto amount, expiry).
type CryptoInvoice struct {
	TxnID                 string      `json:"txn_id"`
	Amount                string      `json:"amount"`
	PendingAmount         string      `json:"pending_amount"`
	WalletHash            string      `json:"wallet_hash"`
	Currency              string      `json:"currency"`
	Status                string      `json:"status"`
	ExpireUTC             interface{} `json:"expire_utc"`
	ExpectedConfirmations string      `json:"expected_confirmations"`
	QRCode                string      `json:"qr_code"`
	IsSimulated           bool        `json:"is_simulated,omitempty"`
}

// CryptoOperation is the live status of an invoice on the gateway.
type CryptoOperation struct {
	Status         string `json:"status"` // new / pending / pending internal / completed / mismatch / expired / cancelled / error
	Amount         string `json:"amount"`
	PendingAmount  string `json:"pending_amount"`
	ReceivedAmount string `json:"received_amount"`
	Currency       string `json:"currency"`
}

// WebhookPayload is the callback body sent by the gateway.
type WebhookPayload struct {
	TxnID         string `json:"txn_id"`
	Status        string `json:"status"`
	OrderNumber   string `json:"order_number"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	SourceAmount  string `json:"source_amount"`
	PendingAmount string `json:"pending_amount"`
	VerifyHash    string `json:"verify_hash"`
}

// ErrGatewayUnavailable is returned when the gateway is not configured.
var ErrGatewayUnavailable = errors.New("payment gateway is not configured")

// ---------------------------------------------------------------------------
// Currencies
// ---------------------------------------------------------------------------

// GetDefaultCurrencies returns standard payable cryptocurrencies.
func GetDefaultCurrencies() []CryptoCurrency {
	return []CryptoCurrency{
		{CID: "USDT_TRX", Currency: "USDT_TRX", Name: "Tether (TRC-20)", Icon: "/crypto-icons/USDT_TRX.svg", PriceUsd: "1.00"},
		{CID: "BTC", Currency: "BTC", Name: "Bitcoin", Icon: "/crypto-icons/BTC.svg"},
		{CID: "ETH", Currency: "ETH", Name: "Ethereum", Icon: "/crypto-icons/ETH.svg"},
		{CID: "LTC", Currency: "LTC", Name: "Litecoin", Icon: "/crypto-icons/LTC.svg"},
		{CID: "SOL", Currency: "SOL", Name: "Solana", Icon: "/crypto-icons/SOL.svg"},
		{CID: "DOGE", Currency: "DOGE", Name: "Dogecoin", Icon: "/crypto-icons/DOGE.svg"},
		{CID: "TRX", Currency: "TRX", Name: "TRON", Icon: "/crypto-icons/TRX.svg"},
		{CID: "BNB", Currency: "BNB", Name: "BNB Chain", Icon: "/crypto-icons/BNB.svg"},
		{CID: "TON", Currency: "TON", Name: "Toncoin", Icon: "/crypto-icons/TON.svg"},
		{CID: "XMR", Currency: "XMR", Name: "Monero", Icon: "/crypto-icons/XMR.svg"},
	}
}

// FetchCurrencies returns the live list of payable coins (cached briefly).
func (g *CryptoGateway) FetchCurrencies(ctx context.Context) ([]CryptoCurrency, error) {
	if !g.IsConfigured() {
		return GetDefaultCurrencies(), nil
	}

	g.cacheMu.Lock()
	if len(g.cachedCoins) > 0 && time.Since(g.cachedCoinsAt) < g.currencyTTL {
		out := g.cachedCoins
		g.cacheMu.Unlock()
		return out, nil
	}
	g.cacheMu.Unlock()

	body, err := g.request(ctx, "/currencies", url.Values{})
	if err != nil {
		return nil, err
	}

	var wrap struct {
		Status string `json:"status"`
		Data   []struct {
			Name        string      `json:"name"`
			Cid         string      `json:"cid"`
			Currency    string      `json:"currency"`
			PriceUsd    interface{} `json:"price_usd"`
			Hidden      interface{} `json:"hidden"`
			Maintenance bool        `json:"maintenance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil || wrap.Status != "success" {
		return nil, errors.New("gateway currencies: unexpected response")
	}

	out := make([]CryptoCurrency, 0, len(wrap.Data))
	for _, c := range wrap.Data {
		if c.Maintenance || isTruthy(c.Hidden) {
			continue
		}
		cid := c.Cid
		if cid == "" {
			cid = c.Currency
		}
		price := ""
		if c.PriceUsd != nil {
			price = fmt.Sprintf("%v", c.PriceUsd)
		}
		out = append(out, CryptoCurrency{
			CID:      cid,
			Currency: c.Currency,
			Name:     c.Name,
			// Icons are served statically from /public — never from the gateway host.
			Icon:     "/crypto-icons/" + cid + ".svg",
			PriceUsd: price,
		})
	}

	g.cacheMu.Lock()
	g.cachedCoins = out
	g.cachedCoinsAt = time.Now()
	g.cacheMu.Unlock()
	return out, nil
}

// IsSupportedCurrency checks a coin code against the live list.
func (g *CryptoGateway) IsSupportedCurrency(ctx context.Context, code string) bool {
	coins, err := g.FetchCurrencies(ctx)
	if err != nil {
		return false
	}
	for _, c := range coins {
		if strings.EqualFold(c.CID, code) || strings.EqualFold(c.Currency, code) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

// CreateInvoice creates a white-label crypto invoice for an order.
func (g *CryptoGateway) CreateInvoice(ctx context.Context, orderID uuid.UUID, planName string, amountCents int, coin string, email string) (*CryptoInvoice, error) {
	if !g.IsConfigured() {
		return nil, ErrGatewayUnavailable
	}

	q := url.Values{}
	q.Set("order_number", orderID.String())
	q.Set("order_name", fmt.Sprintf("X Premium - %s", planName))
	q.Set("source_currency", "USD")
	q.Set("source_amount", fmt.Sprintf("%.2f", float64(amountCents)/100.0))
	q.Set("currency", coin)
	q.Set("callback_url", g.appURL+"/api/v1/webhooks/payments/crypto?json=true")
	q.Set("expire_min", "60")
	q.Set("json", "true")
	if email != "" {
		q.Set("email", email)
	}

	body, err := g.request(ctx, "/invoices/new", q)
	if err != nil {
		return nil, err
	}

	var wrap struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil || wrap.Status != "success" {
		return nil, fmt.Errorf("gateway invoice rejected: %s", gatewayErrorMessage(body))
	}

	var inv CryptoInvoice
	if err := json.Unmarshal(wrap.Data, &inv); err != nil {
		return nil, errors.New("gateway invoice: malformed data")
	}
	if inv.WalletHash == "" {
		// Never fabricate a deposit address — fail loudly instead.
		return nil, errors.New("gateway invoice: deposit address missing (white-label disabled?)")
	}
	if inv.QRCode != "" && !strings.HasPrefix(inv.QRCode, "data:image") && !strings.HasPrefix(inv.QRCode, "http") {
		inv.QRCode = "data:image/png;base64," + inv.QRCode
	}
	if inv.Currency == "" {
		inv.Currency = coin
	}
	if inv.PendingAmount == "" {
		inv.PendingAmount = inv.Amount
	}
	if inv.ExpireUTC == nil {
		inv.ExpireUTC = time.Now().Add(60 * time.Minute).Unix()
	}
	return &inv, nil
}

// CreateSimulatedInvoice is only used in local dev when no key is configured
// AND ALLOW_PAYMENT_SIMULATION=true. It is never used in production.
func (g *CryptoGateway) CreateSimulatedInvoice(amountCents int, coin string) *CryptoInvoice {
	amount := fmt.Sprintf("%.2f", float64(amountCents)/100.0)
	return &CryptoInvoice{
		TxnID:                 "SIM-" + uuid.NewString(),
		Amount:                amount,
		PendingAmount:         amount,
		WalletHash:            "SIMULATED-ADDRESS-DO-NOT-PAY",
		Currency:              coin,
		Status:                "new",
		ExpireUTC:             time.Now().Add(60 * time.Minute).Unix(),
		ExpectedConfirmations: "1",
		IsSimulated:           true,
	}
}

// GetOperation fetches the live status of an invoice by its transaction id.
func (g *CryptoGateway) GetOperation(ctx context.Context, txnID string) (*CryptoOperation, error) {
	if !g.IsConfigured() {
		return nil, ErrGatewayUnavailable
	}
	if txnID == "" || strings.HasPrefix(txnID, "SIM-") {
		return nil, errors.New("no remote operation")
	}

	body, err := g.request(ctx, "/operations/"+url.PathEscape(txnID), url.Values{})
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil || wrap.Status != "success" {
		return nil, errors.New("gateway operation: unexpected response")
	}
	d := wrap.Data
	op := &CryptoOperation{
		Status:         strings.ToLower(firstNonEmpty(d, "status")),
		Amount:         firstNonEmpty(d, "amount", "sum"),
		PendingAmount:  firstNonEmpty(d, "pending_amount", "pending_sum"),
		ReceivedAmount: firstNonEmpty(d, "received_amount", "actual_sum"),
		Currency:       firstNonEmpty(d, "currency", "psys_cid"),
	}
	return op, nil
}

// ---------------------------------------------------------------------------
// Webhook verification (HMAC-SHA1 with the API key)
// ---------------------------------------------------------------------------

var verifyHashJSONRe = regexp.MustCompile(`"verify_hash"\s*:\s*"[^"]*",?`)

// VerifyWebhook validates the callback signature. Supports raw JSON
// (json=true) and form payloads.
func (g *CryptoGateway) VerifyWebhook(data map[string]interface{}, receivedHash string, rawPayload []byte) bool {
	if receivedHash == "" {
		if h, ok := data["verify_hash"].(string); ok {
			receivedHash = h
		}
	}
	if receivedHash == "" || g.apiKey == "" {
		return false
	}

	candidates := make([][]byte, 0, 4)
	if len(rawPayload) > 0 {
		candidates = append(candidates, verifyHashJSONRe.ReplaceAll(rawPayload, nil))
		var m map[string]interface{}
		if json.Unmarshal(rawPayload, &m) == nil {
			delete(m, "verify_hash")
			if b, err := json.Marshal(m); err == nil {
				candidates = append(candidates, b)
			}
		}
	}

	cleaned := make(map[string]interface{}, len(data))
	keys := make([]string, 0, len(data))
	for k, v := range data {
		if k != "verify_hash" {
			cleaned[k] = v
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf("%q:", k))
		vb, _ := json.Marshal(cleaned[k])
		b.Write(vb)
	}
	b.WriteString("}")
	candidates = append(candidates, []byte(b.String()))

	for _, c := range candidates {
		mac := hmac.New(sha1.New, []byte(g.apiKey))
		mac.Write(c)
		calc := hex.EncodeToString(mac.Sum(nil))
		if subtle.ConstantTimeCompare([]byte(calc), []byte(receivedHash)) == 1 {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

// request performs a GET against the gateway. The API key is appended here
// and the resulting URL is never logged or surfaced in errors.
func (g *CryptoGateway) request(ctx context.Context, path string, q url.Values) ([]byte, error) {
	q.Set("api_key", g.apiKey)
	reqURL := g.baseURL + path + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, errors.New("gateway: failed to build request")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		// *url.Error embeds the full URL (incl. api_key) — strip it.
		var ue *url.Error
		if errors.As(err, &ue) {
			return nil, fmt.Errorf("gateway request failed: %v", ue.Err)
		}
		return nil, errors.New("gateway request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, errors.New("gateway: failed to read response")
	}
	if resp.StatusCode >= 400 {
		slog.Warn("payment gateway http error", "path", path, "status", resp.StatusCode)
		return body, fmt.Errorf("gateway http %d: %s", resp.StatusCode, gatewayErrorMessage(body))
	}
	return body, nil
}

func gatewayErrorMessage(body []byte) string {
	var e struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &e) == nil && e.Data.Message != "" {
		return e.Data.Message
	}
	return "unknown error"
}

func firstNonEmpty(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			s := fmt.Sprintf("%v", v)
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func isTruthy(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t != "" && t != "0" && t != "false"
	}
	return false
}

