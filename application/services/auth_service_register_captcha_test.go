package services

import (
	"testing"

	"github.com/drama-generator/backend/application/dto"
)

func TestRegisterRequiresValidCaptcha(t *testing.T) {
	svc, _ := newTestAuthService(t)

	_, err := svc.Register(&dto.RegisterRequest{
		Email:       "register-captcha@example.com",
		Password:    "Passw0rd123",
		CaptchaID:   "missing",
		CaptchaCode: "1234",
	})
	if err == nil {
		t.Fatalf("expected register to fail without valid captcha")
	}
	if err != ErrInvalidCaptcha {
		t.Fatalf("expected invalid captcha error, got %v", err)
	}
}

func TestRegisterSucceedsWithCaptcha(t *testing.T) {
	svc, _ := newTestAuthService(t)
	captcha, err := svc.GenerateCaptcha()
	if err != nil {
		t.Fatalf("expected captcha generation success, got error: %v", err)
	}

	resp, err := svc.Register(&dto.RegisterRequest{
		Email:       "register-ok@example.com",
		Password:    "Passw0rd123",
		CaptchaID:   captcha.CaptchaID,
		CaptchaCode: captcha.Answer,
	})
	if err != nil {
		t.Fatalf("expected register success, got error: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected token")
	}
}
