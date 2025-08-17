package utils

import (
	"errors"
	"log"
	"tappyone/internal/services"

	"github.com/gin-gonic/gin"
)

// ValidateJWTFromHeader valida JWT do header Authorization e retorna userID
func ValidateJWTFromHeader(c *gin.Context, authService *services.AuthService) (string, error) {
	// Extrair token do header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Printf("Missing Authorization header")
		return "", errors.New("authorization header required")
	}

	// Extrair token (Bearer <token>)
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		log.Printf("Invalid authorization format: %s", authHeader)
		return "", errors.New("invalid authorization format")
	}

	// Validar token
	claims, err := authService.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		return "", err
	}

	return claims.UserID, nil
}
