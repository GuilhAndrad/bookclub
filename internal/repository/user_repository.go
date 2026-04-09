package repository

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
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

// Create persiste um novo usuário. Retorna erro se o e-mail já estiver em uso.
func (r *UserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// FindByEmail busca um usuário pelo e-mail. Retorna gorm.ErrRecordNotFound se não existir.
func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// FindByID busca um usuário aprovado pelo ID. Retorna gorm.ErrRecordNotFound se não existir.
func (r *UserRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ? AND status = ?", id, domain.StatusApproved).First(&user).Error
	return &user, err
}

// UpdateProfile aplica atualizações parciais ao perfil do usuário.
func (r *UserRepository) UpdateProfile(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateStatus altera o status de aprovação do usuário.
// Retorna gorm.ErrRecordNotFound se nenhuma linha for afetada.
func (r *UserRepository) UpdateStatus(id uuid.UUID, status domain.UserStatus) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", id).Update("status", status)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// FindPending retorna todos os usuários com cadastro pendente de aprovação.
func (r *UserRepository) FindPending() ([]domain.User, error) {
	var users []domain.User
	err := r.db.
		Select("id, name, email, created_at").
		Where("status = ?", domain.StatusPending).
		Order("created_at ASC").
		Find(&users).Error
	return users, err
}

// CountReviews retorna o total de resenhas escritas pelo usuário.
func (r *UserRepository) CountReviews(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Review{}).Where("user_id = ?", id).Count(&count).Error
	return count, err
}

// AvgRating retorna a média das avaliações dadas pelo usuário. Retorna 0 se não houver resenhas.
func (r *UserRepository) AvgRating(id uuid.UUID) (float64, error) {
	var avg float64
	err := r.db.Model(&domain.Review{}).
		Where("user_id = ?", id).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&avg).Error
	return avg, err
}