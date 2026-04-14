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

// bookService define as operações necessárias para BookHandler.
type bookService interface {
	ListBooks(ctx context.Context, p pagination.Params) (pagination.Page[domain.Book], error)
	GetBook(ctx context.Context, id uuid.UUID) (*service.BookDetailOutput, error)
	GetCurrentBook(ctx context.Context) (*domain.Book, error)
	CreateBook(ctx context.Context, input service.CreateBookInput) (*domain.Book, error)
	UpdateBook(ctx context.Context, id uuid.UUID, input service.UpdateBookInput) error
}

// BookHandler lida com as requisições relacionadas a livros.
type BookHandler struct {
	svc bookService
}

// NewBookHandler cria um BookHandler com o service fornecido.
func NewBookHandler(svc bookService) *BookHandler {
	return &BookHandler{svc: svc}
}

// ListBooks godoc
// GET /books?page=1&limit=20
// Retorna uma página de livros cadastrados.
func (h *BookHandler) ListBooks(c *gin.Context) {
	p, ok := pagination.FromRequest(c)
	if !ok {
		return
	}

	page, err := h.svc.ListBooks(c.Request.Context(), p)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, page)
}

// GetCurrentBook godoc
// GET /books/current
// Retorna o livro do mês corrente.
func (h *BookHandler) GetCurrentBook(c *gin.Context) {
	book, err := h.svc.GetCurrentBook(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"book": book})
}

// GetBook godoc
// GET /books/:id
// Retorna um livro com suas estatísticas de avaliação.
func (h *BookHandler) GetBook(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	out, err := h.svc.GetBook(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

// CreateBook godoc
// POST /admin/books
// Cadastra o livro do mês. Restrito a administradores.
func (h *BookHandler) CreateBook(c *gin.Context) {
	var input struct {
		Title       string `json:"title"  binding:"required"`
		Author      string `json:"author" binding:"required"`
		Description string `json:"description"`
		CoverURL    string `json:"cover_url"`
		Month       int    `json:"month" binding:"required,min=1,max=12"`
		Year        int    `json:"year"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := h.svc.CreateBook(c.Request.Context(), service.CreateBookInput{
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		CoverURL:    input.CoverURL,
		Month:       input.Month,
		Year:        input.Year,
		CreatedByID: middleware.GetUserID(c),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"book": book})
}

// UpdateBook godoc
// PUT /admin/books/:id
// Atualiza os dados de um livro. Campos omitidos não são alterados. Restrito a administradores.
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var input struct {
		Title       string `json:"title"`
		Author      string `json:"author"`
		Description string `json:"description"`
		CoverURL    string `json:"cover_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateBook(c.Request.Context(), id, service.UpdateBookInput{
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		CoverURL:    input.CoverURL,
	}); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "livro atualizado"})
}
