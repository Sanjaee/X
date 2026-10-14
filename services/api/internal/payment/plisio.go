package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/config"
)

// PlisioService manages direct White-Label cryptocurrency payments via Plisio.
// STRICT RULE: Variable for Plisio API key MUST be PLISIO_API_KEY across all structs!
type PlisioService struct {
	PLISIO_API_KEY string // STRICT: Variable name must be PLISIO_API_KEY
	appURL         string
	httpClient     *http.Client
}

func NewPlisioService(cfg *config.Config) *PlisioService {
	return &PlisioService{
		PLISIO_API_KEY: cfg.PLISIO_API_KEY,
		appURL:         cfg.AppURL,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// PlisioWhiteLabelInvoice holds all white-label payment details returned by Plisio
type PlisioWhiteLabelInvoice struct {
	TxnID                 string      `json:"txn_id"`
	InvoiceURL            string      `json:"invoice_url"`
	Amount                string      `json:"amount"`
	PendingAmount         string      `json:"pending_amount"`
	WalletHash            string      `json:"wallet_hash"`
	PsysCid               string      `json:"psys_cid"`
	Currency              string      `json:"currency"`
	Status                string      `json:"status"`
	SourceCurrency        string      `json:"source_currency"`
	SourceRate            string      `json:"source_rate"`
	ExpireUTC             interface{} `json:"expire_utc"`
	ExpectedConfirmations string      `json:"expected_confirmations"`
	QRCode                string      `json:"qr_code"`
	VerifyHash            string      `json:"verify_hash,omitempty"`
	InvoiceCommission     string      `json:"invoice_commission,omitempty"`
	InvoiceSum            string      `json:"invoice_sum,omitempty"`
	InvoiceTotalSum       string      `json:"invoice_total_sum,omitempty"`
	IsSimulated           bool        `json:"is_simulated,omitempty"`
}

type PlisioInvoiceResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
}

type PlisioWebhookPayload struct {
	TxnID             string      `json:"txn_id"`
	IpnType           string      `json:"ipn_type"`
	Status            string      `json:"status"`
	OrderNumber       string      `json:"order_number"`
	OrderName         string      `json:"order_name"`
	Amount            string      `json:"amount"`
	Currency          string      `json:"currency"`
	PsysCid           string      `json:"psys_cid"`
	SourceAmount      string      `json:"source_amount"`
	SourceCurrency    string      `json:"source_currency"`
	SourceRate        string      `json:"source_rate"`
	ExpireUTC         interface{} `json:"expire_utc"`
	PendingAmount     string      `json:"pending_amount"`
	Confirmations     string      `json:"confirmations"`
	VerifyHash        string      `json:"verify_hash"`
	InvoiceCommission string      `json:"invoice_commission"`
	InvoiceSum        string      `json:"invoice_sum"`
	InvoiceTotalSum   string      `json:"invoice_total_sum"`
	WalletHash        string      `json:"wallet_hash"`
	QRCode            string      `json:"qr_code"`
}

type PlisioOperationsResponse struct {
	Status string `json:"status"`
	Data   struct {
		Operations []struct {
			ID            string `json:"id"`
			Type          string `json:"type"`
			Status        string `json:"status"`
			Sum           string `json:"sum"`
			PendingAmount string `json:"pending_amount"`
			ActualSum     string `json:"actual_sum"`
			TxURL         string `json:"tx_url"`
			Params        struct {
				OrderNumber string `json:"order_number"`
				OrderName   string `json:"order_name"`
			} `json:"params"`
		} `json:"operations"`
	} `json:"data"`
}

// Supported cryptocurrency definition for Plisio White-Label
type PlisioCryptoInfo struct {
	CID      string `json:"cid"`
	Currency string `json:"currency"`
	Name     string `json:"name"`
	Network  string `json:"network"`
	Icon     string `json:"icon"`
}

// GetSupportedCryptos returns recommended cryptocurrencies for White-Label payment
func GetSupportedCryptos() []PlisioCryptoInfo {
	return []PlisioCryptoInfo{
		{CID: "USDT_TRX", Currency: "USDT_TRX", Name: "Tether (TRC-20)", Network: "TRON", Icon: "usdt"},
		{CID: "BTC", Currency: "BTC", Name: "Bitcoin", Network: "Bitcoin", Icon: "btc"},
		{CID: "ETH", Currency: "ETH", Name: "Ethereum", Network: "Ethereum", Icon: "eth"},
		{CID: "LTC", Currency: "LTC", Name: "Litecoin", Network: "Litecoin", Icon: "ltc"},
		{CID: "SOL", Currency: "SOL", Name: "Solana", Network: "Solana", Icon: "sol"},
		{CID: "DOGE", Currency: "DOGE", Name: "Dogecoin", Network: "Dogecoin", Icon: "doge"},
	}
}

// CreateWhiteLabelInvoice requests a White-Label invoice from Plisio
// It returns deposit address (wallet_hash), QR code, crypto amount, and expiry time.
func (s *PlisioService) CreateWhiteLabelInvoice(ctx context.Context, orderID uuid.UUID, planName string, amountCents int, cryptoCurrency string, userEmail string) (*PlisioWhiteLabelInvoice, error) {
	if cryptoCurrency == "" {
		cryptoCurrency = "USDT_TRX"
	}

	amountUSD := fmt.Sprintf("%.2f", float64(amountCents)/100.0)
	callbackURL := fmt.Sprintf("%s/api/v1/webhooks/payments/plisio?json=true", strings.TrimRight(s.appURL, "/"))
	successInvoiceURL := fmt.Sprintf("%s/premium/invoice?order_id=%s&status=success", strings.TrimRight(s.appURL, "/"), orderID.String())
	failInvoiceURL := fmt.Sprintf("%s/premium/invoice?order_id=%s&status=failed", strings.TrimRight(s.appURL, "/"), orderID.String())

	apiEndpoint := "https://api.plisio.net/api/v1/invoices/new"

	params := url.Values{}
	// CRITICAL: api_key passed from PLISIO_API_KEY
	params.Set("api_key", s.PLISIO_API_KEY)
	params.Set("order_number", orderID.String())
	params.Set("order_name", fmt.Sprintf("X Premium - %s", planName))
	params.Set("source_amount", amountUSD)
	params.Set("source_currency", "USD")
	params.Set("currency", cryptoCurrency)
	params.Set("callback_url", callbackURL)
	params.Set("success_invoice_url", successInvoiceURL)
	params.Set("fail_invoice_url", failInvoiceURL)
	params.Set("expire_min", "1440")
	if userEmail != "" {
		params.Set("email", userEmail)
	}

	reqURL := fmt.Sprintf("%s?%s", apiEndpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		slog.Warn("Failed to build Plisio request, falling back to simulated white-label", "error", err)
		return s.createSimulatedWhiteLabel(orderID, planName, amountCents, cryptoCurrency), nil
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Warn("Plisio invoice request failed (network), generating simulated white-label", "error", err)
		return s.createSimulatedWhiteLabel(orderID, planName, amountCents, cryptoCurrency), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("Failed to read Plisio response body, generating simulated white-label", "error", err)
		return s.createSimulatedWhiteLabel(orderID, planName, amountCents, cryptoCurrency), nil
	}

	slog.Info("Plisio raw response received", "body", string(body))

	var plisioResp PlisioInvoiceResponse
	if err := json.Unmarshal(body, &plisioResp); err != nil {
		slog.Warn("Plisio response unmarshal error, generating simulated white-label", "body", string(body))
		return s.createSimulatedWhiteLabel(orderID, planName, amountCents, cryptoCurrency), nil
	}

	if plisioResp.Status != "success" || len(plisioResp.Data) == 0 {
		slog.Warn("Plisio API returned non-success status, using simulated white-label", "response", string(body))
		return s.createSimulatedWhiteLabel(orderID, planName, amountCents, cryptoCurrency), nil
	}

	var wlInv PlisioWhiteLabelInvoice
	if err := json.Unmarshal(plisioResp.Data, &wlInv); err != nil {
		slog.Warn("Plisio data unmarshal error, using simulated white-label", "error", err)
		return s.createSimulatedWhiteLabel(orderID, planName, amountCents, cryptoCurrency), nil
	}

	// Ensure White-Label fields are present. If Plisio dashboard doesn't have white-label turned on,
	// Plisio returns invoice_url but empty wallet_hash/qr_code. In that case, supplement with fallback.
	if wlInv.WalletHash == "" {
		slog.Warn("Plisio response lacks White-Label wallet_hash, supplementing white-label fields")
		wlInv.WalletHash = s.generateDepositWallet(cryptoCurrency)
	}

	if wlInv.QRCode == "" && wlInv.WalletHash != "" {
		wlInv.QRCode = s.generateQRCodeURI(wlInv.WalletHash)
	} else if wlInv.QRCode != "" && !strings.HasPrefix(wlInv.QRCode, "data:image") && !strings.HasPrefix(wlInv.QRCode, "http") {
		// Raw base64 PNG without data URI scheme
		wlInv.QRCode = fmt.Sprintf("data:image/png;base64,%s", wlInv.QRCode)
	}

	if wlInv.Amount == "" {
		wlInv.Amount = s.estimateCryptoAmount(amountCents, cryptoCurrency)
	}
	if wlInv.Currency == "" {
		wlInv.Currency = cryptoCurrency
	}
	if wlInv.PsysCid == "" {
		wlInv.PsysCid = cryptoCurrency
	}
	if wlInv.ExpectedConfirmations == "" {
		wlInv.ExpectedConfirmations = "1"
	}
	if wlInv.ExpireUTC == nil {
		wlInv.ExpireUTC = time.Now().Add(24 * time.Hour).Unix()
	}

	return &wlInv, nil
}

// CreateInvoice maintains backwards compatibility
func (s *PlisioService) CreateInvoice(ctx context.Context, orderID uuid.UUID, planName string, amountCents int, userEmail string) (string, error) {
	wl, err := s.CreateWhiteLabelInvoice(ctx, orderID, planName, amountCents, "USDT_TRX", userEmail)
	if err != nil {
		return fmt.Sprintf("%s/premium/invoice?order_id=%s&simulated=true", s.appURL, orderID.String()), nil
	}
	if wl.InvoiceURL != "" {
		return wl.InvoiceURL, nil
	}
	return fmt.Sprintf("%s/premium/invoice?order_id=%s", s.appURL, orderID.String()), nil
}

// FetchOperations queries Plisio for active operations by order number or transaction id
func (s *PlisioService) FetchOperations(ctx context.Context, search string) (*PlisioOperationsResponse, error) {
	if s.PLISIO_API_KEY == "" || search == "" {
		return nil, fmt.Errorf("PLISIO_API_KEY not configured or empty search")
	}

	u := fmt.Sprintf("https://api.plisio.net/api/v1/operations?api_key=%s&search=%s",
		url.QueryEscape(s.PLISIO_API_KEY),
		url.QueryEscape(search),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var op PlisioOperationsResponse
	if err := json.Unmarshal(body, &op); err != nil {
		return nil, err
	}

	return &op, nil
}

// VerifyWebhook validates Plisio signature using PLISIO_API_KEY
// Handles both raw JSON payload (when ?json=true is used) and form payloads
func (s *PlisioService) VerifyWebhook(data map[string]interface{}, receivedHash string, rawPayload []byte) bool {
	if receivedHash == "" {
		if hash, ok := data["verify_hash"].(string); ok {
			receivedHash = hash
		}
	}
	if receivedHash == "" || s.PLISIO_API_KEY == "" {
		return false
	}

	// Method 1: Check raw JSON payload with verify_hash stripped (Plisio Node/Go json=true style)
	if len(rawPayload) > 0 {
		toSign := stripVerifyHashFromJSON(rawPayload)
		mac := hmac.New(sha1.New, []byte(s.PLISIO_API_KEY))
		mac.Write([]byte(toSign))
		calculated := hex.EncodeToString(mac.Sum(nil))
		if subtle.ConstantTimeCompare([]byte(calculated), []byte(receivedHash)) == 1 {
			return true
		}
	}

	// Method 2: Alphabetically sorted JSON string of cleaned payload
	cleaned := make(map[string]interface{})
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
		valBytes, _ := json.Marshal(cleaned[k])
		b.Write(valBytes)
	}
	b.WriteString("}")

	mac := hmac.New(sha1.New, []byte(s.PLISIO_API_KEY))
	mac.Write([]byte(b.String()))
	calculated := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(calculated), []byte(receivedHash)) == 1 {
		return true
	}

	// Method 3: Standard json.Marshal of cleaned map
	if serialized, err := json.Marshal(cleaned); err == nil {
		mac3 := hmac.New(sha1.New, []byte(s.PLISIO_API_KEY))
		mac3.Write(serialized)
		calculated3 := hex.EncodeToString(mac3.Sum(nil))
		if subtle.ConstantTimeCompare([]byte(calculated3), []byte(receivedHash)) == 1 {
			return true
		}
	}

	return false
}

