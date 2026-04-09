package handler

import (
	"net/http"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// reviewService define as operações necessárias para ReviewHandler.
type reviewService interface {
	ListByBook(bookID, requesterID uuid.UUID) ([]domain.Review, error)
	ListByUser(userID uuid.UUID) ([]domain.Review, error)
	Create(input service.CreateReviewInput) (*domain.Review, error)
	Update(id, userID uuid.UUID, input service.UpdateReviewInput) error
	Delete(id, userID uuid.UUID, isAdmin bool) error
	Like(reviewID, userID uuid.UUID) error
	Unlike(reviewID, userID uuid.UUID) error
	ListComments(reviewID uuid.UUID) ([]domain.Comment, error)
	CreateComment(reviewID, userID uuid.UUID, content string) (*domain.Comment, error)
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
// GET /books/:id/reviews
// Retorna as resenhas de um livro com contagens de likes e comentários.
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	reviews, err := h.svc.ListByBook(bookID, middleware.GetUserID(c))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

// ListUserReviews godoc
// GET /users/:id/reviews
// Retorna todas as resenhas escritas por um usuário.
func (h *ReviewHandler) ListUserReviews(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	reviews, err := h.svc.ListByUser(userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

// CreateReview godoc
// POST /books/:id/reviews
// Cria uma resenha para o livro. Cada usuário pode ter apenas uma resenha por livro.
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

	review, err := h.svc.Create(service.CreateReviewInput{
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
// Atualiza uma resenha do usuário autenticado. Campos omitidos não são alterados.
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

	if err := h.svc.Update(reviewID, middleware.GetUserID(c), service.UpdateReviewInput{
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
// Remove uma resenha. Administradores podem remover qualquer resenha.
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.svc.Delete(
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
// Registra a curtida do usuário autenticado em uma resenha.
func (h *ReviewHandler) LikeReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.svc.Like(reviewID, middleware.GetUserID(c)); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "curtida registrada"})
}

// UnlikeReview godoc
// DELETE /reviews/:id/like
// Remove a curtida do usuário autenticado de uma resenha.
func (h *ReviewHandler) UnlikeReview(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	h.svc.Unlike(reviewID, middleware.GetUserID(c)) //nolint:errcheck
	c.JSON(http.StatusOK, gin.H{"message": "curtida removida"})
}

// ListComments godoc
// GET /reviews/:id/comments
// Retorna os comentários de uma resenha em ordem cronológica.
func (h *ReviewHandler) ListComments(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	comments, err := h.svc.ListComments(reviewID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// CreateComment godoc
// POST /reviews/:id/comments
// Adiciona um comentário a uma resenha.
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

	comment, err := h.svc.CreateComment(reviewID, middleware.GetUserID(c), input.Content)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}
