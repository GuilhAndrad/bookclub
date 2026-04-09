package handler

import (
	"net/http"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
)

// authService define as operações de autenticação necessárias para AuthHandler.
type authService interface {
	Register(input service.RegisterInput) (*domain.User, error)
	Login(email, password string) (*service.LoginOutput, error)
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
// Cria um novo usuário com status pendente de aprovação.
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

	user, err := h.svc.Register(service.RegisterInput{
		Name:     input.Name,
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
// Autentica um usuário aprovado e retorna um token JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.svc.Login(input.Email, input.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": out.Token,
		"user":  out.User,
	})
}