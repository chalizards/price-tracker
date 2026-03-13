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
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, use: Bearer <token>"})
			return
		}

		claims, err := authService.ValidateJWT(parts[1])
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
