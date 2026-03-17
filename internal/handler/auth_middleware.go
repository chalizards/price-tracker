package handler

import (
	"net/http"
	"strings"

	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func AuthMiddleware(authService *service.AuthService, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var tokenString string

		if authHeader := ctx.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			cookie, err := ctx.Cookie(jwtCookieName)
			if err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
			return
		}

		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		user, err := userRepo.FindByID(ctx.Request.Context(), claims.UserID)
		if err != nil {
			if err == pgx.ErrNoRows {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		ctx.Set("user", user)
		ctx.Set("userID", claims.UserID)
		ctx.Set("email", claims.Email)

		ctx.Next()
	}
}
