package handler

import (
	"context"
	"net/http"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// userService define as operações necessárias para UserHandler.
type userService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*service.ProfileOutput, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, input service.UpdateProfileInput) error
	GetPendingMembers(ctx context.Context, p pagination.Params) (pagination.Page[domain.User], error)
	ApproveMember(ctx context.Context, id uuid.UUID) error
	RejectMember(ctx context.Context, id uuid.UUID) error
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
func (h *UserHandler) GetProfile(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	out, err := h.svc.GetProfile(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// UpdateProfile godoc
// PUT /users/me/profile
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
	if err := h.svc.UpdateProfile(c.Request.Context(), middleware.GetUserID(c), service.UpdateProfileInput{
		Name:          input.Name,
		Bio:           input.Bio,
		FavoriteGenre: input.FavoriteGenre,
		AvatarURL:     input.AvatarURL,
	}); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "perfil atualizado"})
}

// GetPendingMembers godoc
// GET /admin/members/pending?page=1&limit=20
func (h *UserHandler) GetPendingMembers(c *gin.Context) {
	p, ok := pagination.FromRequest(c)
	if !ok {
		return
	}
	page, err := h.svc.GetPendingMembers(c.Request.Context(), p)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// ApproveMember godoc
// PUT /admin/members/:id/approve
func (h *UserHandler) ApproveMember(c *gin.Context) {
	h.changeMemberStatus(c, true)
}

// RejectMember godoc
// PUT /admin/members/:id/reject
func (h *UserHandler) RejectMember(c *gin.Context) {
	h.changeMemberStatus(c, false)
}

func (h *UserHandler) changeMemberStatus(c *gin.Context, approve bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if approve {
		err = h.svc.ApproveMember(c.Request.Context(), id)
	} else {
		err = h.svc.RejectMember(c.Request.Context(), id)
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status atualizado"})
}
