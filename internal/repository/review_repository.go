package repository

import (
	"context"
	"fmt"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewRepository gerencia a persistência de resenhas, likes e comentários.
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository cria um ReviewRepository com a conexão fornecida.
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// reviewStats agrupa as contagens de uma resenha retornadas em batch.
type reviewStats struct {
	ReviewID uuid.UUID
	Count    int64
}

// Create persiste uma nova resenha no banco.
func (r *ReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return fmt.Errorf("review_repository.Create: %w", err)
	}
	return nil
}

// FindByBookID retorna uma página de resenhas de um livro com o autor pré-carregado.
// O preload de Author carrega apenas id e name — sem campos sensíveis do usuário.
func (r *ReviewRepository) FindByBookID(ctx context.Context, bookID uuid.UUID, p pagination.Params) ([]domain.Review, int64, error) {
	var reviews []domain.Review
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Review{}).Where("book_id = ?", bookID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("review_repository.FindByBookID count: %w", err)
	}

	err := r.db.WithContext(ctx).
		Preload("Author").
		Where("book_id = ?", bookID).
		Order("created_at DESC").
		Limit(p.Limit).
		Offset(p.Offset()).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, fmt.Errorf("review_repository.FindByBookID: %w", err)
	}

	return reviews, total, nil
}

// FindByUserID retorna uma página de resenhas de um usuário com o livro pré-carregado.
func (r *ReviewRepository) FindByUserID(ctx context.Context, userID uuid.UUID, p pagination.Params) ([]domain.Review, int64, error) {
	var reviews []domain.Review
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Review{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("review_repository.FindByUserID count: %w", err)
	}

	err := r.db.WithContext(ctx).
		Preload("Book").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(p.Limit).
		Offset(p.Offset()).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, fmt.Errorf("review_repository.FindByUserID: %w", err)
	}

	return reviews, total, nil
}

// FindByID busca uma resenha pelo ID. Retorna gorm.ErrRecordNotFound se não existir.
func (r *ReviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	var review domain.Review
	if err := r.db.WithContext(ctx).First(&review, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("review_repository.FindByID: %w", err)
	}
	return &review, nil
}

// ExistsByBookAndUser verifica se já existe uma resenha do usuário para o livro.
// Usa SELECT EXISTS com LIMIT 1 — para ao encontrar a primeira linha.
func (r *ReviewRepository) ExistsByBookAndUser(ctx context.Context, bookID, userID uuid.UUID) bool {
	var exists bool
	r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM reviews WHERE book_id = ? AND user_id = ? AND deleted_at IS NULL LIMIT 1)",
		bookID, userID,
	).Scan(&exists)
	return exists
}

// Update aplica atualizações parciais a uma resenha.
func (r *ReviewRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&domain.Review{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("review_repository.Update: %w", err)
	}
	return nil
}

// Delete remove uma resenha. Quando isAdmin é false, restringe a deleção ao dono.
// Retorna (false, nil) quando nenhuma linha foi afetada.
func (r *ReviewRepository) Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) (bool, error) {
	query := r.db.WithContext(ctx).Where("id = ?", id)
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}
	result := query.Delete(&domain.Review{})
	if result.Error != nil {
		return false, fmt.Errorf("review_repository.Delete: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// BulkLikeCounts retorna o número de curtidas para um conjunto de resenhas em uma única query.
func (r *ReviewRepository) BulkLikeCounts(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	var rows []reviewStats
	err := r.db.WithContext(ctx).
		Model(&domain.Like{}).
		Select("review_id, COUNT(*) as count").
		Where("review_id IN ?", reviewIDs).
		Group("review_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("review_repository.BulkLikeCounts: %w", err)
	}

	result := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		result[row.ReviewID] = row.Count
	}
	return result, nil
}

// BulkCommentCounts retorna o número de comentários para um conjunto de resenhas em uma única query.
func (r *ReviewRepository) BulkCommentCounts(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	var rows []reviewStats
	err := r.db.WithContext(ctx).
		Model(&domain.Comment{}).
		Select("review_id, COUNT(*) as count").
		Where("review_id IN ?", reviewIDs).
		Group("review_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("review_repository.BulkCommentCounts: %w", err)
	}

	result := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		result[row.ReviewID] = row.Count
	}
	return result, nil
}

// BulkLikedByUser retorna o conjunto de IDs de resenhas curtidas pelo usuário em uma única query.
func (r *ReviewRepository) BulkLikedByUser(ctx context.Context, reviewIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	var likedIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&domain.Like{}).
		Select("review_id").
		Where("review_id IN ? AND user_id = ?", reviewIDs, userID).
		Scan(&likedIDs).Error
	if err != nil {
		return nil, fmt.Errorf("review_repository.BulkLikedByUser: %w", err)
	}

	result := make(map[uuid.UUID]bool, len(likedIDs))
	for _, id := range likedIDs {
		result[id] = true
	}
	return result, nil
}

// Like registra a curtida de um usuário em uma resenha.
func (r *ReviewRepository) Like(ctx context.Context, reviewID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Create(&domain.Like{ReviewID: reviewID, UserID: userID}).Error; err != nil {
		return fmt.Errorf("review_repository.Like: %w", err)
	}
	return nil
}

// Unlike remove a curtida de um usuário de uma resenha.
func (r *ReviewRepository) Unlike(ctx context.Context, reviewID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("review_id = ? AND user_id = ?", reviewID, userID).Delete(&domain.Like{}).Error; err != nil {
		return fmt.Errorf("review_repository.Unlike: %w", err)
	}
	return nil
}
