package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/drama-generator/backend/application/dto"
	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/domain/repository"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserDisabled         = errors.New("user disabled")
	ErrAdminAccessDenied    = errors.New("admin access denied")
	ErrTokenRefreshTooEarly = errors.New("token refresh too early")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidOldPassword   = errors.New("invalid old password")
)

type AuthService struct {
	repo             repository.UserRepository
	log              *logger.Logger
	jwtSecret        string
	tokenExpire      time.Duration
	refreshThreshold time.Duration
	initialCredits   int
}

type TokenClaims struct {
	UserID uint            `json:"user_id"`
	Role   models.UserRole `json:"role"`
	Email  string          `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(repo repository.UserRepository, cfg *config.Config, log *logger.Logger) *AuthService {
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
		repo:             repo,
		log:              log,
		jwtSecret:        secret,
		tokenExpire:      time.Duration(expireHours) * time.Hour,
		refreshThreshold: 24 * time.Hour,
		initialCredits:   initialCredits,
	}
}

func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, ErrEmailAlreadyExists
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

	if err := s.repo.CreateWithInitialCredits(&user, s.initialCredits); err != nil {
		return nil, err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status == models.UserStatusDisabled {
		return nil, ErrUserDisabled
	}

	token, err := s.GenerateToken(*user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) GenerateToken(user models.User) (string, error) {
	return s.generateTokenWithAudience(user, "user")
}

func (s *AuthService) GenerateAdminToken(user models.User) (string, error) {
	return s.generateTokenWithAudience(user, "admin")
}

func (s *AuthService) generateTokenWithAudience(user models.User, audience string) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenExpire)),
			Audience:  []string{audience},
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        fmt.Sprintf("%d", now.UnixNano()),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) AdminLogin(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	resp, err := s.Login(req)
	if err != nil {
		return nil, err
	}
	if !IsPlatformAdminRole(resp.User.Role) {
		return nil, ErrAdminAccessDenied
	}

	token, err := s.GenerateAdminToken(resp.User)
	if err != nil {
		return nil, err
	}
	resp.Token = token
	return resp, nil
}

func IsPlatformAdminRole(role models.UserRole) bool {
	return role == models.RolePlatformAdmin || role == models.RoleAdmin
}

func (s *AuthService) RefreshToken(oldToken string) (*dto.AuthResponse, error) {
	claims, err := s.ParseToken(oldToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}

	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
		return nil, ErrTokenExpired
	}
	if time.Until(claims.ExpiresAt.Time) > s.refreshThreshold {
		return nil, ErrTokenRefreshTooEarly
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status == models.UserStatusDisabled {
		return nil, ErrUserDisabled
	}

	audience := "user"
	if len(claims.Audience) > 0 {
		audience = claims.Audience[0]
	}
	token, err := s.generateTokenWithAudience(*user, audience)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) ChangePassword(userID uint, req *dto.ChangePasswordRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrInvalidOldPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.UpdatePassword(userID, string(hash))
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
	return s.repo.FindByID(userID)
}
