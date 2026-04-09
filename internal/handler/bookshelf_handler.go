package handler

import (
	"net/http"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/internal/middleware"
	"github.com/GuilhAndrad/bookclub/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bookshelfService define as operações necessárias para BookshelfHandler.
type bookshelfService interface {
	GetShelf(userID uuid.UUID) ([]domain.UserBook, error)
	Upsert(input service.UpsertBookshelfInput) error
}

// BookshelfHandler lida com as requisições da estante virtual do usuário.
type BookshelfHandler struct {
	svc bookshelfService
}

// NewBookshelfHandler cria um BookshelfHandler com o service fornecido.
func NewBookshelfHandler(svc bookshelfService) *BookshelfHandler {
	return &BookshelfHandler{svc: svc}
}

// GetShelf godoc
// GET /users/me/bookshelf
// Retorna todos os livros da estante do usuário autenticado.
func (h *BookshelfHandler) GetShelf(c *gin.Context) {
	shelf, err := h.svc.GetShelf(middleware.GetUserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"bookshelf": shelf})
}

// UpsertEntry godoc
// PUT /users/me/bookshelf/:bookId
// Adiciona ou atualiza o status de um livro na estante do usuário autenticado.
func (h *BookshelfHandler) UpsertEntry(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("bookId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do livro inválido"})
		return
	}

	var input struct {
		Status        domain.BookshelfStatus `json:"status" binding:"required"`
		PersonalNotes string                 `json:"personal_notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Upsert(service.UpsertBookshelfInput{
		UserID:        middleware.GetUserID(c),
		BookID:        bookID,
		Status:        input.Status,
		PersonalNotes: input.PersonalNotes,
	}); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "estante atualizada"})
}
