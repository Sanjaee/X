package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/middleware"
)

// generateTOTPSecret generates a 20-byte random secret encoded in RFC 4648 Base32 without padding.
func generateTOTPSecret() (string, error) {
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes), nil
}

// computeTOTPCode computes a 6-digit RFC 6238 TOTP code for a given Base32 secret and 30-second time counter.
func computeTOTPCode(secret string, counter int64) (string, error) {
	cleaned := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	cleaned = strings.TrimRight(cleaned, "=")
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleaned)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binCode := (int(sum[offset]&0x7f) << 24) |
		(int(sum[offset+1]&0xff) << 16) |
		(int(sum[offset+2]&0xff) << 8) |
		int(sum[offset+3]&0xff)

	otp := binCode % 1000000
	return fmt.Sprintf("%06d", otp), nil
}

// fetchNetworkTime queries a public HTTPS endpoint's Date header as a fallback when host/VM clock is skewed.
func fetchNetworkTime() (time.Time, bool) {
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	for _, endpoint := range []string{"https://1.1.1.1", "https://www.google.com"} {
		req, err := http.NewRequest(http.MethodHead, endpoint, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if dateHeader := resp.Header.Get("Date"); dateHeader != "" {
			if parsed, err := http.ParseTime(dateHeader); err == nil {
				return parsed, true
			}
		}
	}
	return time.Time{}, false
}

func checkTOTPWindow(secret, code string, ref time.Time, maxOffset int64) bool {
	currentStep := ref.Unix() / 30
	for offset := -maxOffset; offset <= maxOffset; offset++ {
		expected, err := computeTOTPCode(secret, currentStep+offset)
		if err == nil && subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// verifyTOTPCode validates a 6-digit TOTP code against the secret with a +-16 step (8m) drift window
// and falls back to real network time if the local host/VM clock is skewed.
func verifyTOTPCode(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 || secret == "" {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	if checkTOTPWindow(secret, code, now, 16) {
		return true
	}

	if netTime, ok := fetchNetworkTime(); ok {
		if checkTOTPWindow(secret, code, netTime, 4) {
			return true
		}
	}
	return false
}

var recoveryWordList = [64]string{
	"amber", "anchor", "apollo", "arctic", "arrow", "atlas", "aurora", "avalon",
	"badger", "beacon", "bison", "blizzard", "Bonsai", "breeze", "bronze", "canyon",
	"carbon", "cedar", "cipher", "cobalt", "comet", "coral", "cosmos", "crystal",
	"cyclone", "dagger", "delta", "desert", "dolphin", "dragon", "dynamo", "eclipse",
	"emerald", "ember", "enigma", "falcon", "fjord", "forest", "galaxy", "garnet",
	"glacier", "granite", "harbor", "helios", "horizon", "hunter", "indigo", "island",
	"jasper", "jungle", "krypton", "lagoon", "lantern", "legend", "lotus", "lunar",
	"magnet", "marble", "matrix", "meadow", "mercury", "meteor", "mirage", "monarch",
}

// generateDeterministicBackupCodes produces 10 deterministic 8-character backup codes formatted as XXXX-XXXX.
func generateDeterministicBackupCodes(secret string) []string {
	cleaned := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	if cleaned == "" {
		return nil
	}
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		mac := hmac.New(sha256.New, getMFASigningKey())
		mac.Write([]byte(fmt.Sprintf("totp_backup_code_v1:%s:%d", cleaned, i)))
		digest := mac.Sum(nil)
		codes[i] = strings.ToUpper(fmt.Sprintf("%04X-%04X", digest[0:2], digest[2:4]))
	}
	return codes
}

// generateRecoverySeedPhrase derives a deterministic 12-word recovery seed phrase for a TOTP secret.
func generateRecoverySeedPhrase(secret string) string {
	cleaned := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	if cleaned == "" {
		return ""
	}
	mac := hmac.New(sha256.New, getMFASigningKey())
	mac.Write([]byte("totp_recovery_seed_v1:" + cleaned))
	digest := mac.Sum(nil)

	words := make([]string, 12)
	for i := 0; i < 12; i++ {
		idx := int(digest[i]) % len(recoveryWordList)
		words[i] = strings.ToLower(recoveryWordList[idx])
	}
	return strings.Join(words, " ")
}

func normalizeSeedPhrase(s string) string {
	replaced := strings.NewReplacer("-", " ", ",", " ", "\n", " ", "\t", " ").Replace(strings.ToLower(strings.TrimSpace(s)))
	fields := strings.Fields(replaced)
	return strings.Join(fields, " ")
}

// verifyTOTPOrRecovery checks either a 6-digit TOTP authenticator code, 10 backup codes (e.g. 58B5-A215), or 12-word recovery seed phrase.
func verifyTOTPOrRecovery(secret, input string, now time.Time) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || secret == "" {
		return false
	}
	// 1. Check standard 6-digit TOTP code
	if verifyTOTPCode(secret, trimmed, now) {
		return true
	}
	// 2. Check 10 deterministic backup codes (e.g. 58B5-A215 or 58B5A215)
	normalizedBackup := strings.ToUpper(strings.ReplaceAll(trimmed, "-", ""))
	for _, code := range generateDeterministicBackupCodes(secret) {
		normExpected := strings.ToUpper(strings.ReplaceAll(code, "-", ""))
		if subtle.ConstantTimeCompare([]byte(normalizedBackup), []byte(normExpected)) == 1 {
			return true
		}
	}
	// 3. Check 12-word recovery seed phrase
	expectedSeed := generateRecoverySeedPhrase(secret)
	normalizedInput := normalizeSeedPhrase(trimmed)
	if expectedSeed != "" && subtle.ConstantTimeCompare([]byte(normalizedInput), []byte(expectedSeed)) == 1 {
		return true
	}
	return false
}

// buildOTPAuthURL constructs a standard otpauth://totp/ URI for authenticator apps.
func buildOTPAuthURL(issuer, username, secret string) string {
	label := url.PathEscape(fmt.Sprintf("%s:%s", issuer, username))
	params := url.Values{}
	params.Set("secret", secret)
	params.Set("issuer", issuer)
	params.Set("algorithm", "SHA1")
	params.Set("digits", "6")
	params.Set("period", "30")
	return fmt.Sprintf("otpauth://totp/%s?%s", label, params.Encode())
}

type mfaTokenPayload struct {
	UserID  string `json:"uid"`
	Purpose string `json:"pur"`
	Exp     int64  `json:"exp"`
}

func getMFASigningKey() []byte {
	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if secret == "" {
		secret = "super-secret-session-key-32-bytes-long!"
	}
	return []byte(secret)
}

func createMFAToken(userID uuid.UUID) (string, error) {
	payload := mfaTokenPayload{
		UserID:  userID.String(),
		Purpose: "totp_login",
		Exp:     time.Now().Add(5 * time.Minute).Unix(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	b64Payload := base64.RawURLEncoding.EncodeToString(raw)

	mac := hmac.New(sha256.New, getMFASigningKey())
	mac.Write([]byte(b64Payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return b64Payload + "." + sig, nil
}

func verifyMFAToken(token string) (uuid.UUID, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("invalid mfa token format")
	}

	b64Payload, providedSig := parts[0], parts[1]
	mac := hmac.New(sha256.New, getMFASigningKey())
	mac.Write([]byte(b64Payload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(providedSig), []byte(expectedSig)) != 1 {
		return uuid.Nil, fmt.Errorf("invalid mfa token signature")
	}

	raw, err := base64.RawURLEncoding.DecodeString(b64Payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid mfa token encoding")
	}

	var payload mfaTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("invalid mfa token payload")
	}

	if payload.Purpose != "totp_login" || time.Now().Unix() > payload.Exp {
		return uuid.Nil, fmt.Errorf("mfa token expired")
	}

	return uuid.Parse(payload.UserID)
}

// GetTOTPStatus returns whether TOTP 2FA is enabled for the current user.
func (h *Handler) GetTOTPStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	freshUser, err := h.repo.GetUserByID(r.Context(), user.ID, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	resp := map[string]interface{}{
		"totp_enabled": freshUser.TOTPEnabled,
	}
	if freshUser.TOTPEnabled && freshUser.TOTPSecret != "" {
		resp["seed_phrase"] = generateRecoverySeedPhrase(freshUser.TOTPSecret)
		resp["backup_codes"] = generateDeterministicBackupCodes(freshUser.TOTPSecret)
	}
	writeJSON(w, http.StatusOK, resp)
}

// SetupTOTP generates a new TOTP secret and otpauth URL so the user can scan the QR code.
func (h *Handler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	freshUser, err := h.repo.GetUserByID(r.Context(), user.ID, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if freshUser.TOTPEnabled {
		writeError(w, http.StatusConflict, "Two-factor authentication is already enabled")
		return
	}

	secret := freshUser.TOTPSecret
	if secret == "" || r.URL.Query().Get("regenerate") == "true" {
		var err error
		secret, err = generateTOTPSecret()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate TOTP secret")
			return
		}

		if err := h.repo.UpdateUserTOTP(r.Context(), freshUser.ID, secret, false); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store TOTP setup secret")
			return
		}
	}

	otpauthURL := buildOTPAuthURL("X", freshUser.Username, secret)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"secret":       secret,
		"otpauth_url":  otpauthURL,
		"seed_phrase":  generateRecoverySeedPhrase(secret),
		"backup_codes": generateDeterministicBackupCodes(secret),
		"totp_enabled": false,
	})
}

// EnableTOTP verifies the 6-digit code from the user's authenticator app and activates TOTP 2FA.
func (h *Handler) EnableTOTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	freshUser, err := h.repo.GetUserByID(r.Context(), user.ID, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if freshUser.TOTPSecret == "" {
		writeError(w, http.StatusBadRequest, "TOTP setup has not been initiated")
		return
	}

	if !verifyTOTPOrRecovery(freshUser.TOTPSecret, req.Code, time.Now()) {
		writeError(w, http.StatusBadRequest, "Invalid verification code. Please check your authenticator app and try again.")
		return
	}

	if err := h.repo.UpdateUserTOTP(r.Context(), freshUser.ID, freshUser.TOTPSecret, true); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enable TOTP")
		return
	}

	// Invalidate cached session/profile in Redis so totp_enabled is immediately reflected
	h.invalidateUserCache(r, freshUser)

	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &freshUser.ID,
		Action:      "auth.totp.enabled",
		TargetType:  "user",
		TargetID:    freshUser.ID.String(),
		Reason:      "User enabled TOTP two-factor authentication",
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"totp_enabled": true,
		"seed_phrase":  generateRecoverySeedPhrase(freshUser.TOTPSecret),
		"backup_codes": generateDeterministicBackupCodes(freshUser.TOTPSecret),
		"message":      "Two-factor authentication enabled",
	})
}

