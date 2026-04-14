package repository

import (
	"context"
	"fmt"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookRepository gerencia a persistência de livros no banco de dados.
type BookRepository struct {
	db *gorm.DB
}

// NewBookRepository cria um BookRepository com a conexão fornecida.
func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db: db}
}

// bookStats agrega contagem e média de avaliações de um livro.
type bookStats struct {
	ReviewCount int64
	AvgRating   float64
}

// Create persiste um novo livro no banco.
func (r *BookRepository) Create(ctx context.Context, book *domain.Book) error {
	if err := r.db.WithContext(ctx).Create(book).Error; err != nil {
		return fmt.Errorf("book_repository.Create: %w", err)
	}
	return nil
}

// FindAll retorna uma página de livros ordenados do mais recente para o mais antigo.
func (r *BookRepository) FindAll(ctx context.Context, p pagination.Params) ([]domain.Book, int64, error) {
	var books []domain.Book
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Book{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("book_repository.FindAll count: %w", err)
	}

	err := r.db.WithContext(ctx).
		Order("year DESC, month DESC").
		Limit(p.Limit).
		Offset(p.Offset()).
		Find(&books).Error
	if err != nil {
		return nil, 0, fmt.Errorf("book_repository.FindAll: %w", err)
	}

	return books, total, nil
}

// FindByID busca um livro pelo ID. Retorna gorm.ErrRecordNotFound se não existir.
func (r *BookRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	var book domain.Book
	if err := r.db.WithContext(ctx).First(&book, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("book_repository.FindByID: %w", err)
	}
	return &book, nil
}

// FindByMonthYear busca o livro de um mês e ano específicos.
// Beneficia-se do índice composto idx_books_month_year definido no domain.
func (r *BookRepository) FindByMonthYear(ctx context.Context, month, year int) (*domain.Book, error) {
	var book domain.Book
	if err := r.db.WithContext(ctx).Where("month = ? AND year = ?", month, year).First(&book).Error; err != nil {
		return nil, fmt.Errorf("book_repository.FindByMonthYear: %w", err)
	}
	return &book, nil
}

// Update aplica atualizações parciais a um livro.
func (r *BookRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&domain.Book{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("book_repository.Update: %w", err)
	}
	return nil
}

// GetStats retorna a contagem de resenhas e a média de avaliações de um livro em uma única query.
func (r *BookRepository) GetStats(ctx context.Context, bookID uuid.UUID) (int64, float64, error) {
	var stats bookStats
	err := r.db.WithContext(ctx).
		Model(&domain.Review{}).
		Select("COUNT(*) as review_count, COALESCE(AVG(rating), 0) as avg_rating").
		Where("book_id = ?", bookID).
		Scan(&stats).Error
	if err != nil {
		return 0, 0, fmt.Errorf("book_repository.GetStats: %w", err)
	}
	return stats.ReviewCount, stats.AvgRating, nil
}
