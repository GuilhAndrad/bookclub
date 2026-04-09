package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Review representa a resenha de um membro sobre um livro.
type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	BookID    uuid.UUID `gorm:"type:uuid;not null"   json:"book_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"   json:"user_id"`
	Content   string    `gorm:"not null"             json:"content"`
	Rating    int       `gorm:"not null"             json:"rating"`
	Spoiler   bool      `gorm:"default:false"        json:"spoiler"`
	Book      Book      `gorm:"foreignKey:BookID"    json:"book,omitempty"`
	User      User      `gorm:"foreignKey:UserID"    json:"user,omitempty"`
	CreatedAt time.Time `                            json:"created_at"`
	UpdatedAt time.Time `                            json:"updated_at"`

	// LikesCount é calculado em tempo de execução e não persiste no banco.
	LikesCount int `gorm:"-" json:"likes_count"`

	// CommentsCount é calculado em tempo de execução e não persiste no banco.
	CommentsCount int `gorm:"-" json:"comments_count"`

	// LikedByMe indica se o usuário autenticado curtiu esta resenha.
	LikedByMe bool `gorm:"-" json:"liked_by_me"`
}

// BeforeCreate popula o UUID antes de persistir caso ainda não tenha sido definido.
func (r *Review) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}