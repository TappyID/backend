package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"tappyone/internal/models"
	"tappyone/internal/repositories"
)

type ConnectionService struct {
	connectionRepo *repositories.ConnectionRepository
	wahaAPIURL     string
	wahaAPIKey     string
}

func NewConnectionService(connectionRepo *repositories.ConnectionRepository, wahaAPIURL, wahaAPIKey string) *ConnectionService {
	return &ConnectionService{
		connectionRepo: connectionRepo,
		wahaAPIURL:     wahaAPIURL,
		wahaAPIKey:     wahaAPIKey,
	}
}

// GetUserConnection retrieves a user's connection for a specific platform
func (s *ConnectionService) GetUserConnection(userID uuid.UUID, platform models.Platform) (*models.UserConnection, error) {
	connection, err := s.connectionRepo.GetUserConnection(userID, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get user connection: %w", err)
	}
	
	if connection == nil {
		return nil, nil
	}
	
	return connection, nil
}

// GetUserConnections retrieves all connections for a user
func (s *ConnectionService) GetUserConnections(userID uuid.UUID) ([]models.UserConnection, error) {
	connections, err := s.connectionRepo.GetUserConnections(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user connections: %w", err)
	}
	
	return connections, nil
}

// CreateOrUpdateConnection creates or updates a user's connection
func (s *ConnectionService) CreateOrUpdateConnection(userID uuid.UUID, req *models.CreateUserConnectionRequest) (*models.UserConnection, error) {
	connection, err := s.connectionRepo.CreateOrUpdateUserConnection(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update connection: %w", err)
	}
	
	return connection, nil
}

// UpdateConnection updates an existing user connection
func (s *ConnectionService) UpdateConnection(userID uuid.UUID, platform models.Platform, req *models.UpdateUserConnectionRequest) (*models.UserConnection, error) {
	connection, err := s.connectionRepo.UpdateUserConnection(userID, platform, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update connection: %w", err)
	}
	
	return connection, nil
}

// SyncWhatsAppConnection synchronizes WhatsApp connection status with WAHA API
func (s *ConnectionService) SyncWhatsAppConnection(userID uuid.UUID, sessionName string) (*models.UserConnection, error) {
	// Check status in WAHA API
	wahaStatus, err := s.checkWAHASessionStatus(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check WAHA status: %w", err)
	}
	
	// Convert WAHA status to our status
	var status models.ConnectionStatus
	switch wahaStatus {
	case "WORKING":
		status = models.ConnectionStatusConnected
	case "STARTING", "SCAN_QR_CODE":
		status = models.ConnectionStatusConnecting
	case "FAILED":
		status = models.ConnectionStatusError
	case "STOPPED":
		status = models.ConnectionStatusDisconnected
	default:
		status = models.ConnectionStatusDisconnected
	}
	
	// Update in database
	updateReq := &models.UpdateUserConnectionRequest{
		Status:      &status,
		SessionName: &sessionName,
		SessionData: models.SessionData{
			"waha_status":    wahaStatus,
			"last_sync":      time.Now().Format(time.RFC3339),
			"session_name":   sessionName,
		},
	}
	
	return s.UpdateConnection(userID, models.PlatformWhatsApp, updateReq)
}

// checkWAHASessionStatus checks the status of a session in WAHA API
func (s *ConnectionService) checkWAHASessionStatus(sessionName string) (string, error) {
	url := fmt.Sprintf("%s/sessions/%s", s.wahaAPIURL, sessionName)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("X-Api-Key", s.wahaAPIKey)
	req.Header.Set("Accept", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return "STOPPED", nil
	}
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("WAHA API returned status %d", resp.StatusCode)
	}
	
	var result struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	return result.Status, nil
}

// DisconnectWhatsApp disconnects WhatsApp session both in WAHA and database
func (s *ConnectionService) DisconnectWhatsApp(userID uuid.UUID, sessionName string) error {
	// Delete session in WAHA API
	url := fmt.Sprintf("%s/sessions/%s", s.wahaAPIURL, sessionName)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	
	req.Header.Set("X-Api-Key", s.wahaAPIKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete WAHA session: %w", err)
	}
	defer resp.Body.Close()
	
	// Update status in database regardless of WAHA response
	updateReq := &models.UpdateUserConnectionRequest{
		Status: func() *models.ConnectionStatus {
			status := models.ConnectionStatusDisconnected
			return &status
		}(),
		SessionData: models.SessionData{
			"disconnected_at": time.Now().Format(time.RFC3339),
			"waha_deleted":    resp.StatusCode == 200,
		},
	}
	
	_, err = s.UpdateConnection(userID, models.PlatformWhatsApp, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update connection status: %w", err)
	}
	
	return nil
}

// toResponse converts a UserConnection model to response format
func (s *ConnectionService) toResponse(connection *models.UserConnection) *models.UserConnectionResponse {
	return &models.UserConnectionResponse{
		ID:          connection.ID,
		Platform:    connection.Platform,
		Status:      connection.Status,
		SessionName: connection.SessionName,
		SessionData: connection.SessionData,
		ConnectedAt: connection.ConnectedAt,
		LastSyncAt:  connection.LastSyncAt,
		CreatedAt:   connection.CreatedAt,
		UpdatedAt:   connection.UpdatedAt,
	}
}
