package services

import (
	"testing"
	"time"

	"github.com/drama-generator/backend/application/dto"
	"github.com/drama-generator/backend/domain/models"
	"github.com/golang-jwt/jwt/v5"
)

func TestRefreshToken_Success(t *testing.T) {
	svc, db := newTestAuthService(t)
	seedAuthUser(t, db, "refresh@example.com", models.RoleUser, models.UserStatusActive)

	loginResp, err := svc.Login(&dto.LoginRequest{
		Email:    "refresh@example.com",
		Password: "Passw0rd123",
	})
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}

	refreshed, err := svc.RefreshToken(loginResp.Token)
	if err != nil {
		t.Fatalf("expected refresh success, got error: %v", err)
	}
	if refreshed.Token == "" {
		t.Fatalf("expected refreshed token")
	}
	if refreshed.Token == loginResp.Token {
		t.Fatalf("expected refreshed token to be different")
	}
}

func TestRefreshToken_Expired(t *testing.T) {
	svc, db := newTestAuthService(t)
	user := seedAuthUser(t, db, "expired@example.com", models.RoleUser, models.UserStatusActive)

	now := time.Now()
	claims := TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			Audience:  []string{"user"},
			Subject:   "1",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to build expired token: %v", err)
	}

	_, err = svc.RefreshToken(token)
	if err == nil {
		t.Fatalf("expected refresh to fail for expired token")
	}
}

func TestChangePassword_Success(t *testing.T) {
	svc, db := newTestAuthService(t)
	user := seedAuthUser(t, db, "change@example.com", models.RoleUser, models.UserStatusActive)

	err := svc.ChangePassword(user.ID, &dto.ChangePasswordRequest{
		OldPassword: "Passw0rd123",
		NewPassword: "NewPassw0rd123",
	})
	if err != nil {
		t.Fatalf("expected change password success, got: %v", err)
	}

	_, err = svc.Login(&dto.LoginRequest{
		Email:    "change@example.com",
		Password: "Passw0rd123",
	})
	if err == nil {
		t.Fatalf("expected old password login to fail")
	}

	_, err = svc.Login(&dto.LoginRequest{
		Email:    "change@example.com",
		Password: "NewPassw0rd123",
	})
	if err != nil {
		t.Fatalf("expected new password login success, got: %v", err)
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	svc, db := newTestAuthService(t)
	user := seedAuthUser(t, db, "wrong-old@example.com", models.RoleUser, models.UserStatusActive)

	err := svc.ChangePassword(user.ID, &dto.ChangePasswordRequest{
		OldPassword: "wrong-password",
		NewPassword: "NewPassw0rd123",
	})
	if err == nil {
		t.Fatalf("expected wrong old password to fail")
	}
}