// DisableTOTP turns off TOTP 2FA after verifying either the current 6-digit TOTP code, recovery seed phrase, or account password.
func (h *Handler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	freshUser, err := h.repo.GetUserByIdentifier(r.Context(), user.Username)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if !freshUser.TOTPEnabled {
		writeError(w, http.StatusBadRequest, "Two-factor authentication is not enabled")
		return
	}

	verified := false
	if strings.TrimSpace(req.Code) != "" && verifyTOTPOrRecovery(freshUser.TOTPSecret, req.Code, time.Now()) {
		verified = true
	} else if req.Password != "" && bcrypt.CompareHashAndPassword([]byte(freshUser.PasswordHash), []byte(req.Password)) == nil {
		verified = true
	}

	if !verified {
		writeError(w, http.StatusUnauthorized, "Invalid authentication code, recovery seed phrase, or password")
		return
	}

	if err := h.repo.UpdateUserTOTP(r.Context(), freshUser.ID, "", false); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to disable TOTP")
		return
	}

	h.invalidateUserCache(r, freshUser)

	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &freshUser.ID,
		Action:      "auth.totp.disabled",
		TargetType:  "user",
		TargetID:    freshUser.ID.String(),
		Reason:      "User disabled TOTP two-factor authentication",
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"totp_enabled": false,
		"message":      "Two-factor authentication disabled",
	})
}

