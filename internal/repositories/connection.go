package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"tappyone/internal/models"
)

type ConnectionRepository struct {
	db *gorm.DB
}

func NewConnectionRepository(db *gorm.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

// GetUserConnection retrieves a user's connection for a specific platform
func (r *ConnectionRepository) GetUserConnection(userID uuid.UUID, platform models.Platform) (*models.UserConnection, error) {
	var connection models.UserConnection
	
	err := r.db.Where("user_id = ? AND platform = ?", userID, platform).First(&connection).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &connection, nil
}

// GetUserConnections retrieves all connections for a user
func (r *ConnectionRepository) GetUserConnections(userID uuid.UUID) ([]models.UserConnection, error) {
	var connections []models.UserConnection
	
	err := r.db.Where("user_id = ?", userID).Order("platform").Find(&connections).Error
	if err != nil {
		return nil, err
	}
	
	return connections, nil
}

// CreateOrUpdateUserConnection creates or updates a user's connection
func (r *ConnectionRepository) CreateOrUpdateUserConnection(userID uuid.UUID, req *models.CreateUserConnectionRequest) (*models.UserConnection, error) {
	now := time.Now()
	
	connection := models.UserConnection{
		UserID:      userID,
		Platform:    req.Platform,
		Status:      req.Status,
		SessionName: req.SessionName,
		SessionData: req.SessionData,
		LastSyncAt:  now,
	}
	
	if req.Status == models.ConnectionStatusConnected {
		connection.ConnectedAt = &now
	} else if req.Status == models.ConnectionStatusDisconnected {
		connection.DisconnectedAt = &now
	}
	
	// Use GORM's Upsert functionality
	err := r.db.Where("user_id = ? AND platform = ?", userID, req.Platform).
		Assign(map[string]interface{}{
			"status":          req.Status,
			"session_name":    req.SessionName,
			"session_data":    req.SessionData,
			"connected_at":    connection.ConnectedAt,
			"disconnected_at": connection.DisconnectedAt,
			"last_sync_at":    now,
			"updated_at":      now,
		}).
		FirstOrCreate(&connection).Error
	
	if err != nil {
		return nil, err
	}
	
	return &connection, nil
}

// UpdateUserConnection updates an existing user connection
func (r *ConnectionRepository) UpdateUserConnection(userID uuid.UUID, platform models.Platform, req *models.UpdateUserConnectionRequest) (*models.UserConnection, error) {
	var connection models.UserConnection
	
	// Find existing connection
	err := r.db.Where("user_id = ? AND platform = ?", userID, platform).First(&connection).Error
	if err != nil {
		return nil, err
	}
	
	// Update fields
	updates := map[string]interface{}{
		"last_sync_at": time.Now(),
		"updated_at":   time.Now(),
	}
	
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == models.ConnectionStatusConnected {
			updates["connected_at"] = time.Now()
			updates["disconnected_at"] = nil
		} else if *req.Status == models.ConnectionStatusDisconnected {
			updates["disconnected_at"] = time.Now()
		}
	}
	
	if req.SessionName != nil {
		updates["session_name"] = *req.SessionName
	}
	
	if req.SessionData != nil {
		updates["session_data"] = req.SessionData
	}
	
	err = r.db.Model(&connection).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	
	// Reload the updated connection
	err = r.db.Where("user_id = ? AND platform = ?", userID, platform).First(&connection).Error
	if err != nil {
		return nil, err
	}
	
	return &connection, nil
}

// DeleteUserConnection deletes a user's connection
func (r *ConnectionRepository) DeleteUserConnection(userID uuid.UUID, platform models.Platform) error {
	return r.db.Where("user_id = ? AND platform = ?", userID, platform).Delete(&models.UserConnection{}).Error
}

// GetConnectionsByStatus retrieves connections by status (for monitoring/admin)
func (r *ConnectionRepository) GetConnectionsByStatus(status models.ConnectionStatus) ([]models.UserConnection, error) {
	var connections []models.UserConnection
	
	err := r.db.Where("status = ?", status).Order("updated_at DESC").Find(&connections).Error
	if err != nil {
		return nil, err
	}
	
	return connections, nil
}
