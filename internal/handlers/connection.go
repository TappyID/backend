package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"tappyone/internal/models"
	"tappyone/internal/services"
)

type ConnectionHandler struct {
	connectionService *services.ConnectionService
}

func NewConnectionHandler(connectionService *services.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{
		connectionService: connectionService,
	}
}

// GetUserConnections retrieves all connections for the authenticated user
func (h *ConnectionHandler) GetUserConnections(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	connections, err := h.connectionService.GetUserConnections(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"connections": connections})
}

// GetWhatsAppConnection retrieves WhatsApp connection for the authenticated user
func (h *ConnectionHandler) GetWhatsAppConnection(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	connection, err := h.connectionService.GetUserConnection(userID, models.PlatformWhatsApp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connection"})
		return
	}

	if connection == nil {
		// Criar conexão inicial se não existir
		createReq := &models.CreateUserConnectionRequest{
			Platform: models.PlatformWhatsApp,
			Status:   models.ConnectionStatusDisconnected,
		}
		
		connection, err = h.connectionService.CreateOrUpdateConnection(userID, createReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create connection"})
			return
		}
	}

	c.JSON(http.StatusOK, connection)
}

// GetUserConnection retrieves a specific connection for the authenticated user
func (h *ConnectionHandler) GetUserConnection(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	platform := models.Platform(c.Param("platform"))

	// Validate platform
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
		return
	}

	connection, err := h.connectionService.GetUserConnection(userID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get connection"})
		return
	}

	if connection == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	c.JSON(http.StatusOK, connection)
}

// CreateOrUpdateConnection creates or updates a connection
func (h *ConnectionHandler) CreateOrUpdateConnection(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.CreateUserConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate platform
	if !isValidPlatform(req.Platform) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
		return
	}

	// Validate status
	if !isValidStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	connection, err := h.connectionService.CreateOrUpdateConnection(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create/update connection"})
		return
	}

	c.JSON(http.StatusOK, connection)
}

// UpdateConnection updates an existing connection
func (h *ConnectionHandler) UpdateConnection(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	platform := models.Platform(c.Param("platform"))

	// Validate platform
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
		return
	}

	var req models.UpdateUserConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate status if provided
	if req.Status != nil && !isValidStatus(*req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	connection, err := h.connectionService.UpdateConnection(userID, platform, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connection"})
		return
	}

	c.JSON(http.StatusOK, connection)
}

// SyncWhatsAppConnection synchronizes WhatsApp connection with WAHA API
func (h *ConnectionHandler) SyncWhatsAppConnection(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	sessionName := c.Param("sessionName")
	if sessionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session name is required"})
		return
	}

	connection, err := h.connectionService.SyncWhatsAppConnection(userID, sessionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync WhatsApp connection"})
		return
	}

	c.JSON(http.StatusOK, connection)
}

// DisconnectWhatsApp disconnects WhatsApp session
func (h *ConnectionHandler) DisconnectWhatsApp(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	sessionName := c.Param("sessionName")
	if sessionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session name is required"})
		return
	}

	err = h.connectionService.DisconnectWhatsApp(userID, sessionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect WhatsApp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "WhatsApp disconnected successfully"})
}

// Helper functions for validation
func isValidPlatform(platform models.Platform) bool {
	switch platform {
	case models.PlatformWhatsApp, models.PlatformFacebook, models.PlatformInstagram,
		 models.PlatformLinkedIn, models.PlatformEmail, models.PlatformSMS:
		return true
	default:
		return false
	}
}

func isValidStatus(status models.ConnectionStatus) bool {
	switch status {
	case models.ConnectionStatusDisconnected, models.ConnectionStatusConnecting,
		 models.ConnectionStatusConnected, models.ConnectionStatusError:
		return true
	default:
		return false
	}
}
