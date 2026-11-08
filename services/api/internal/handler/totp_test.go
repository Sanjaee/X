package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/repository"
)

func TestTOTPFlow_SetupEnableLoginDisable(t *testing.T) {
	repo := repository.NewMemoryRepository()
	h := NewHandler(repo, t.TempDir(), nil, nil, nil, nil)

	// 1. Register a user
	regBody, _ := json.Marshal(map[string]string{
		"username":     "totpuser",
		"email":        "totpuser@example.com",
		"password":     "Password123!",
		"display_name": "TOTP User",
	})
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	recReg := httptest.NewRecorder()
	h.Register(recReg, reqReg)
	if recReg.Code != http.StatusCreated {
		t.Fatalf("expected 201 on register, got %d: %s", recReg.Code, recReg.Body.String())
	}

	var regResp struct {
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(recReg.Body.Bytes(), &regResp)
	authMiddleware := middleware.AuthMiddleware(repo, false)

	// 2. Setup TOTP
	reqSetup := httptest.NewRequest(http.MethodPost, "/api/v1/me/totp/setup", nil)
	reqSetup.Header.Set("Authorization", "Bearer "+regResp.SessionToken)
	recSetup := httptest.NewRecorder()
	authMiddleware(http.HandlerFunc(h.SetupTOTP)).ServeHTTP(recSetup, reqSetup)
	if recSetup.Code != http.StatusOK {
		t.Fatalf("expected 200 on setup totp, got %d: %s", recSetup.Code, recSetup.Body.String())
	}

	var setupResp struct {
		Secret     string `json:"secret"`
		OTPAuthURL string `json:"otpauth_url"`
	}
	_ = json.Unmarshal(recSetup.Body.Bytes(), &setupResp)
	if setupResp.Secret == "" || setupResp.OTPAuthURL == "" {
		t.Fatalf("expected secret and otpauth_url in setup response, got %+v", setupResp)
	}

	// 3. Enable TOTP with valid code
	validCode, err := computeTOTPCode(setupResp.Secret, time.Now().Unix()/30)
	if err != nil {
		t.Fatalf("failed to compute totp code: %v", err)
	}

	enableBody, _ := json.Marshal(map[string]string{"code": validCode})
	reqEnable := httptest.NewRequest(http.MethodPost, "/api/v1/me/totp/enable", bytes.NewReader(enableBody))
	reqEnable.Header.Set("Authorization", "Bearer "+regResp.SessionToken)
	recEnable := httptest.NewRecorder()
	authMiddleware(http.HandlerFunc(h.EnableTOTP)).ServeHTTP(recEnable, reqEnable)
	if recEnable.Code != http.StatusOK {
		t.Fatalf("expected 200 on enable totp, got %d: %s", recEnable.Code, recEnable.Body.String())
	}

	// 4. Login with password only -> should require TOTP
	loginBody, _ := json.Marshal(map[string]string{
		"identifier": "totpuser",
		"password":   "Password123!",
	})
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	recLogin := httptest.NewRecorder()
	h.Login(recLogin, reqLogin)
	if recLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 on login challenge, got %d: %s", recLogin.Code, recLogin.Body.String())
	}

	var loginChallenge struct {
		RequiresTOTP bool   `json:"requires_totp"`
		MFAToken     string `json:"mfa_token"`
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(recLogin.Body.Bytes(), &loginChallenge)
	if !loginChallenge.RequiresTOTP || loginChallenge.MFAToken == "" || loginChallenge.SessionToken != "" {
		t.Fatalf("expected requires_totp=true and mfa_token without session_token, got %+v", loginChallenge)
	}

	// 5. Verify login with wrong TOTP code -> should fail 401
	badVerifyBody, _ := json.Marshal(map[string]string{
		"mfa_token": loginChallenge.MFAToken,
		"code":      "000000",
	})
	reqBadVerify := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp", bytes.NewReader(badVerifyBody))
	recBadVerify := httptest.NewRecorder()
	h.VerifyLoginTOTP(recBadVerify, reqBadVerify)
	if recBadVerify.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong totp code, got %d", recBadVerify.Code)
	}

	// 6. Verify login with valid TOTP code -> should return session_token
	goodVerifyBody, _ := json.Marshal(map[string]string{
		"mfa_token": loginChallenge.MFAToken,
		"code":      validCode,
	})
	reqGoodVerify := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/totp", bytes.NewReader(goodVerifyBody))
	recGoodVerify := httptest.NewRecorder()
	h.VerifyLoginTOTP(recGoodVerify, reqGoodVerify)
	if recGoodVerify.Code != http.StatusOK {
		t.Fatalf("expected 200 on valid totp verify, got %d: %s", recGoodVerify.Code, recGoodVerify.Body.String())
	}

	var finalAuth struct {
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(recGoodVerify.Body.Bytes(), &finalAuth)
	if finalAuth.SessionToken == "" {
		t.Fatalf("expected session_token after valid totp verification")
	}

	// 7. Disable TOTP
	disableBody, _ := json.Marshal(map[string]string{"code": validCode})
	reqDisable := httptest.NewRequest(http.MethodPost, "/api/v1/me/totp/disable", bytes.NewReader(disableBody))
	reqDisable.Header.Set("Authorization", "Bearer "+finalAuth.SessionToken)
	recDisable := httptest.NewRecorder()
	authMiddleware(http.HandlerFunc(h.DisableTOTP)).ServeHTTP(recDisable, reqDisable)
	if recDisable.Code != http.StatusOK {
		t.Fatalf("expected 200 on disable totp, got %d: %s", recDisable.Code, recDisable.Body.String())
	}
}

