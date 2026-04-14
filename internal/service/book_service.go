package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// bookRepo define as operações de repositório necessárias para BookService.
type bookRepo interface {
	Create(ctx context.Context, book *domain.Book) error
	FindAll(ctx context.Context, p pagination.Params) ([]domain.Book, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	FindByMonthYear(ctx context.Context, month, year int) (*domain.Book, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	GetStats(ctx context.Context, bookID uuid.UUID) (count int64, avg float64, err error)
}

// BookService gerencia as operações sobre livros do clube.
type BookService struct {
	books bookRepo
}

// NewBookService cria um BookService com o repositório fornecido.
func NewBookService(books bookRepo) *BookService {
	return &BookService{books: books}
}

// BookDetailOutput agrega um livro com suas estatísticas de avaliação.
type BookDetailOutput struct {
	Book        *domain.Book `json:"book"`
	ReviewCount int64        `json:"review_count"`
	AvgRating   float64      `json:"avg_rating"`
}

// CreateBookInput contém os dados para cadastrar um novo livro do mês.
type CreateBookInput struct {
	Title       string
	Author      string
	Description string
	CoverURL    string
	Month       int
	Year        int
	CreatedByID uuid.UUID
}

// UpdateBookInput contém os campos atualizáveis de um livro.
type UpdateBookInput struct {
	Title       string
	Author      string
	Description string
	CoverURL    string
}

// ListBooks retorna uma página de livros cadastrados.
func (s *BookService) ListBooks(ctx context.Context, p pagination.Params) (pagination.Page[domain.Book], error) {
	books, total, err := s.books.FindAll(ctx, p)
	if err != nil {
		return pagination.Page[domain.Book]{}, fmt.Errorf("book_service.ListBooks: %w", err)
	}
	return pagination.New(books, total, p), nil
}

// GetBook retorna um livro com suas estatísticas de avaliação.
// Retorna apperr.ErrNotFound se o livro não existir.
func (s *BookService) GetBook(ctx context.Context, id uuid.UUID) (*BookDetailOutput, error) {
	book, err := s.books.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	count, avg, err := s.books.GetStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("book_service.GetBook stats: %w", err)
	}

	return &BookDetailOutput{Book: book, ReviewCount: count, AvgRating: avg}, nil
}

// GetCurrentBook retorna o livro do mês corrente.
// Retorna apperr.ErrNotFound se nenhum livro estiver cadastrado para o mês atual.
func (s *BookService) GetCurrentBook(ctx context.Context) (*domain.Book, error) {
	now := time.Now()
	book, err := s.books.FindByMonthYear(ctx, int(now.Month()), now.Year())
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return book, nil
}

// CreateBook cadastra um novo livro para o mês especificado.
// Retorna apperr.ErrConflict se já existir um livro para o mesmo mês e ano.
func (s *BookService) CreateBook(ctx context.Context, input CreateBookInput) (*domain.Book, error) {
	_, err := s.books.FindByMonthYear(ctx, input.Month, input.Year)
	if err == nil {
		return nil, apperr.ErrConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("book_service.CreateBook check: %w", err)
	}

	book := &domain.Book{
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		CoverURL:    input.CoverURL,
		Month:       input.Month,
		Year:        input.Year,
		CreatedByID: input.CreatedByID,
	}

	if err := s.books.Create(ctx, book); err != nil {
		return nil, fmt.Errorf("book_service.CreateBook: %w", err)
	}

	return book, nil
}

// UpdateBook atualiza os dados de um livro existente. Campos vazios são ignorados.
// Retorna apperr.ErrNotFound se o livro não existir.
func (s *BookService) UpdateBook(ctx context.Context, id uuid.UUID, input UpdateBookInput) error {
	if _, err := s.books.FindByID(ctx, id); err != nil {
		return apperr.ErrNotFound
	}

	updates := make(map[string]interface{}, 4)
	if input.Title != ""       { updates["title"] = input.Title }
	if input.Author != ""      { updates["author"] = input.Author }
	if input.Description != "" { updates["description"] = input.Description }
	if input.CoverURL != ""    { updates["cover_url"] = input.CoverURL }

	if err := s.books.Update(ctx, id, updates); err != nil {
		return fmt.Errorf("book_service.UpdateBook: %w", err)
	}
	return nil
}