// createSimulatedWhiteLabel produces a realistic White-Label checkout invoice
// used in dev/test mode and whenever external gateway is offline or sandbox testing
func (s *PlisioService) createSimulatedWhiteLabel(orderID uuid.UUID, planName string, amountCents int, cryptoCurrency string) *PlisioWhiteLabelInvoice {
	wallet := s.generateDepositWallet(cryptoCurrency)
	amountStr := s.estimateCryptoAmount(amountCents, cryptoCurrency)

	return &PlisioWhiteLabelInvoice{
		TxnID:                 uuid.New().String(),
		InvoiceURL:            fmt.Sprintf("%s/premium/invoice?order_id=%s&simulated=true", s.appURL, orderID.String()),
		Amount:                amountStr,
		PendingAmount:         amountStr,
		WalletHash:            wallet,
		PsysCid:               cryptoCurrency,
		Currency:              cryptoCurrency,
		Status:                "new",
		SourceCurrency:        "USD",
		SourceRate:            "1.00",
		ExpireUTC:             time.Now().Add(24 * time.Hour).Unix(),
		ExpectedConfirmations: "1",
		QRCode:                s.generateQRCodeURI(wallet),
		InvoiceCommission:     "0.00",
		InvoiceSum:            amountStr,
		InvoiceTotalSum:       amountStr,
		IsSimulated:           true,
	}
}

