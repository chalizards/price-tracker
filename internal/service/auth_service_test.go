package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-key-that-is-long-enough"

func newTestAuthService() *AuthService {
	return &AuthService{
		jwtSecret: []byte(testJWTSecret),
	}
}

func TestGenerateAndValidateJWT(test *testing.T) {
	authService := newTestAuthService()
	user := &models.User{
		ID:    1,
		Email: "pikachu@gmail.com",
	}

	tokenStr, err := authService.generateJWT(user)
	if err != nil {
		test.Fatalf("generateJWT() error = %v", err)
	}

	claims, err := authService.ValidateJWT(tokenStr)
	if err != nil {
		test.Fatalf("ValidateJWT() error = %v", err)
	}

	if claims.UserID != user.ID {
		test.Errorf("expected UserID %d, got %d", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		test.Errorf("expected Email %s, got %s", user.Email, claims.Email)
	}
}

func TestValidateJWT_Expired(test *testing.T) {
	authService := newTestAuthService()

	claims := &Claims{
		UserID: 1,
		Email:  "pikachu@gmail.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(authService.jwtSecret)
	if err != nil {
		test.Fatalf("failed to sign token: %v", err)
	}

	_, err = authService.ValidateJWT(tokenStr)
	if err == nil {
		test.Error("expected error for expired token, got nil")
	}
}

func TestValidateJWT_InvalidSignature(test *testing.T) {
	authService := newTestAuthService()

	claims := &Claims{
		UserID: 1,
		Email:  "pikachu@gmail.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("wrong-secret"))
	if err != nil {
		test.Fatalf("failed to sign token: %v", err)
	}

	_, err = authService.ValidateJWT(tokenStr)
	if err == nil {
		test.Error("expected error for invalid signature, got nil")
	}
}

func TestValidateJWT_InvalidFormat(test *testing.T) {
	authService := newTestAuthService()

	_, err := authService.ValidateJWT("not-a-valid-token")
	if err == nil {
		test.Error("expected error for invalid token format, got nil")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	state2, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	if state1 == "" {
		t.Error("expected non-empty state")
	}
	if state1 == state2 {
		t.Error("expected different states on successive calls")
	}
}

func TestFetchGoogleUserInfo_Success(test *testing.T) {
	expected := GoogleUserInfo{
		Sub:           "123456",
		Email:         "pikachu@gmail.com",
		Name:          "Pikachu",
		Picture:       "https://pikachu.com/photo.jpg",
		EmailVerified: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			test.Errorf("expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expected); err != nil {
			test.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	originalClient := googleHTTPClient
	googleHTTPClient = server.Client()
	defer func() { googleHTTPClient = originalClient }()

	result, err := fetchGoogleUserInfoFromURL(context.Background(), server.URL, "test-token")

	if err != nil {
		test.Fatalf("fetchGoogleUserInfo() error = %v", err)
	}

	if result.Sub != expected.Sub {
		test.Errorf("expected Sub %s, got %s", expected.Sub, result.Sub)
	}
	if result.Email != expected.Email {
		test.Errorf("expected Email %s, got %s", expected.Email, result.Email)
	}
	if result.EmailVerified != expected.EmailVerified {
		test.Errorf("expected EmailVerified %v, got %v", expected.EmailVerified, result.EmailVerified)
	}
}

func TestFetchGoogleUserInfo_Error(test *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"error": "invalid_token"}`)); err != nil {
			test.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	originalClient := googleHTTPClient
	googleHTTPClient = server.Client()
	defer func() { googleHTTPClient = originalClient }()

	_, err := fetchGoogleUserInfoFromURL(context.Background(), server.URL, "bad-token")
	if err == nil {
		test.Error("expected error for unauthorized response, got nil")
	}
}

func TestGetGoogleLoginURL(test *testing.T) {
	authService := NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)

	url, state, err := authService.GetGoogleLoginURL()
	if err != nil {
		test.Fatalf("GetGoogleLoginURL() error = %v", err)
	}

	if url == "" {
		test.Error("expected non-empty URL")
	}
	if state == "" {
		test.Error("expected non-empty state")
	}
}