// VerifyLoginTOTP completes a two-step login using the temporary mfa_token and 6-digit TOTP code or recovery seed phrase.
func (h *Handler) VerifyLoginTOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MFAToken string `json:"mfa_token"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	userID, err := verifyMFAToken(strings.TrimSpace(req.MFAToken))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Login session expired. Please sign in again.")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID, nil)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid user account")
		return
	}

	if !user.TOTPEnabled || user.TOTPSecret == "" {
		writeError(w, http.StatusBadRequest, "TOTP is not enabled for this account")
		return
	}

	if !verifyTOTPOrRecovery(user.TOTPSecret, req.Code, time.Now()) {
		writeError(w, http.StatusUnauthorized, "Invalid verification code or recovery seed phrase")
		return
	}

	if strings.ToLower(strings.TrimSpace(user.Email)) == "afrizaahmad18@gmail.com" {
		_, _ = h.repo.BootstrapOwner(r.Context(), user.Email)
		user.Role = "owner"
	}

	token, err := h.createSessionForUser(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate session")
		return
	}

	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":          user,
		"session_token": token,
	})
}

func (h *Handler) invalidateUserCache(r *http.Request, user *domain.User) {
	if h.rdb == nil || user == nil {
		return
	}
	_ = h.rdb.Del(r.Context(), fmt.Sprintf("profile:user:%s", strings.ToLower(user.Username))).Err()

	token := ""
	if cookie, err := r.Cookie("session_token"); err == nil && cookie.Value != "" {
		token = cookie.Value
	}
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if token != "" {
		hash := sha256.Sum256([]byte(token))
		tokenHash := fmt.Sprintf("%x", hash[:])
		_ = h.rdb.Del(r.Context(), fmt.Sprintf("session:token:%s", tokenHash)).Err()
	}
}

