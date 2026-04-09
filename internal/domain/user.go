package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole representa o papel de um usuário dentro do clube.
type UserRole string

// UserStatus representa o estado de aprovação de um usuário.
type UserStatus string

const (
	// RoleAdmin identifica um administrador do clube.
	RoleAdmin UserRole = "admin"

	// RoleMember identifica um membro comum do clube.
	RoleMember UserRole = "member"

	// StatusPending indica que o cadastro aguarda aprovação.
	StatusPending UserStatus = "pending"

	// StatusApproved indica que o membro foi aprovado.
	StatusApproved UserStatus = "approved"

	// StatusRejected indica que o cadastro foi recusado.
	StatusRejected UserStatus = "rejected"
)

// User representa um membro do clube do livro.
type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"  json:"id"`
	Name          string     `gorm:"not null"              json:"name"`
	Email         string     `gorm:"uniqueIndex;not null"  json:"email"`
	PasswordHash  string     `gorm:"not null"              json:"-"`
	Role          UserRole   `gorm:"default:'member'"      json:"role"`
	Status        UserStatus `gorm:"default:'pending'"     json:"status"`
	AvatarURL     string     `                             json:"avatar_url,omitempty"`
	Bio           string     `                             json:"bio,omitempty"`
	FavoriteGenre string     `                             json:"favorite_genre,omitempty"`
	CreatedAt     time.Time  `                             json:"created_at"`
	UpdatedAt     time.Time  `                             json:"updated_at"`
}

// BeforeCreate popula o UUID antes de persistir caso ainda não tenha sido definido.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
