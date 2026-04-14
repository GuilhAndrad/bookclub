package repository

import (
	"context"
	"fmt"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
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
func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("comment_repository.Create: %w", err)
	}
	return nil
}

// FindByReviewID retorna uma página de comentários de uma resenha em ordem cronológica.
// O preload de Author carrega apenas id e name — sem campos sensíveis do usuário.
func (r *CommentRepository) FindByReviewID(ctx context.Context, reviewID uuid.UUID, p pagination.Params) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Comment{}).Where("review_id = ?", reviewID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("comment_repository.FindByReviewID count: %w", err)
	}

	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("review_id = ?", reviewID).
		Order("created_at ASC").
		Limit(p.Limit).
		Offset(p.Offset()).
		Find(&comments).Error
	if err != nil {
		return nil, 0, fmt.Errorf("comment_repository.FindByReviewID: %w", err)
	}

	return comments, total, nil
}
