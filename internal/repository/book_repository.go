package repository

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
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

// Create persiste um novo livro no banco.
func (r *BookRepository) Create(book *domain.Book) error {
	return r.db.Create(book).Error
}

// FindAll retorna todos os livros ordenados do mais recente para o mais antigo.
func (r *BookRepository) FindAll() ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.Order("year DESC, month DESC").Find(&books).Error
	return books, err
}

// FindByID busca um livro pelo ID. Retorna gorm.ErrRecordNotFound se não existir.
func (r *BookRepository) FindByID(id uuid.UUID) (*domain.Book, error) {
	var book domain.Book
	err := r.db.First(&book, "id = ?", id).Error
	return &book, err
}

// FindByMonthYear busca o livro de um mês e ano específicos.
// Retorna gorm.ErrRecordNotFound se não houver livro cadastrado para o período.
func (r *BookRepository) FindByMonthYear(month, year int) (*domain.Book, error) {
	var book domain.Book
	err := r.db.Where("month = ? AND year = ?", month, year).First(&book).Error
	return &book, err
}

// Update aplica atualizações parciais a um livro.
func (r *BookRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.Book{}).Where("id = ?", id).Updates(updates).Error
}

// CountReviews retorna o total de resenhas do livro.
func (r *BookRepository) CountReviews(bookID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Review{}).Where("book_id = ?", bookID).Count(&count).Error
	return count, err
}

// AvgRating retorna a média das avaliações do livro. Retorna 0 se não houver resenhas.
func (r *BookRepository) AvgRating(bookID uuid.UUID) (float64, error) {
	var avg float64
	err := r.db.Model(&domain.Review{}).
		Where("book_id = ?", bookID).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&avg).Error
	return avg, err
}
