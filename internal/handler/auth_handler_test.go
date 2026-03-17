package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

func TestAuthHandler_GoogleLogin(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	handler := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.GET("/auth/google/login", handler.GoogleLogin)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		test.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location == "" {
		test.Error("expected Location header to be set")
	}
}

func TestAuthHandler_GoogleCallback_MissingStateCookie(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	handler := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.GET("/auth/google/callback", handler.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=abc&code=xyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		test.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAuthHandler_GoogleCallback_InvalidState(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	handler := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.GET("/auth/google/callback", handler.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=wrong&code=xyz", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "correct"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		test.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAuthHandler_GoogleCallback_MissingCode(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	handler := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.GET("/auth/google/callback", handler.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=abc", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "abc"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		test.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestAuthHandler_Me_NoUserInContext(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	handler := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.GET("/auth/me", handler.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		test.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAuthHandler_Me_WithUser(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	handler := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.GET("/auth/me", func(ctx *gin.Context) {
		ctx.Set("user", &models.User{ID: 1, Email: "pikachu@gmail.com", Name: "Pikachu"})
		ctx.Next()
	}, handler.Me)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		test.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAuthHandler_Logout(test *testing.T) {
	gin.SetMode(gin.TestMode)

	authSvc := service.NewAuthService("client-id", "client-secret", "http://localhost/callback", testJWTSecret, nil)
	h := NewAuthHandler(authSvc, "http://localhost:3000")

	router := gin.New()
	router.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: jwtCookieName, Value: "some-token"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		test.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	cookies := recorder.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == jwtCookieName && c.MaxAge < 0 {
			found = true
			break
		}
	}
	if !found {
		test.Error("expected auth_token cookie to be cleared")
	}
}
