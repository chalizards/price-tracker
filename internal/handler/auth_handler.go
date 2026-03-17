package handler

import (
	"net/http"

	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

const jwtCookieName = "auth_token"
const jwtCookieMaxAge = 86400 // 24h, matching JWT expiry

type AuthHandler struct {
	authService *service.AuthService
	frontendURL string
}

func NewAuthHandler(authService *service.AuthService, frontendURL string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		frontendURL: frontendURL,
	}
}

func (handler *AuthHandler) GoogleLogin(ctx *gin.Context) {
	url, state, err := handler.authService.GetGoogleLoginURL()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate login"})
		return
	}
	ctx.SetCookie("oauth_state", state, 300, "/", "", false, true)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

func (handler *AuthHandler) GoogleCallback(ctx *gin.Context) {
	stateCookie, err := ctx.Cookie("oauth_state")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing state cookie"})
		return
	}

	stateParam := ctx.Query("state")
	if stateParam != stateCookie {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	token, _, err := handler.authService.HandleGoogleCallback(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(jwtCookieName, token, jwtCookieMaxAge, "/", "", false, true)

	ctx.Redirect(http.StatusTemporaryRedirect, handler.frontendURL)
}

func (handler *AuthHandler) Logout(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(jwtCookieName, "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (handler *AuthHandler) Me(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}
