package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db            *gorm.DB
	log           *logger.Logger
	jwtSecret     string
	tokenExpire   time.Duration
	initialCredits int
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type TokenClaims struct {
	UserID uint            `json:"user_id"`
	Role   models.UserRole `json:"role"`
	Email  string          `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AuthService {
	secret := cfg.Auth.JWTSecret
	if secret == "" {
		secret = "change-me-in-production"
	}
	expireHours := cfg.Auth.TokenExpireHours
	if expireHours <= 0 {
		expireHours = 72
	}
	initialCredits := cfg.Auth.InitialCredits
	if initialCredits < 0 {
		initialCredits = 0
	}

	return &AuthService{
		db:            db,
		log:           log,
		jwtSecret:     secret,
		tokenExpire:   time.Duration(expireHours) * time.Hour,
		initialCredits: initialCredits,
	}
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	var existed models.User
	if err := s.db.Where("email = ?", req.Email).First(&existed).Error; err == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.RoleUser,
		Credits:      s.initialCredits,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		desc := "initial credits for new user"
		txn := models.CreditTransaction{
			UserID:      user.ID,
			Amount:      s.initialCredits,
			Type:        models.CreditTxnRecharge,
			Description: &desc,
		}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) GenerateToken(user models.User) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenExpire)),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(token string) (*TokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
