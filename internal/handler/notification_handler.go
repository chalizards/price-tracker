package handler

import (
	"net/http"
	"strconv"

	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	repo *repository.NotificationRepository
}

func NewNotificationHandler(repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func (handler *NotificationHandler) GetUnreadNotifications(ctx *gin.Context) {
	notifications, err := handler.repo.GetUnread(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread notifications"})
		return
	}

	ctx.JSON(http.StatusOK, notifications)
}

func (handler *NotificationHandler) GetNotificationsByProductID(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	notifications, err := handler.repo.GetByProductID(ctx.Request.Context(), productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
		return
	}

	ctx.JSON(http.StatusOK, notifications)
}

func (handler *NotificationHandler) MarkAsRead(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	if err := handler.repo.MarkAsRead(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (handler *NotificationHandler) MarkAllAsRead(ctx *gin.Context) {
	if err := handler.repo.MarkAllAsRead(ctx.Request.Context()); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}
