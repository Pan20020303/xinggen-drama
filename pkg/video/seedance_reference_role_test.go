package video

import (
	"errors"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVolcesArkClient_SeedanceMultipleReferenceImages_UseReferenceRole(t *testing.T) {
	type requestCapture struct {
		Model    string `json:"model"`
		TaskType string `json:"task_type"`
		Content  []struct {
			Type string `json:"type"`
			Role string `json:"role,omitempty"`
		} `json:"content"`
	}

	var captured requestCapture
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/contents/generations/tasks" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"task_1","status":"queued"}`))
	}))
	defer srv.Close()

	client := NewVolcesArkClient(srv.URL, "test-key", "seedance1.5pro", "/contents/generations/tasks", "/contents/generations/tasks/{taskId}")
	_, err := client.GenerateVideo("", "test prompt", WithReferenceImages([]string{"img1", "img2"}))
	if err != nil {
		t.Fatalf("GenerateVideo returned error: %v", err)
	}

	if captured.TaskType != "i2v" {
		t.Fatalf("expected task_type=i2v, got %q", captured.TaskType)
	}

	imageCount := 0
	for _, c := range captured.Content {
		if c.Type != "image_url" {
			continue
		}
		imageCount++
		if c.Role != "reference_image" {
			t.Fatalf("expected image role reference_image, got %q", c.Role)
		}
	}
	if imageCount != 2 {
		t.Fatalf("expected 2 image content items, got %d", imageCount)
	}
}

func TestChatfireClient_SeedanceMultipleReferenceImages_UseReferenceRole(t *testing.T) {
	type requestCapture struct {
		Model    string `json:"model"`
		TaskType string `json:"task_type"`
		Content  []struct {
			Type string `json:"type"`
			Role string `json:"role,omitempty"`
		} `json:"content"`
	}

	var captured requestCapture
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/video/generations" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"task_2","status":"processing"}`))
	}))
	defer srv.Close()

	client := NewChatfireClient(srv.URL, "test-key", "seedance1.5pro", "/video/generations", "/video/task/{taskId}")
	_, err := client.GenerateVideo("", "test prompt", WithReferenceImages([]string{"img1", "img2"}))
	if err != nil {
		t.Fatalf("GenerateVideo returned error: %v", err)
	}

	if captured.TaskType != "i2v" {
		t.Fatalf("expected task_type=i2v, got %q", captured.TaskType)
	}

	imageCount := 0
	for _, c := range captured.Content {
		if c.Type != "image_url" {
			continue
		}
		imageCount++
		if c.Role != "reference_image" {
			t.Fatalf("expected image role reference_image, got %q", c.Role)
		}
	}
	if imageCount != 2 {
		t.Fatalf("expected 2 image content items, got %d", imageCount)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestVolcesArkClient_GenerateVideo_RetriesOnTLSHandshakeTimeout(t *testing.T) {
	attempts := 0
	client := NewVolcesArkClient("https://example.com", "test-key", "doubao-seedance-1-5-pro-251215", "/api/v3/contents/generations/tasks", "/api/v3/contents/generations/tasks/{taskId}")
	client.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("net/http: TLS handshake timeout")
			}
			if req.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", req.Method)
			}
			if req.URL.Path != "/api/v3/contents/generations/tasks" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"id":"task_retry_ok","status":"queued"}`)),
				Request:    req,
			}, nil
		}),
	}

	result, err := client.GenerateVideo("", "test prompt")
	if err != nil {
		t.Fatalf("GenerateVideo returned error: %v", err)
	}
	if result.TaskID != "task_retry_ok" {
		t.Fatalf("expected task id task_retry_ok, got %q", result.TaskID)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestVolcesArkClient_GetTaskStatus_UpdatesUsageFromResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/contents/generations/tasks/task_usage_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"task_usage_1",
			"status":"succeeded",
			"content":{"video_url":"https://example.com/video.mp4"},
			"usage":{"completion_tokens":777,"total_tokens":777}
		}`))
	}))
	defer srv.Close()

	client := NewVolcesArkClient(srv.URL, "test-key", "doubao-seedance-1-5-pro-251215", "/contents/generations/tasks", "/contents/generations/tasks/{taskId}")
	result, err := client.GetTaskStatus("task_usage_1")
	if err != nil {
		t.Fatalf("GetTaskStatus returned error: %v", err)
	}

	if result.Usage.TotalTokens != 777 || result.Usage.CompletionTokens != 777 {
		t.Fatalf("expected usage total/completion to be 777, got total=%d completion=%d", result.Usage.TotalTokens, result.Usage.CompletionTokens)
	}
	last := client.GetLastUsage()
	if last.TotalTokens != 777 || last.CompletionTokens != 777 {
		t.Fatalf("expected last usage total/completion to be 777, got total=%d completion=%d", last.TotalTokens, last.CompletionTokens)
	}
}
