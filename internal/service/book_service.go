package service

import (
	"errors"
	"time"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// bookRepo define as operações de repositório necessárias para BookService.
type bookRepo interface {
	Create(book *domain.Book) error
	FindAll() ([]domain.Book, error)
	FindByID(id uuid.UUID) (*domain.Book, error)
	FindByMonthYear(month, year int) (*domain.Book, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	CountReviews(bookID uuid.UUID) (int64, error)
	AvgRating(bookID uuid.UUID) (float64, error)
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
	ISBN        string
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

// ListBooks retorna todos os livros cadastrados.
func (s *BookService) ListBooks() ([]domain.Book, error) {
	return s.books.FindAll()
}

// GetBook retorna um livro com suas estatísticas de avaliação.
// Retorna apperr.ErrNotFound se o livro não existir.
func (s *BookService) GetBook(id uuid.UUID) (*BookDetailOutput, error) {
	book, err := s.books.FindByID(id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	count, _ := s.books.CountReviews(id)
	avg, _ := s.books.AvgRating(id)

	return &BookDetailOutput{Book: book, ReviewCount: count, AvgRating: avg}, nil
}

// GetCurrentBook retorna o livro do mês corrente.
// Retorna apperr.ErrNotFound se nenhum livro estiver cadastrado para o mês atual.
func (s *BookService) GetCurrentBook() (*domain.Book, error) {
	now := time.Now()
	book, err := s.books.FindByMonthYear(int(now.Month()), now.Year())
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return book, nil
}

// CreateBook cadastra um novo livro para o mês especificado.
// Retorna apperr.ErrConflict se já existir um livro para o mesmo mês e ano.
func (s *BookService) CreateBook(input CreateBookInput) (*domain.Book, error) {
	_, err := s.books.FindByMonthYear(input.Month, input.Year)
	if err == nil {
		return nil, apperr.ErrConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	book := &domain.Book{
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		CoverURL:    input.CoverURL,
		ISBN:        input.ISBN,
		Month:       input.Month,
		Year:        input.Year,
		CreatedByID: input.CreatedByID,
	}

	if err := s.books.Create(book); err != nil {
		return nil, err
	}

	return book, nil
}

// UpdateBook atualiza os dados de um livro existente. Campos vazios são ignorados.
// Retorna apperr.ErrNotFound se o livro não existir.
func (s *BookService) UpdateBook(id uuid.UUID, input UpdateBookInput) error {
	if _, err := s.books.FindByID(id); err != nil {
		return apperr.ErrNotFound
	}

	updates := make(map[string]interface{})
	if input.Title != ""       { updates["title"] = input.Title }
	if input.Author != ""      { updates["author"] = input.Author }
	if input.Description != "" { updates["description"] = input.Description }
	if input.CoverURL != ""    { updates["cover_url"] = input.CoverURL }

	return s.books.Update(id, updates)
}
