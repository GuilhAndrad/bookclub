package handler

import (
	"net/http"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userService define as operações necessárias para UserHandler.
type userService interface {
	GetProfile(id uuid.UUID) (*service.ProfileOutput, error)
	UpdateProfile(id uuid.UUID, input service.UpdateProfileInput) error
	GetPendingMembers() ([]domain.User, error)
	ApproveMember(id uuid.UUID) error
	RejectMember(id uuid.UUID) error
}

// UserHandler lida com as requisições relacionadas a perfis e aprovação de membros.
type UserHandler struct {
	svc userService
}

// NewUserHandler cria um UserHandler com o service fornecido.
func NewUserHandler(svc userService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile godoc
// GET /users/:id/profile
// Retorna o perfil público de um usuário aprovado com suas estatísticas.
func (h *UserHandler) GetProfile(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	out, err := h.svc.GetProfile(id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

// UpdateProfile godoc
// PUT /users/me/profile
// Atualiza o perfil do usuário autenticado. Campos omitidos não são alterados.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var input struct {
		Name          string `json:"name"`
		Bio           string `json:"bio"`
		FavoriteGenre string `json:"favorite_genre"`
		AvatarURL     string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.UpdateProfile(middleware.GetUserID(c), service.UpdateProfileInput{
		Name:          input.Name,
		Bio:           input.Bio,
		FavoriteGenre: input.FavoriteGenre,
		AvatarURL:     input.AvatarURL,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "perfil atualizado"})
}

// GetPendingMembers godoc
// GET /admin/members/pending
// Retorna todos os usuários com cadastro aguardando aprovação.
func (h *UserHandler) GetPendingMembers(c *gin.Context) {
	members, err := h.svc.GetPendingMembers()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

// ApproveMember godoc
// PUT /admin/members/:id/approve
// Aprova o cadastro de um usuário pendente.
func (h *UserHandler) ApproveMember(c *gin.Context) {
	h.changeMemberStatus(c, true)
}

// RejectMember godoc
// PUT /admin/members/:id/reject
// Recusa o cadastro de um usuário pendente.
func (h *UserHandler) RejectMember(c *gin.Context) {
	h.changeMemberStatus(c, false)
}

// changeMemberStatus centraliza a lógica de aprovação e rejeição.
func (h *UserHandler) changeMemberStatus(c *gin.Context, approve bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if approve {
		err = h.svc.ApproveMember(id)
	} else {
		err = h.svc.RejectMember(id)
	}

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status atualizado"})
}
