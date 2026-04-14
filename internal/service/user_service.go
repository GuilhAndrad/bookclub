package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepo define as operações de repositório necessárias para UserService.
type userRepo interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	FindPending(ctx context.Context, p pagination.Params) ([]domain.User, int64, error)
	GetStats(ctx context.Context, userID uuid.UUID) (count int64, avg float64, err error)
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
func (s *UserService) GetProfile(ctx context.Context, id uuid.UUID) (*ProfileOutput, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	count, avg, err := s.users.GetStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user_service.GetProfile stats: %w", err)
	}

	return &ProfileOutput{User: user, ReviewCount: count, AvgRating: avg}, nil
}

// UpdateProfile aplica atualizações parciais ao perfil. Campos vazios são ignorados.
func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, input UpdateProfileInput) error {
	updates := make(map[string]interface{}, 4)
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Bio != "" {
		updates["bio"] = input.Bio
	}
	if input.FavoriteGenre != "" {
		updates["favorite_genre"] = input.FavoriteGenre
	}
	if input.AvatarURL != "" {
		updates["avatar_url"] = input.AvatarURL
	}

	if err := s.users.UpdateProfile(ctx, id, updates); err != nil {
		return fmt.Errorf("user_service.UpdateProfile: %w", err)
	}
	return nil
}

// GetPendingMembers retorna uma página de usuários com cadastro pendente.
func (s *UserService) GetPendingMembers(ctx context.Context, p pagination.Params) (pagination.Page[domain.User], error) {
	users, total, err := s.users.FindPending(ctx, p)
	if err != nil {
		return pagination.Page[domain.User]{}, fmt.Errorf("user_service.GetPendingMembers: %w", err)
	}
	return pagination.New(users, total, p), nil
}

// ApproveMember aprova o cadastro de um usuário pendente.
// Retorna apperr.ErrNotFound se o usuário não existir.
func (s *UserService) ApproveMember(ctx context.Context, id uuid.UUID) error {
	return s.changeMemberStatus(ctx, id, domain.StatusApproved)
}

// RejectMember recusa o cadastro de um usuário pendente.
// Retorna apperr.ErrNotFound se o usuário não existir.
func (s *UserService) RejectMember(ctx context.Context, id uuid.UUID) error {
	return s.changeMemberStatus(ctx, id, domain.StatusRejected)
}

func (s *UserService) changeMemberStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	err := s.users.UpdateStatus(ctx, id, status)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("user_service.changeMemberStatus: %w", err)
	}
	return nil
}
