package services

import (
	"errors"
	"time"
	"tappyone/internal/config"
	"tappyone/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required"`
}

type LoginResponse struct {
	Token   string         `json:"token"`
	Usuario models.Usuario `json:"usuario"`
}

func NewAuthService(db *gorm.DB, redis *redis.Client, config *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// Login autentica um usuário e retorna um JWT token
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var usuario models.Usuario
	
	// Buscar usuário pelo email
	if err := s.db.Where("email = ? AND ativo = ?", req.Email, true).First(&usuario).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("credenciais inválidas")
		}
		return nil, err
	}

	// Verificar senha
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.Senha), []byte(req.Senha)); err != nil {
		return nil, errors.New("credenciais inválidas")
	}

	// Gerar token JWT
	token, err := s.generateJWT(usuario)
	if err != nil {
		return nil, err
	}

	// Limpar senha antes de retornar
	usuario.Senha = ""

	return &LoginResponse{
		Token:   token,
		Usuario: usuario,
	}, nil
}

// generateJWT gera um token JWT para o usuário
func (s *AuthService) generateJWT(usuario models.Usuario) (string, error) {
	claims := JWTClaims{
		UserID: usuario.ID,
		Email:  usuario.Email,
		Role:   string(usuario.Tipo),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 dias
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tappyone-crm",
			Subject:   usuario.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken valida um token JWT
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

// HashPassword cria um hash da senha
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// GetUserByID busca um usuário pelo ID
func (s *AuthService) GetUserByID(userID string) (*models.Usuario, error) {
	var usuario models.Usuario
	if err := s.db.Where("id = ? AND ativo = ?", userID, true).First(&usuario).Error; err != nil {
		return nil, err
	}
	return &usuario, nil
}

// ResetPassword atualiza a senha do usuário
func (s *AuthService) ResetPassword(email, hashedPassword string) error {
	var usuario models.Usuario
	
	// Buscar usuário pelo email
	if err := s.db.Where("email = ? AND ativo = ?", email, true).First(&usuario).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("usuário não encontrado")
		}
		return err
	}

	// Atualizar senha
	usuario.Senha = hashedPassword
	if err := s.db.Save(&usuario).Error; err != nil {
		return err
	}

	return nil
}
