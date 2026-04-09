package repository

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
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

// Create persiste uma nova resenha no banco.
func (r *ReviewRepository) Create(review *domain.Review) error {
	return r.db.Create(review).Error
}

// FindByBookID retorna todas as resenhas de um livro, com o autor pré-carregado.
func (r *ReviewRepository) FindByBookID(bookID uuid.UUID) ([]domain.Review, error) {
	var reviews []domain.Review
	err := r.db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, avatar_url")
		}).
		Where("book_id = ?", bookID).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

// FindByUserID retorna todas as resenhas de um usuário, com o livro pré-carregado.
func (r *ReviewRepository) FindByUserID(userID uuid.UUID) ([]domain.Review, error) {
	var reviews []domain.Review
	err := r.db.
		Preload("Book").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

// FindByID busca uma resenha pelo ID. Retorna gorm.ErrRecordNotFound se não existir.
func (r *ReviewRepository) FindByID(id uuid.UUID) (*domain.Review, error) {
	var review domain.Review
	err := r.db.First(&review, "id = ?", id).Error
	return &review, err
}

// ExistsByBookAndUser verifica se já existe uma resenha do usuário para o livro.
func (r *ReviewRepository) ExistsByBookAndUser(bookID, userID uuid.UUID) bool {
	var count int64
	r.db.Model(&domain.Review{}).Where("book_id = ? AND user_id = ?", bookID, userID).Count(&count)
	return count > 0
}

// Update aplica atualizações parciais a uma resenha.
func (r *ReviewRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.Review{}).Where("id = ?", id).Updates(updates).Error
}

// Delete remove uma resenha. Quando isAdmin é false, restringe a deleção ao dono.
// Retorna (false, nil) quando nenhuma linha foi afetada.
func (r *ReviewRepository) Delete(id, userID uuid.UUID, isAdmin bool) (bool, error) {
	query := r.db.Where("id = ?", id)
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}
	result := query.Delete(&domain.Review{})
	return result.RowsAffected > 0, result.Error
}

// CountLikes retorna o número de curtidas de uma resenha.
func (r *ReviewRepository) CountLikes(reviewID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Like{}).Where("review_id = ?", reviewID).Count(&count).Error
	return count, err
}

// CountComments retorna o número de comentários de uma resenha.
func (r *ReviewRepository) CountComments(reviewID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Comment{}).Where("review_id = ?", reviewID).Count(&count).Error
	return count, err
}

// IsLikedBy verifica se um usuário curtiu a resenha.
func (r *ReviewRepository) IsLikedBy(reviewID, userID uuid.UUID) bool {
	var count int64
	r.db.Model(&domain.Like{}).Where("review_id = ? AND user_id = ?", reviewID, userID).Count(&count)
	return count > 0
}

// Like registra a curtida de um usuário em uma resenha.
// Retorna erro se o usuário já tiver curtido.
func (r *ReviewRepository) Like(reviewID, userID uuid.UUID) error {
	return r.db.Create(&domain.Like{ReviewID: reviewID, UserID: userID}).Error
}

// Unlike remove a curtida de um usuário de uma resenha.
func (r *ReviewRepository) Unlike(reviewID, userID uuid.UUID) error {
	return r.db.Where("review_id = ? AND user_id = ?", reviewID, userID).Delete(&domain.Like{}).Error
}