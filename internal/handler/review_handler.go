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

// reviewService define as operações necessárias para ReviewHandler.
type reviewService interface {
	ListByBook(ctx context.Context, bookID, requesterID uuid.UUID, p pagination.Params) (pagination.Page[domain.Review], error)
	ListByUser(ctx context.Context, userID uuid.UUID, p pagination.Params) (pagination.Page[domain.Review], error)
	Create(ctx context.Context, input service.CreateReviewInput) (*domain.Review, error)
	Update(ctx context.Context, id, userID uuid.UUID, input service.UpdateReviewInput) error
	Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error
	Like(ctx context.Context, reviewID, userID uuid.UUID) error
	Unlike(ctx context.Context, reviewID, userID uuid.UUID) error
	ListComments(ctx context.Context, reviewID uuid.UUID, p pagination.Params) (pagination.Page[domain.Comment], error)
	CreateComment(ctx context.Context, reviewID, userID uuid.UUID, content string) (*domain.Comment, error)
}

// ReviewHandler lida com as requisições de resenhas, likes e comentários.
type ReviewHandler struct {
	svc reviewService
}

// NewReviewHandler cria um ReviewHandler com o service fornecido.
func NewReviewHandler(svc reviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// ListReviews godoc
// GET /books/:id/reviews?page=1&limit=20
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	p, ok := pagination.FromRequest(c)
	if !ok {
		return
	}

	page, err := h.svc.ListByBook(c.Request.Context(), bookID, middleware.GetUserID(c), p)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, page)
}

// ListUserReviews godoc
// GET /users/:id/reviews?page=1&limit=20
func (h *ReviewHandler) ListUserReviews(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	p, ok := pagination.FromRequest(c)
	if !ok {
		return
	}

	page, err := h.svc.ListByUser(c.Request.Context(), userID, p)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, page)
}

// CreateReview godoc
// POST /books/:id/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required,min=10"`
		Rating  int    `json:"rating"  binding:"required,min=1,max=5"`
		Spoiler bool   `json:"spoiler"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.svc.Create(c.Request.Context(), service.CreateReviewInput{
		BookID:  bookID,
		UserID:  middleware.GetUserID(c),
		Content: input.Content,
		Rating:  input.Rating,
		Spoiler: input.Spoiler,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"review": review})
}

// UpdateReview godoc
// PUT /reviews/:id
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var input struct {
		Content string `json:"content"`
		Rating  int    `json:"rating" binding:"omitempty,min=1,max=5"`
		Spoiler *bool  `json:"spoiler"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Update(c.Request.Context(), reviewID, middleware.GetUserID(c), service.UpdateReviewInput{
		Content: input.Content,
		Rating:  input.Rating,
		Spoiler: input.Spoiler,
	}); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resenha atualizada"})
}

// DeleteReview godoc
// DELETE /reviews/:id
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.svc.Delete(
		c.Request.Context(),
		reviewID,
		middleware.GetUserID(c),
		middleware.GetUserRole(c) == "admin",
	); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resenha removida"})
}

// LikeReview godoc
// POST /reviews/:id/like
func (h *ReviewHandler) LikeReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.svc.Like(c.Request.Context(), reviewID, middleware.GetUserID(c)); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "curtida registrada"})
}

// UnlikeReview godoc
// DELETE /reviews/:id/like
func (h *ReviewHandler) UnlikeReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.svc.Unlike(c.Request.Context(), reviewID, middleware.GetUserID(c)); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "curtida removida"})
}

// ListComments godoc
// GET /reviews/:id/comments?page=1&limit=20
func (h *ReviewHandler) ListComments(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	p, ok := pagination.FromRequest(c)
	if !ok {
		return
	}

	page, err := h.svc.ListComments(c.Request.Context(), reviewID, p)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, page)
}

// CreateComment godoc
// POST /reviews/:id/comments
func (h *ReviewHandler) CreateComment(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.svc.CreateComment(c.Request.Context(), reviewID, middleware.GetUserID(c), input.Content)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}
