package domain

import (
	"time"

	"github.com/google/uuid"
)

// BookshelfStatus representa o estado de leitura de um livro na estante do usuário.
type BookshelfStatus string

const (
	// BookshelfReading indica que o usuário está lendo o livro.
	BookshelfReading BookshelfStatus = "reading"

	// BookshelfRead indica que o usuário já leu o livro.
	BookshelfRead BookshelfStatus = "read"

	// BookshelfWantToRead indica que o usuário deseja ler o livro.
	BookshelfWantToRead BookshelfStatus = "want_to_read"
)

// UserBook representa um livro na estante virtual de um usuário.
// Utiliza chave composta (UserID + BookID).
type UserBook struct {
	UserID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"user_id"`
	BookID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"book_id"`
	Status        BookshelfStatus `gorm:"not null"             json:"status"`
	StartedAt     *time.Time      `                            json:"started_at,omitempty"`
	FinishedAt    *time.Time      `                            json:"finished_at,omitempty"`
	PersonalNotes string          `                            json:"personal_notes,omitempty"`
	Book          Book            `gorm:"foreignKey:BookID"    json:"book,omitempty"`
}