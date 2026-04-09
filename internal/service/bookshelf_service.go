package service

import (
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/google/uuid"
)

// bookshelfRepo define as operações de repositório necessárias para BookshelfService.
type bookshelfRepo interface {
	Upsert(entry *domain.UserBook) error
	FindByUserID(userID uuid.UUID) ([]domain.UserBook, error)
}

// BookshelfService gerencia a estante virtual de livros dos usuários.
type BookshelfService struct {
	shelf bookshelfRepo
}

// NewBookshelfService cria um BookshelfService com o repositório fornecido.
func NewBookshelfService(shelf bookshelfRepo) *BookshelfService {
	return &BookshelfService{shelf: shelf}
}

// UpsertBookshelfInput contém os dados para adicionar ou atualizar um livro na estante.
type UpsertBookshelfInput struct {
	UserID        uuid.UUID
	BookID        uuid.UUID
	Status        domain.BookshelfStatus
	PersonalNotes string
}

// GetShelf retorna todos os livros da estante de um usuário.
func (s *BookshelfService) GetShelf(userID uuid.UUID) ([]domain.UserBook, error) {
	return s.shelf.FindByUserID(userID)
}

// Upsert adiciona ou atualiza um livro na estante do usuário.
func (s *BookshelfService) Upsert(input UpsertBookshelfInput) error {
	return s.shelf.Upsert(&domain.UserBook{
		UserID:        input.UserID,
		BookID:        input.BookID,
		Status:        input.Status,
		PersonalNotes: input.PersonalNotes,
	})
}
