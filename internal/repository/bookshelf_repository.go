package repository

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookshelfRepository gerencia a estante virtual dos usuários no banco de dados.
type BookshelfRepository struct {
	db *gorm.DB
}

// NewBookshelfRepository cria um BookshelfRepository com a conexão fornecida.
func NewBookshelfRepository(db *gorm.DB) *BookshelfRepository {
	return &BookshelfRepository{db: db}
}

// Upsert cria ou atualiza a entrada de um livro na estante do usuário.
func (r *BookshelfRepository) Upsert(entry *domain.UserBook) error {
	return r.db.Save(entry).Error
}

// FindByUserID retorna todos os livros da estante de um usuário com detalhes do livro.
func (r *BookshelfRepository) FindByUserID(userID uuid.UUID) ([]domain.UserBook, error) {
	var shelf []domain.UserBook
	err := r.db.Preload("Book").Where("user_id = ?", userID).Find(&shelf).Error
	return shelf, err
}