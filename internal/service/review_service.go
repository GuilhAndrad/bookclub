package service

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/google/uuid"
)

// reviewRepo define as operações de repositório necessárias para ReviewService.
type reviewRepo interface {
	Create(review *domain.Review) error
	FindByBookID(bookID uuid.UUID) ([]domain.Review, error)
	FindByUserID(userID uuid.UUID) ([]domain.Review, error)
	FindByID(id uuid.UUID) (*domain.Review, error)
	ExistsByBookAndUser(bookID, userID uuid.UUID) bool
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id, userID uuid.UUID, isAdmin bool) (bool, error)
	CountLikes(reviewID uuid.UUID) (int64, error)
	CountComments(reviewID uuid.UUID) (int64, error)
	IsLikedBy(reviewID, userID uuid.UUID) bool
	Like(reviewID, userID uuid.UUID) error
	Unlike(reviewID, userID uuid.UUID) error
}

// commentRepo define as operações de repositório necessárias para ReviewService.
type commentRepo interface {
	Create(comment *domain.Comment) error
	FindByReviewID(reviewID uuid.UUID) ([]domain.Comment, error)
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

// ListByBook retorna as resenhas de um livro enriquecidas com contagens e status de like.
func (s *ReviewService) ListByBook(bookID, requesterID uuid.UUID) ([]domain.Review, error) {
	reviews, err := s.reviews.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}

	for i := range reviews {
		likes, _    := s.reviews.CountLikes(reviews[i].ID)
		comments, _ := s.reviews.CountComments(reviews[i].ID)

		reviews[i].LikesCount    = int(likes)
		reviews[i].CommentsCount = int(comments)
		reviews[i].LikedByMe     = s.reviews.IsLikedBy(reviews[i].ID, requesterID)
	}

	return reviews, nil
}

// ListByUser retorna todas as resenhas de um usuário.
func (s *ReviewService) ListByUser(userID uuid.UUID) ([]domain.Review, error) {
	return s.reviews.FindByUserID(userID)
}

// Create cria uma nova resenha.
// Retorna apperr.ErrConflict se o usuário já tiver resenha para o livro.
func (s *ReviewService) Create(input CreateReviewInput) (*domain.Review, error) {
	if s.reviews.ExistsByBookAndUser(input.BookID, input.UserID) {
		return nil, apperr.ErrConflict
	}

	review := &domain.Review{
		BookID:  input.BookID,
		UserID:  input.UserID,
		Content: input.Content,
		Rating:  input.Rating,
		Spoiler: input.Spoiler,
	}

	if err := s.reviews.Create(review); err != nil {
		return nil, err
	}

	return review, nil
}

// Update atualiza uma resenha. Retorna apperr.ErrForbidden se o usuário não for o autor.
func (s *ReviewService) Update(id, userID uuid.UUID, input UpdateReviewInput) error {
	review, err := s.reviews.FindByID(id)
	if err != nil {
		return apperr.ErrNotFound
	}
	if review.UserID != userID {
		return apperr.ErrForbidden
	}

	updates := make(map[string]interface{})
	if input.Content != "" { updates["content"] = input.Content }
	if input.Rating != 0   { updates["rating"] = input.Rating }
	if input.Spoiler != nil { updates["spoiler"] = *input.Spoiler }

	return s.reviews.Update(id, updates)
}

// Delete remove uma resenha. Admins podem deletar qualquer resenha.
// Retorna apperr.ErrNotFound se a resenha não existir ou o usuário não tiver permissão.
func (s *ReviewService) Delete(id, userID uuid.UUID, isAdmin bool) error {
	deleted, err := s.reviews.Delete(id, userID, isAdmin)
	if err != nil {
		return err
	}
	if !deleted {
		return apperr.ErrNotFound
	}
	return nil
}

// Like registra a curtida de um usuário. Retorna apperr.ErrConflict se já curtiu.
func (s *ReviewService) Like(reviewID, userID uuid.UUID) error {
	if err := s.reviews.Like(reviewID, userID); err != nil {
		return apperr.ErrConflict
	}
	return nil
}

// Unlike remove a curtida de um usuário de uma resenha.
func (s *ReviewService) Unlike(reviewID, userID uuid.UUID) error {
	return s.reviews.Unlike(reviewID, userID)
}

// ListComments retorna os comentários de uma resenha em ordem cronológica.
func (s *ReviewService) ListComments(reviewID uuid.UUID) ([]domain.Comment, error) {
	return s.comments.FindByReviewID(reviewID)
}

// CreateComment adiciona um comentário a uma resenha.
func (s *ReviewService) CreateComment(reviewID, userID uuid.UUID, content string) (*domain.Comment, error) {
	comment := &domain.Comment{
		ReviewID: reviewID,
		UserID:   userID,
		Content:  content,
	}
	if err := s.comments.Create(comment); err != nil {
		return nil, err
	}
	return comment, nil
}
