package services

import (
	"encoding/base64"
	"fmt"
	"html"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CaptchaService struct {
	ttl   time.Duration
	mu    sync.Mutex
	items map[string]captchaItem
}

type captchaItem struct {
	answer    string
	expiresAt time.Time
}

type CaptchaPayload struct {
	CaptchaID string
	ImageData string
	Answer    string
}

func NewCaptchaService(ttl time.Duration) *CaptchaService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &CaptchaService{
		ttl:   ttl,
		items: make(map[string]captchaItem),
	}
}

func (s *CaptchaService) Generate() (*CaptchaPayload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked()

	answer := randomCaptchaCode(4)
	id := uuid.NewString()
	imageData := buildCaptchaDataURL(answer)

	s.items[id] = captchaItem{
		answer:    strings.ToUpper(answer),
		expiresAt: time.Now().Add(s.ttl),
	}

	return &CaptchaPayload{
		CaptchaID: id,
		ImageData: imageData,
		Answer:    answer,
	}, nil
}

func (s *CaptchaService) Verify(id string, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked()

	item, ok := s.items[id]
	if !ok {
		return false
	}
	delete(s.items, id)

	return item.answer == strings.ToUpper(strings.TrimSpace(code))
}

func (s *CaptchaService) cleanupExpiredLocked() {
	now := time.Now()
	for id, item := range s.items {
		if now.After(item.expiresAt) {
			delete(s.items, id)
		}
	}
}

func randomCaptchaCode(length int) string {
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	builder := strings.Builder{}
	builder.Grow(length)
	for i := 0; i < length; i++ {
		builder.WriteByte(alphabet[rng.Intn(len(alphabet))])
	}
	return builder.String()
}

func buildCaptchaDataURL(code string) string {
	escapedCode := html.EscapeString(code)
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="140" height="46" viewBox="0 0 140 46">
<rect width="140" height="46" rx="8" fill="#132331"/>
<path d="M12 34 C28 10, 44 10, 60 34" stroke="#2ea8ff" stroke-width="2" opacity="0.35" fill="none"/>
<path d="M80 12 C96 34, 112 34, 128 12" stroke="#7d5cff" stroke-width="2" opacity="0.3" fill="none"/>
<circle cx="24" cy="14" r="2" fill="#ffffff" opacity="0.35"/>
<circle cx="102" cy="31" r="2" fill="#ffffff" opacity="0.3"/>
<text x="70" y="30" text-anchor="middle" font-size="24" font-weight="700" fill="#ffffff" letter-spacing="6" font-family="Arial, sans-serif">%s</text>
</svg>`,
		escapedCode,
	)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
