package service

import (
	"errors"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepo define as operações de repositório necessárias para UserService.
type userRepo interface {
	FindByID(id uuid.UUID) (*domain.User, error)
	UpdateProfile(id uuid.UUID, updates map[string]interface{}) error
	UpdateStatus(id uuid.UUID, status domain.UserStatus) error
	FindPending() ([]domain.User, error)
	CountReviews(id uuid.UUID) (int64, error)
	AvgRating(id uuid.UUID) (float64, error)
}

// UserService gerencia operações de perfil e aprovação de membros.
type UserService struct {
	users userRepo
}

// NewUserService cria um UserService com o repositório fornecido.
func NewUserService(users userRepo) *UserService {
	return &UserService{users: users}
}

// ProfileOutput agrega o perfil do usuário com suas estatísticas.
type ProfileOutput struct {
	User        *domain.User `json:"user"`
	ReviewCount int64        `json:"review_count"`
	AvgRating   float64      `json:"avg_rating"`
}

// UpdateProfileInput contém os campos atualizáveis do perfil.
type UpdateProfileInput struct {
	Name          string
	Bio           string
	FavoriteGenre string
	AvatarURL     string
}

// GetProfile retorna o perfil de um usuário com suas estatísticas.
// Retorna apperr.ErrNotFound se o usuário não existir ou não estiver aprovado.
func (s *UserService) GetProfile(id uuid.UUID) (*ProfileOutput, error) {
	user, err := s.users.FindByID(id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	count, _ := s.users.CountReviews(id)
	avg, _ := s.users.AvgRating(id)

	return &ProfileOutput{User: user, ReviewCount: count, AvgRating: avg}, nil
}

// UpdateProfile aplica atualizações parciais ao perfil. Campos vazios são ignorados.
func (s *UserService) UpdateProfile(id uuid.UUID, input UpdateProfileInput) error {
	updates := make(map[string]interface{})
	if input.Name != ""          { updates["name"] = input.Name }
	if input.Bio != ""           { updates["bio"] = input.Bio }
	if input.FavoriteGenre != "" { updates["favorite_genre"] = input.FavoriteGenre }
	if input.AvatarURL != ""     { updates["avatar_url"] = input.AvatarURL }
	return s.users.UpdateProfile(id, updates)
}

// GetPendingMembers retorna todos os usuários com cadastro pendente.
func (s *UserService) GetPendingMembers() ([]domain.User, error) {
	return s.users.FindPending()
}

// ApproveMember aprova o cadastro de um usuário pendente.
// Retorna apperr.ErrNotFound se o usuário não existir.
func (s *UserService) ApproveMember(id uuid.UUID) error {
	return s.changeMemberStatus(id, domain.StatusApproved)
}

// RejectMember recusa o cadastro de um usuário pendente.
// Retorna apperr.ErrNotFound se o usuário não existir.
func (s *UserService) RejectMember(id uuid.UUID) error {
	return s.changeMemberStatus(id, domain.StatusRejected)
}

func (s *UserService) changeMemberStatus(id uuid.UUID, status domain.UserStatus) error {
	err := s.users.UpdateStatus(id, status)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.ErrNotFound
	}
	return err
}
