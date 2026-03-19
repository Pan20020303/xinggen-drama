package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/drama-generator/backend/pkg/usage"
)

func TestDispatchVideoGeneration_PublishesMQJob(t *testing.T) {
	dispatcher := &capturingDispatcher{}
	svc := &VideoGenerationService{dispatcher: dispatcher}

	if err := svc.dispatchVideoGeneration(77); err != nil {
		t.Fatalf("dispatch error: %v", err)
	}

	if dispatcher.job.Type != JobTypeVideoGeneration {
		t.Fatalf("expected job type %s, got %s", JobTypeVideoGeneration, dispatcher.job.Type)
	}

	var payload VideoGenerationJobPayload
	if err := json.Unmarshal(dispatcher.job.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload error: %v", err)
	}
	if payload.VideoGenerationID != 77 {
		t.Fatalf("expected video generation id 77, got %d", payload.VideoGenerationID)
	}
}

func TestDispatchVideoPollStatus_PublishesDelayedMQJob(t *testing.T) {
	dispatcher := &capturingDispatcher{}
	svc := &VideoGenerationService{dispatcher: dispatcher}

	payload := VideoPollStatusJobPayload{
		VideoGenerationID: 77,
		TaskID:            "remote-task-1",
		RecordedUsage:     usage.TokenUsage{CompletionTokens: 12, TotalTokens: 12},
		Attempt:           3,
	}

	if err := svc.dispatchVideoPollStatus(payload, 15*time.Second); err != nil {
		t.Fatalf("dispatch error: %v", err)
	}

	if dispatcher.delayedJob.Type != JobTypeVideoPollStatus {
		t.Fatalf("expected delayed job type %s, got %s", JobTypeVideoPollStatus, dispatcher.delayedJob.Type)
	}
	if dispatcher.delay != 15*time.Second {
		t.Fatalf("expected delay 15s, got %s", dispatcher.delay)
	}

	var actual VideoPollStatusJobPayload
	if err := json.Unmarshal(dispatcher.delayedJob.Payload, &actual); err != nil {
		t.Fatalf("unmarshal payload error: %v", err)
	}
	if actual.VideoGenerationID != payload.VideoGenerationID || actual.TaskID != payload.TaskID || actual.Attempt != payload.Attempt {
		t.Fatalf("unexpected payload: %#v", actual)
	}
}
