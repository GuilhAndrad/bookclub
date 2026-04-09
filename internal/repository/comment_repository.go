package repository

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentRepository gerencia a persistência de comentários no banco de dados.
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository cria um CommentRepository com a conexão fornecida.
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create persiste um novo comentário no banco.
func (r *CommentRepository) Create(comment *domain.Comment) error {
	return r.db.Create(comment).Error
}

// FindByReviewID retorna todos os comentários de uma resenha em ordem cronológica.
func (r *CommentRepository) FindByReviewID(reviewID uuid.UUID) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, avatar_url")
		}).
		Where("review_id = ?", reviewID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}