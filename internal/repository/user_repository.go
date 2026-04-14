package repository

import (
	"context"
	"fmt"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository gerencia a persistência de usuários no banco de dados.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository cria um UserRepository com a conexão fornecida.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// userStats agrega estatísticas de resenhas de um usuário.
type userStats struct {
	ReviewCount int64
	AvgRating   float64
}

// Create persiste um novo usuário. Retorna erro se o e-mail já estiver em uso.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("user_repository.Create: %w", err)
	}
	return nil
}

// FindByEmail busca um usuário pelo e-mail. Retorna gorm.ErrRecordNotFound se não existir.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user_repository.FindByEmail: %w", err)
	}
	return &user, nil
}

// FindByID busca um usuário aprovado pelo ID. Retorna gorm.ErrRecordNotFound se não existir.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, domain.StatusApproved).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user_repository.FindByID: %w", err)
	}
	return &user, nil
}

// UpdateProfile aplica atualizações parciais ao perfil do usuário.
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("user_repository.UpdateProfile: %w", err)
	}
	return nil
}

// UpdateStatus altera o status de aprovação do usuário.
// Retorna gorm.ErrRecordNotFound se nenhuma linha for afetada.
func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("user_repository.UpdateStatus: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindPending retorna uma página de usuários com cadastro pendente de aprovação.
func (r *UserRepository) FindPending(ctx context.Context, p pagination.Params) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("status = ?", domain.StatusPending).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("user_repository.FindPending count: %w", err)
	}

	err := r.db.WithContext(ctx).
		Select("id, name, email, created_at").
		Where("status = ?", domain.StatusPending).
		Order("created_at ASC").
		Limit(p.Limit).
		Offset(p.Offset()).
		Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("user_repository.FindPending: %w", err)
	}

	return users, total, nil
}

// GetStats retorna a contagem de resenhas e a média de avaliações de um usuário em uma única query.
func (r *UserRepository) GetStats(ctx context.Context, userID uuid.UUID) (int64, float64, error) {
	var stats userStats
	err := r.db.WithContext(ctx).
		Model(&domain.Review{}).
		Select("COUNT(*) as review_count, COALESCE(AVG(rating), 0) as avg_rating").
		Where("user_id = ?", userID).
		Scan(&stats).Error
	if err != nil {
		return 0, 0, fmt.Errorf("user_repository.GetStats: %w", err)
	}
	return stats.ReviewCount, stats.AvgRating, nil
}
