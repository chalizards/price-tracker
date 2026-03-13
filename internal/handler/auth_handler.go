package handler

import (
	"net/http"

	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
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

	token, user, err := handler.authService.HandleGoogleCallback(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (handler *AuthHandler) Me(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}
