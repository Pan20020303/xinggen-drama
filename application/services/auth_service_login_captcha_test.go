package services

import (
	"testing"

	"github.com/drama-generator/backend/application/dto"
	"github.com/drama-generator/backend/domain/models"
)

func TestLoginRequiresValidCaptcha(t *testing.T) {
	svc, db := newTestAuthService(t)
	seedAuthUser(t, db, "captcha@example.com", models.RoleUser, models.UserStatusActive)

	_, err := svc.Login(&dto.LoginRequest{
		Email:       "captcha@example.com",
		Password:    "Passw0rd123",
		CaptchaID:   "missing",
		CaptchaCode: "1234",
	})
	if err == nil {
		t.Fatalf("expected login to fail without valid captcha")
	}
	if err != ErrInvalidCaptcha {
		t.Fatalf("expected invalid captcha error, got %v", err)
	}
}

func TestLoginSucceedsWithCaptcha(t *testing.T) {
	svc, db := newTestAuthService(t)
	seedAuthUser(t, db, "captcha-ok@example.com", models.RoleUser, models.UserStatusActive)

	captcha, err := svc.GenerateCaptcha()
	if err != nil {
		t.Fatalf("expected captcha generation success, got error: %v", err)
	}

	resp, err := svc.Login(&dto.LoginRequest{
		Email:       "captcha-ok@example.com",
		Password:    "Passw0rd123",
		CaptchaID:   captcha.CaptchaID,
		CaptchaCode: captcha.Answer,
	})
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected token")
	}
}
