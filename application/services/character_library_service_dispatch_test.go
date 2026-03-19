package services

import (
	"encoding/json"
	"testing"
)

func TestDispatchCharacterExtraction_PublishesMQJob(t *testing.T) {
	dispatcher := &capturingDispatcher{}
	svc := &CharacterLibraryService{dispatcher: dispatcher}

	payload := CharacterExtractionJobPayload{
		UserID:    8,
		TaskID:    "task-character",
		EpisodeID: 23,
	}

	if err := svc.dispatchCharacterExtraction(payload); err != nil {
		t.Fatalf("dispatch error: %v", err)
	}

	if dispatcher.job.Type != JobTypeCharacterExtraction {
		t.Fatalf("expected job type %s, got %s", JobTypeCharacterExtraction, dispatcher.job.Type)
	}

	var actual CharacterExtractionJobPayload
	if err := json.Unmarshal(dispatcher.job.Payload, &actual); err != nil {
		t.Fatalf("unmarshal payload error: %v", err)
	}
	if actual != payload {
		t.Fatalf("unexpected payload: %#v", actual)
	}
}
