package middleware

import (
	"net/http"
	"strings"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// chaves de contexto não exportadas evitam colisões com outros pacotes.
type contextKey string

const (
	keyUserID contextKey = "userID"
	keyRole   contextKey = "userRole"
)

// tokenValidator define a operação mínima que o middleware precisa do AuthService.
// Seguindo "aceite interfaces" — o middleware não depende do concreto *service.AuthService.
type tokenValidator interface {
	ValidateToken(tokenString string) (*service.Claims, error)
}

// AuthMiddleware agrupa os middlewares de autenticação e autorização.
type AuthMiddleware struct {
	validator tokenValidator
}

// NewAuthMiddleware cria um AuthMiddleware com o validator fornecido.
func NewAuthMiddleware(validator tokenValidator) *AuthMiddleware {
	return &AuthMiddleware{validator: validator}
}

// AuthRequired valida o token JWT e injeta userID e role no contexto Gin.
// Deve ser usado antes de qualquer handler que exija autenticação.
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token não fornecido"})
			return
		}

		claims, err := m.validator.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			return
		}

		c.Set(string(keyUserID), claims.UserID)
		c.Set(string(keyRole), claims.Role)
		c.Next()
	}
}

// AdminRequired garante que o usuário autenticado seja um administrador.
// Deve ser usado após AuthRequired.
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetUserRole(c) != string(domain.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acesso restrito a administradores"})
			return
		}
		c.Next()
	}
}

// GetUserID extrai o UUID do usuário autenticado do contexto Gin.
// Retorna uuid.Nil se a chave não existir ou o tipo for inesperado.
func GetUserID(c *gin.Context) uuid.UUID {
	v, ok := c.Get(string(keyUserID))
	if !ok {
		return uuid.Nil
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

// GetUserRole extrai a role do usuário autenticado do contexto Gin.
// Retorna string vazia se a chave não existir ou o tipo for inesperado.
func GetUserRole(c *gin.Context) string {
	v, ok := c.Get(string(keyRole))
	if !ok {
		return ""
	}
	role, ok := v.(string)
	if !ok {
		return ""
	}
	return role
}
