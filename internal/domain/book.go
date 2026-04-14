package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Book representa o livro escolhido para um determinado mês do clube.
type Book struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"                      json:"id"`
	Title       string    `gorm:"not null"                                  json:"title"`
	Author      string    `gorm:"not null"                                  json:"author"`
	Description string    `                                                 json:"description,omitempty"`
	CoverURL    string    `                                                 json:"cover_url,omitempty"`
	Month       int       `gorm:"not null;uniqueIndex:idx_books_month_year" json:"month"`
	Year        int       `gorm:"not null;uniqueIndex:idx_books_month_year" json:"year"`
	CreatedByID uuid.UUID `gorm:"type:uuid;not null"                        json:"created_by_id"`
	CreatedAt   time.Time `                                                 json:"created_at"`
	UpdatedAt   time.Time `                                                 json:"updated_at"`
}

// BeforeCreate popula o UUID antes de persistir caso ainda não tenha sido definido.
func (b *Book) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
