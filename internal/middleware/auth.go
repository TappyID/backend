package middleware

import (
	"log"
	"net/http"
	"strings"
	"tappyone/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter token do header Authorization
		authHeader := c.GetHeader("Authorization")
		log.Printf("[AUTH] Path: %s, AuthHeader: %s", c.Request.URL.Path, authHeader)
		if authHeader == "" {
			log.Printf("[AUTH] ERRO: Token não fornecido para %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token de autorização necessário"})
			c.Abort()
			return
		}

		// Verificar formato do token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Printf("[AUTH] ERRO: Formato de token inválido para %s: %v", c.Request.URL.Path, tokenParts)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validar token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			log.Printf("[AUTH] ERRO: Token inválido para %s: %v", c.Request.URL.Path, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		// Adicionar informações do usuário ao contexto
		c.Set("user_id", claims.UserID)
		c.Set("userID", claims.UserID)  // Para compatibilidade
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		log.Printf("[AUTH] SUCESSO: %s autenticado para %s (UserID: %s)", claims.Email, c.Request.URL.Path, claims.UserID)
		c.Next()
	}
}

// AdminMiddleware verifica se o usuário é admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			c.Abort()
			return
		}

		if userRole != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Apenas administradores."})
			c.Abort()
			return
		}

		c.Next()
	}
}
