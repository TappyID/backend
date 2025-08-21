package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ConnectionStatus represents the status of a user connection
type ConnectionStatus string

const (
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusError        ConnectionStatus = "error"
)

// Platform represents the social media platform
type Platform string

const (
	PlatformWhatsApp  Platform = "whatsapp"
	PlatformFacebook  Platform = "facebook"
	PlatformInstagram Platform = "instagram"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformEmail     Platform = "email"
	PlatformSMS       Platform = "sms"
)

// SessionData represents additional session metadata
type SessionData map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (s SessionData) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *SessionData) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, s)
}

// UserConnection represents a user's connection to a social platform
type UserConnection struct {
	ID             uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID        `json:"user_id" gorm:"type:uuid;not null"`
	Platform       Platform         `json:"platform" gorm:"type:varchar(50);not null"`
	Status         ConnectionStatus `json:"status" gorm:"type:varchar(50);not null"`
	SessionName    *string          `json:"session_name,omitempty" gorm:"type:varchar(255)"`
	SessionData    SessionData      `json:"session_data,omitempty" gorm:"type:jsonb"`
	ConnectedAt    *time.Time       `json:"connected_at,omitempty"`
	DisconnectedAt *time.Time       `json:"disconnected_at,omitempty"`
	LastSyncAt     time.Time        `json:"last_sync_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateUserConnectionRequest represents the request to create/update a connection
type CreateUserConnectionRequest struct {
	Platform    Platform         `json:"platform" validate:"required"`
	Status      ConnectionStatus `json:"status" validate:"required"`
	SessionName *string          `json:"session_name,omitempty"`
	SessionData SessionData      `json:"session_data,omitempty"`
}

// UpdateUserConnectionRequest represents the request to update a connection
type UpdateUserConnectionRequest struct {
	Status      *ConnectionStatus `json:"status,omitempty"`
	SessionName *string           `json:"session_name,omitempty"`
	SessionData SessionData       `json:"session_data,omitempty"`
}

// UserConnectionResponse represents the response for connection data
type UserConnectionResponse struct {
	ID          uuid.UUID        `json:"id"`
	Platform    Platform         `json:"platform"`
	Status      ConnectionStatus `json:"status"`
	SessionName *string          `json:"session_name,omitempty"`
	SessionData SessionData      `json:"session_data,omitempty"`
	ConnectedAt *time.Time       `json:"connected_at,omitempty"`
	LastSyncAt  time.Time        `json:"last_sync_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
