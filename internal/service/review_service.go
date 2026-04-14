package service

import (
	"context"
	"fmt"

	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/GuilhAndrad/bookclub/pkg/pagination"
	"github.com/google/uuid"
)

// reviewRepo define as operações de repositório necessárias para ReviewService.
type reviewRepo interface {
	Create(ctx context.Context, review *domain.Review) error
	FindByBookID(ctx context.Context, bookID uuid.UUID, p pagination.Params) ([]domain.Review, int64, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, p pagination.Params) ([]domain.Review, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	ExistsByBookAndUser(ctx context.Context, bookID, userID uuid.UUID) bool
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) (bool, error)
	BulkLikeCounts(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	BulkCommentCounts(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	BulkLikedByUser(ctx context.Context, reviewIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]bool, error)
	Like(ctx context.Context, reviewID, userID uuid.UUID) error
	Unlike(ctx context.Context, reviewID, userID uuid.UUID) error
}

// commentRepo define as operações de repositório necessárias para ReviewService.
type commentRepo interface {
	Create(ctx context.Context, comment *domain.Comment) error
	FindByReviewID(ctx context.Context, reviewID uuid.UUID, p pagination.Params) ([]domain.Comment, int64, error)
}

// ReviewService gerencia resenhas, likes e comentários.
type ReviewService struct {
	reviews  reviewRepo
	comments commentRepo
}

// NewReviewService cria um ReviewService com os repositórios fornecidos.
func NewReviewService(reviews reviewRepo, comments commentRepo) *ReviewService {
	return &ReviewService{reviews: reviews, comments: comments}
}

// CreateReviewInput contém os dados para criar uma nova resenha.
type CreateReviewInput struct {
	BookID  uuid.UUID
	UserID  uuid.UUID
	Content string
	Rating  int
	Spoiler bool
}

// UpdateReviewInput contém os campos atualizáveis de uma resenha.
type UpdateReviewInput struct {
	Content string
	Rating  int
	Spoiler *bool
}

// extractIDs extrai os UUIDs de uma lista de resenhas.
func extractIDs(reviews []domain.Review) []uuid.UUID {
	ids := make([]uuid.UUID, len(reviews))
	for i, r := range reviews {
		ids[i] = r.ID
	}
	return ids
}

// ListByBook retorna uma página de resenhas enriquecidas com contagens e status de like.
// Usa 3 queries fixas (bulk) independente do tamanho da página — sem N+1.
func (s *ReviewService) ListByBook(ctx context.Context, bookID, requesterID uuid.UUID, p pagination.Params) (pagination.Page[domain.Review], error) {
	reviews, total, err := s.reviews.FindByBookID(ctx, bookID, p)
	if err != nil {
		return pagination.Page[domain.Review]{}, fmt.Errorf("review_service.ListByBook: %w", err)
	}

	if len(reviews) > 0 {
		if err := s.enrichReviews(ctx, reviews, requesterID); err != nil {
			return pagination.Page[domain.Review]{}, err
		}
	}

	return pagination.New(reviews, total, p), nil
}

// ListByUser retorna uma página de resenhas de um usuário.
func (s *ReviewService) ListByUser(ctx context.Context, userID uuid.UUID, p pagination.Params) (pagination.Page[domain.Review], error) {
	reviews, total, err := s.reviews.FindByUserID(ctx, userID, p)
	if err != nil {
		return pagination.Page[domain.Review]{}, fmt.Errorf("review_service.ListByUser: %w", err)
	}
	return pagination.New(reviews, total, p), nil
}

// enrichReviews preenche LikesCount, CommentsCount e LikedByMe usando 3 queries bulk.
func (s *ReviewService) enrichReviews(ctx context.Context, reviews []domain.Review, requesterID uuid.UUID) error {
	ids := extractIDs(reviews)

	likeCounts, err := s.reviews.BulkLikeCounts(ctx, ids)
	if err != nil {
		return fmt.Errorf("review_service.enrichReviews likes: %w", err)
	}

	commentCounts, err := s.reviews.BulkCommentCounts(ctx, ids)
	if err != nil {
		return fmt.Errorf("review_service.enrichReviews comments: %w", err)
	}

	likedByUser, err := s.reviews.BulkLikedByUser(ctx, ids, requesterID)
	if err != nil {
		return fmt.Errorf("review_service.enrichReviews liked: %w", err)
	}

	for i := range reviews {
		reviews[i].LikesCount = int(likeCounts[reviews[i].ID])
		reviews[i].CommentsCount = int(commentCounts[reviews[i].ID])
		reviews[i].LikedByMe = likedByUser[reviews[i].ID]
	}

	return nil
}

// Create cria uma nova resenha.
// Retorna apperr.ErrConflict se o usuário já tiver resenha para o livro.
func (s *ReviewService) Create(ctx context.Context, input CreateReviewInput) (*domain.Review, error) {
	if s.reviews.ExistsByBookAndUser(ctx, input.BookID, input.UserID) {
		return nil, apperr.ErrConflict
	}

	review := &domain.Review{
		BookID:  input.BookID,
		UserID:  input.UserID,
		Content: input.Content,
		Rating:  input.Rating,
		Spoiler: input.Spoiler,
	}

	if err := s.reviews.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("review_service.Create: %w", err)
	}

	return review, nil
}

