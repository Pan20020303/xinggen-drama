package services

import (
	"strings"
	"testing"
	"time"
)

func TestCaptchaService_GenerateAndVerify(t *testing.T) {
	svc := NewCaptchaService(5 * time.Minute)

	resp, err := svc.Generate()
	if err != nil {
		t.Fatalf("expected generate success, got error: %v", err)
	}
	if resp.CaptchaID == "" {
		t.Fatalf("expected captcha id")
	}
	if !strings.HasPrefix(resp.ImageData, "data:image/svg+xml;base64,") {
		t.Fatalf("expected svg data url, got %q", resp.ImageData)
	}

	if !svc.Verify(resp.CaptchaID, resp.Answer) {
		t.Fatalf("expected captcha verification success")
	}

	if svc.Verify(resp.CaptchaID, resp.Answer) {
		t.Fatalf("expected captcha to be one-time use")
	}
}

func TestCaptchaService_VerifyWrongCode(t *testing.T) {
	svc := NewCaptchaService(5 * time.Minute)
	resp, err := svc.Generate()
	if err != nil {
		t.Fatalf("expected generate success, got error: %v", err)
	}

	if svc.Verify(resp.CaptchaID, "wrong") {
		t.Fatalf("expected wrong captcha code to fail")
	}
}

func TestCaptchaService_VerifyExpired(t *testing.T) {
	svc := NewCaptchaService(1 * time.Millisecond)
	resp, err := svc.Generate()
	if err != nil {
		t.Fatalf("expected generate success, got error: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	if svc.Verify(resp.CaptchaID, resp.Answer) {
		t.Fatalf("expected expired captcha to fail")
	}
}
