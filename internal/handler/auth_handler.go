package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
)

// authService define as operações de autenticação necessárias para AuthHandler.
type authService interface {
	Register(ctx context.Context, input service.RegisterInput) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*service.LoginOutput, error)
	RefreshToken(userID string, role string) (string, error)
	Logout(token string) error
	ValidateToken(tokenString string) (*service.Claims, error)
}

// AuthHandler lida com as requisições de autenticação.
type AuthHandler struct {
	svc authService
}

// NewAuthHandler cria um AuthHandler com o service fornecido.
func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register godoc
// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name"     binding:"required,min=2"`
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	user, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Name:     strings.TrimSpace(input.Name),
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "cadastro realizado, aguarde a aprovação de um administrador",
		"user": gin.H{
			"id":     user.ID,
			"name":   user.Name,
			"email":  user.Email,
			"status": user.Status,
		},
	})
}

// Login godoc
// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.svc.Login(c.Request.Context(), strings.ToLower(strings.TrimSpace(input.Email)), input.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": out.Token,
		"user":  out.User,
	})
}

// Logout godoc
// POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if err := h.svc.Logout(token); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sessão encerrada"})
}

// RefreshToken godoc
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	newToken, err := h.svc.RefreshToken(userID.String(), role)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}
