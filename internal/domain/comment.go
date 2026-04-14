package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment representa um comentário feito em uma resenha.
type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"      json:"id"`
	ReviewID  uuid.UUID `gorm:"type:uuid;not null;index"  json:"review_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"        json:"user_id"`
	Content   string    `gorm:"not null"                  json:"content"`
	User      User      `gorm:"foreignKey:UserID"         json:"user,omitempty"`
	CreatedAt time.Time `gorm:"index"                     json:"created_at"`
	UpdatedAt time.Time `                                  json:"updated_at"`
}

// BeforeCreate popula o UUID antes de persistir caso ainda não tenha sido definido.
func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
