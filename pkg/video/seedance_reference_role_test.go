package video

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
