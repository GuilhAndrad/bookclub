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

// AuthMiddleware agrupa middlewares que dependem do AuthService.
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware cria um novo conjunto de middlewares de autenticação.
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// AuthRequired valida o token JWT e injeta userID e role no contexto.
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token não fornecido"})
			return
		}

		claims, err := m.authService.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			return
		}

		c.Set(string(keyUserID), claims.UserID)
		c.Set(string(keyRole), claims.Role)
		c.Next()
	}
}

// AdminRequired garante que o usuário seja administrador.
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetUserRole(c) != string(domain.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acesso restrito a administradores"})
			return
		}
		c.Next()
	}
}

// GetUserID extrai o UUID do usuário autenticado.
func GetUserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(string(keyUserID))
	return v.(uuid.UUID)
}

// GetUserRole extrai a role do usuário autenticado.
func GetUserRole(c *gin.Context) string {
	v, _ := c.Get(string(keyRole))
	return v.(string)
}