// Update atualiza uma resenha. Retorna apperr.ErrForbidden se o usuário não for o autor.
func (s *ReviewService) Update(ctx context.Context, id, userID uuid.UUID, input UpdateReviewInput) error {
	review, err := s.reviews.FindByID(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}
	if review.UserID != userID {
		return apperr.ErrForbidden
	}

	updates := make(map[string]interface{}, 3)
	if input.Content != "" {
		updates["content"] = input.Content
	}
	if input.Rating != 0 {
		updates["rating"] = input.Rating
	}
	if input.Spoiler != nil {
		updates["spoiler"] = *input.Spoiler
	}

	if err := s.reviews.Update(ctx, id, updates); err != nil {
		return fmt.Errorf("review_service.Update: %w", err)
	}
	return nil
}

// Delete remove uma resenha. Admins podem deletar qualquer resenha.
// Retorna apperr.ErrNotFound se a resenha não existir ou o usuário não tiver permissão.
func (s *ReviewService) Delete(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error {
	deleted, err := s.reviews.Delete(ctx, id, userID, isAdmin)
	if err != nil {
		return fmt.Errorf("review_service.Delete: %w", err)
	}
	if !deleted {
		return apperr.ErrNotFound
	}
	return nil
}

// Like registra a curtida de um usuário. Retorna apperr.ErrConflict se já curtiu.
func (s *ReviewService) Like(ctx context.Context, reviewID, userID uuid.UUID) error {
	if err := s.reviews.Like(ctx, reviewID, userID); err != nil {
		return apperr.ErrConflict
	}
	return nil
}

// Unlike remove a curtida de um usuário de uma resenha.
func (s *ReviewService) Unlike(ctx context.Context, reviewID, userID uuid.UUID) error {
	if err := s.reviews.Unlike(ctx, reviewID, userID); err != nil {
		return fmt.Errorf("review_service.Unlike: %w", err)
	}
	return nil
}

// ListComments retorna uma página de comentários de uma resenha.
func (s *ReviewService) ListComments(ctx context.Context, reviewID uuid.UUID, p pagination.Params) (pagination.Page[domain.Comment], error) {
	comments, total, err := s.comments.FindByReviewID(ctx, reviewID, p)
	if err != nil {
		return pagination.Page[domain.Comment]{}, fmt.Errorf("review_service.ListComments: %w", err)
	}
	return pagination.New(comments, total, p), nil
}

// CreateComment adiciona um comentário a uma resenha.
func (s *ReviewService) CreateComment(ctx context.Context, reviewID, userID uuid.UUID, content string) (*domain.Comment, error) {
	comment := &domain.Comment{
		ReviewID: reviewID,
		UserID:   userID,
		Content:  content,
	}
	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("review_service.CreateComment: %w", err)
	}
	return comment, nil
}
