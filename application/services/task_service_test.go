package services

import (
	"encoding/json"
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newTaskServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&models.AsyncTask{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return db
}

func TestTaskService_CreateOrGetActiveTask_ReusesPendingTask(t *testing.T) {
	db := newTaskServiceTestDB(t)
	svc := NewTaskService(db, logger.NewLogger(true))

	first, created, err := svc.CreateOrGetActiveTask("storyboard_generation", "101")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if !created {
		t.Fatalf("expected first call to create task")
	}

	second, created, err := svc.CreateOrGetActiveTask("storyboard_generation", "101")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if created {
		t.Fatalf("expected second call to reuse active task")
	}
	if second.ID != first.ID {
		t.Fatalf("expected same task id, got first=%s second=%s", first.ID, second.ID)
	}

	var count int64
	if err := db.Model(&models.AsyncTask{}).
		Where("type = ? AND resource_id = ?", "storyboard_generation", "101").
		Count(&count).Error; err != nil {
		t.Fatalf("count error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 task row, got %d", count)
	}
}

func TestTaskService_CreateOrGetActiveTask_CreatesNewAfterCompleted(t *testing.T) {
	db := newTaskServiceTestDB(t)
	svc := NewTaskService(db, logger.NewLogger(true))

	first, created, err := svc.CreateOrGetActiveTask("storyboard_generation", "202")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if !created {
		t.Fatalf("expected first call to create task")
	}

	if err := svc.UpdateTaskStatus(first.ID, "completed", 100, "done"); err != nil {
		t.Fatalf("failed to complete first task: %v", err)
	}

	second, created, err := svc.CreateOrGetActiveTask("storyboard_generation", "202")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if !created {
		t.Fatalf("expected second call to create new task after completion")
	}
	if second.ID == first.ID {
		t.Fatalf("expected new task id after completion")
	}

	var count int64
	if err := db.Model(&models.AsyncTask{}).
		Where("type = ? AND resource_id = ?", "storyboard_generation", "202").
		Count(&count).Error; err != nil {
		t.Fatalf("count error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 task rows, got %d", count)
	}
}

func TestTaskService_UpdateTaskProgressResult_PersistsProcessingPayload(t *testing.T) {
	db := newTaskServiceTestDB(t)
	svc := NewTaskService(db, logger.NewLogger(true))

	task, _, err := svc.CreateOrGetActiveTask("storyboard_generation", "303")
	if err != nil {
		t.Fatalf("create task error: %v", err)
	}

	payload := map[string]any{
		"storyboards": []map[string]any{
			{"shot_number": 1, "title": "开场"},
		},
		"is_partial": true,
	}

	if err := svc.UpdateTaskProgressResult(task.ID, "processing", 35, "已完成 1/3 段", payload); err != nil {
		t.Fatalf("update task progress result error: %v", err)
	}

	saved, err := svc.GetTask(task.ID)
	if err != nil {
		t.Fatalf("get task error: %v", err)
	}
	if saved.Status != "processing" {
		t.Fatalf("expected processing status, got %s", saved.Status)
	}
	if saved.Progress != 35 {
		t.Fatalf("expected progress 35, got %d", saved.Progress)
	}
	if saved.CompletedAt != nil {
		t.Fatalf("expected completed_at to stay nil while processing")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(saved.Result), &parsed); err != nil {
		t.Fatalf("unmarshal saved result error: %v", err)
	}
	if parsed["is_partial"] != true {
		t.Fatalf("expected is_partial=true, got %#v", parsed["is_partial"])
	}
}
