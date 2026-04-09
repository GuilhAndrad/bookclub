package service

import (
	"errors"
	"time"

	"github.com/GuilhAndrad/bookclub/config"
	"github.com/GuilhAndrad/bookclub/internal/domain"
	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// authUserRepo define as operações de repositório necessárias para autenticação.
// A interface é pequena e focada — apenas o que AuthService precisa.
type authUserRepo interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
}

// AuthService gerencia o registro e autenticação de usuários.
type AuthService struct {
	users authUserRepo
	config *config.Config
}

// NewAuthService cria um AuthService com o repositório fornecido.
func NewAuthService(users authUserRepo, cfg *config.Config) *AuthService {
    return &AuthService{
        users:  users,
        config: cfg,
    }
}

// RegisterInput contém os dados necessários para registrar um novo usuário.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

// LoginOutput contém o token JWT e os dados do usuário autenticado.
type LoginOutput struct {
	Token string
	User  *domain.User
}

// Register cria um novo usuário com status pendente de aprovação.
// Retorna apperr.ErrConflict se o e-mail já estiver cadastrado.
func (s *AuthService) Register(input RegisterInput) (*domain.User, error) {
	_, err := s.users.FindByEmail(input.Email)
	if err == nil {
		return nil, apperr.ErrConflict
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         domain.RoleMember,
		Status:       domain.StatusPending,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login autentica um usuário e retorna um token JWT.
// Retorna apperr.ErrUnauthorized para credenciais inválidas,
// apperr.ErrPendingApproval ou apperr.ErrAccountRejected conforme o status.
func (s *AuthService) Login(email, password string) (*LoginOutput, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, apperr.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperr.ErrUnauthorized
	}

	switch user.Status {
	case domain.StatusPending:
		return nil, apperr.ErrPendingApproval
	case domain.StatusRejected:
		return nil, apperr.ErrAccountRejected
	}

	token, err := s.generateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &LoginOutput{Token: token, User: user}, nil
}

// Claims contém os dados codificados no JWT.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// ValidateToken valida um token JWT e retorna os claims extraídos.
// Retorna apperr.ErrUnauthorized para tokens inválidos ou expirados.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, apperr.ErrUnauthorized
        }
        return []byte(s.config.JWTSecret), nil
    })
    if err != nil || !token.Valid {
        return nil, apperr.ErrUnauthorized
    }
    return token.Claims.(*Claims), nil
}

// generateToken gera um JWT assinado com os dados do usuário.
func (s *AuthService) generateToken(userID uuid.UUID, role string) (string, error) {
    exp := time.Duration(s.config.JWTExpirationHours) * time.Hour
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(s.config.JWTSecret))
}