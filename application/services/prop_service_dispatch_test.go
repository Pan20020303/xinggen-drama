package services

import (
	"encoding/json"
	"testing"
)

func TestDispatchPropExtraction_PublishesMQJob(t *testing.T) {
	dispatcher := &capturingDispatcher{}
	svc := &PropService{dispatcher: dispatcher}

	payload := PropExtractionJobPayload{
		UserID:    5,
		TaskID:    "task-prop",
		EpisodeID: 12,
	}

	if err := svc.dispatchPropExtraction(payload); err != nil {
		t.Fatalf("dispatch error: %v", err)
	}

	if dispatcher.job.Type != JobTypePropExtraction {
		t.Fatalf("expected job type %s, got %s", JobTypePropExtraction, dispatcher.job.Type)
	}

	var actual PropExtractionJobPayload
	if err := json.Unmarshal(dispatcher.job.Payload, &actual); err != nil {
		t.Fatalf("unmarshal payload error: %v", err)
	}
	if actual != payload {
		t.Fatalf("unexpected payload: %#v", actual)
	}
}
