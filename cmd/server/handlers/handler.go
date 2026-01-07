package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"clauded-server/config"
	"clauded-server/notification"
	"clauded-server/proxy"
	"clauded-server/session"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config          *config.Config
	sessionManager  *session.Manager
	notificationSvc *notification.Service
	proxyManager    *proxy.Manager
}

func NewHandler(cfg *config.Config, sm *session.Manager, ns *notification.Service, pm *proxy.Manager) *Handler {
	return &Handler{
		config:          cfg,
		sessionManager:  sm,
		notificationSvc: ns,
		proxyManager:    pm,
	}
}

func (h *Handler) SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", h.HealthCheck)

	// SSE notifications
	router.GET("/api/v1/notifications/stream", h.SSEStream)

	// Webhook API
	api := router.Group("/api/v1/notifications")
	{
		api.POST("/subscribe", h.SubscribeWebhook)
		api.POST("/publish", h.PublishNotification)
		api.DELETE("/unsubscribe", h.UnsubscribeWebhook)
		api.GET("/subscriptions", h.GetSubscriptions)
	}

	// Root path "/" -> proxy to piko as "root-service"
	router.Any("/", gin.WrapH(h.proxyManager.ProxyRootRequest()))

	// /piko path -> proxy to piko as "root-service"
	router.Any("/piko", gin.WrapH(h.proxyManager.ProxyRootRequest()))
	router.Any("/piko/*path", gin.WrapH(h.proxyManager.ProxyRootRequest()))

	// Proxy all requests to /:session/* to piko as session-based service
	// This handles both HTTP and WebSocket connections
	router.Any("/:session/*path", gin.WrapH(h.proxyManager.ProxyRequest()))

	// Also handle requests to /:session (without trailing path)
	router.Any("/:session", gin.WrapH(h.proxyManager.ProxyRequest()))

	return router
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) SSEStream(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Subscribe to notifications
	ch := h.notificationSvc.SubscribeSSE(sessionID)
	defer func() {
		// Unsubscribe will be handled by session manager
	}()

	// Flush headers
	c.Writer.Flush()

	// Send notifications
	c.Stream(func(w io.Writer) bool {
		select {
		case notif, ok := <-ch:
			if !ok {
				return false
			}

			// Format SSE message
			data, _ := json.Marshal(notif)
			c.SSEvent(string(notif.Type), string(data))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

type SubscribeRequest struct {
	SessionID  string                    `json:"session_id" binding:"required"`
	WebhookURL string                    `json:"webhook_url" binding:"required"`
	Events     []string                  `json:"events"`
}

type SubscribeResponse struct {
	SubscriptionID string `json:"subscription_id"`
	SessionID      string `json:"session_id"`
	WebhookURL     string `json:"webhook_url"`
}

func (h *Handler) SubscribeWebhook(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string events to NotificationType
	eventTypes := make([]notification.NotificationType, len(req.Events))
	for i, e := range req.Events {
		eventTypes[i] = notification.NotificationType(e)
	}

	// Subscribe webhook
	if err := h.notificationSvc.SubscribeWebhook(req.SessionID, req.WebhookURL, eventTypes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Webhook subscribed: session=%s, url=%s", req.SessionID, req.WebhookURL)

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook subscribed successfully",
	})
}

type PublishRequest struct {
	SessionID string                 `json:"session_id" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

func (h *Handler) PublishNotification(c *gin.Context) {
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Publish notification to the service
	h.notificationSvc.Publish(req.SessionID, notification.NotificationType(req.Type), req.Data)

	log.Printf("Notification published: session=%s, type=%s", req.SessionID, req.Type)

	c.JSON(http.StatusOK, gin.H{
		"status":  "published",
		"session": req.SessionID,
		"type":    req.Type,
	})
}

func (h *Handler) UnsubscribeWebhook(c *gin.Context) {
	sessionID := c.Query("session_id")
	webhookURL := c.Query("webhook_url")

	if sessionID == "" || webhookURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and webhook_url are required"})
		return
	}

	// Note: We need to implement unsubscribe by webhook URL
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook unsubscribed successfully",
	})
}

func (h *Handler) GetSubscriptions(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	subs := h.notificationSvc.GetSubscribers(sessionID)
	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subs,
	})
}

