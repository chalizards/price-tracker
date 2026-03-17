package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test-secret-key-that-is-long-enough"

func setupTestRouter(authSvc *service.AuthService, userRepo *repository.UserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(authSvc, userRepo))
	router.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	return router
}

func TestAuthMiddleware_MissingHeader(test *testing.T) {
	authSvc := service.NewAuthService("id", "secret", "http://localhost/callback", testJWTSecret, nil)
	router := setupTestRouter(authSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(test *testing.T) {
	authSvc := service.NewAuthService("id", "secret", "http://localhost/callback", testJWTSecret, nil)
	router := setupTestRouter(authSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAuthMiddleware_InvalidToken(test *testing.T) {
	authSvc := service.NewAuthService("id", "secret", "http://localhost/callback", testJWTSecret, nil)
	router := setupTestRouter(authSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(test *testing.T) {
	authService := service.NewAuthService("id", "secret", "http://localhost/callback", testJWTSecret, nil)
	router := setupTestRouter(authService, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE2MDAwMDAwMDAsImlhdCI6MTYwMDAwMDAwMH0.invalid")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAuthMiddleware_InvalidCookie(test *testing.T) {
	authSvc := service.NewAuthService("id", "secret", "http://localhost/callback", testJWTSecret, nil)
	router := setupTestRouter(authSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: jwtCookieName, Value: "invalid-token"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
