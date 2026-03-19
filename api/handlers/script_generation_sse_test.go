package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := writeSSE(c, w, "chunk", gin.H{
		"content":  "hello",
		"progress": 0.2,
	})
	if err != nil {
		t.Fatalf("writeSSE returned error: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "event: chunk\n") {
		t.Fatalf("expected chunk event, got %q", body)
	}
	if !strings.Contains(body, "data: ") {
		t.Fatalf("expected data line, got %q", body)
	}

	lines := strings.Split(body, "\n")
	if len(lines) < 2 {
		t.Fatalf("unexpected sse body: %q", body)
	}

	raw := strings.TrimPrefix(lines[1], "data: ")
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["content"] != "hello" {
		t.Fatalf("expected content hello, got %#v", payload["content"])
	}
}