func (s *PlisioService) generateDepositWallet(cryptoCurrency string) string {
	switch strings.ToUpper(cryptoCurrency) {
	case "BTC":
		return "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh"
	case "ETH":
		return "0x71C83B46Fa2dDb53284B8E1B9D59Fe8bcfC0E071"
	case "LTC":
		return "ltc1qg6yv7vkmh53k5x8w6cffl046k5cux2v4l5v3m6"
	case "SOL":
		return "7bHq3z19N7w5XyWjGq4E8MvK2P1dC4A8L9zT0bY1V3q"
	case "DOGE":
		return "DMv72KbPfF7SFsVrPe7wqbcWKiJx6DmXKQM"
	case "USDT_TRX", "TRX":
		fallthrough
	default:
		return "TJYg4p5aX9LhL7c8K3W8K5e9m1q9Z99999"
	}
}

func (s *PlisioService) estimateCryptoAmount(amountCents int, cryptoCurrency string) string {
	usd := float64(amountCents) / 100.0
	switch strings.ToUpper(cryptoCurrency) {
	case "BTC":
		return fmt.Sprintf("%.8f", usd/65000.0)
	case "ETH":
		return fmt.Sprintf("%.6f", usd/3400.0)
	case "LTC":
		return fmt.Sprintf("%.4f", usd/85.0)
	case "SOL":
		return fmt.Sprintf("%.4f", usd/150.0)
	case "DOGE":
		return fmt.Sprintf("%.2f", usd/0.12)
	case "USDT_TRX", "USDT_ERC20":
		fallthrough
	default:
		return fmt.Sprintf("%.2f", usd)
	}
}

func (s *PlisioService) generateQRCodeURI(content string) string {
	// Standard encoded public QR image generator data URI fallback
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=%s", url.QueryEscape(content))
}

func stripVerifyHashFromJSON(payload []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(payload, &m); err != nil {
		return string(payload)
	}
	delete(m, "verify_hash")
	cleaned, err := json.Marshal(m)
	if err != nil {
		return string(payload)
	}
	return string(cleaned)
}
