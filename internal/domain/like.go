package domain

import (
	"time"

	"github.com/google/uuid"
)

// Like representa a curtida de um usuário em uma resenha.
// Utiliza chave composta (ReviewID + UserID) para garantir unicidade.
type Like struct {
	ReviewID  uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"review_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"       json:"user_id"`
	CreatedAt time.Time `                                  json:"created_at"`
}
