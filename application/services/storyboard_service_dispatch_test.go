package services

import (
	"encoding/json"
	"testing"
)

func TestDispatchStoryboardGeneration_PublishesMQJob(t *testing.T) {
	dispatcher := &capturingDispatcher{}
	svc := &StoryboardService{dispatcher: dispatcher}

	payload := StoryboardGenerationJobPayload{
		UserID:        9,
		TaskID:        "task-1",
		EpisodeID:     "101",
		Model:         "gpt-4.1",
		ScriptContent: "剧情内容",
		CharacterList: "[{\"id\":1}]",
		SceneList:     "[{\"id\":2}]",
	}

	if err := svc.dispatchStoryboardGeneration(payload); err != nil {
		t.Fatalf("dispatch error: %v", err)
	}

	if dispatcher.job.Type != JobTypeStoryboard {
		t.Fatalf("expected job type %s, got %s", JobTypeStoryboard, dispatcher.job.Type)
	}

	var actual StoryboardGenerationJobPayload
	if err := json.Unmarshal(dispatcher.job.Payload, &actual); err != nil {
		t.Fatalf("unmarshal payload error: %v", err)
	}
	if actual.TaskID != payload.TaskID || actual.EpisodeID != payload.EpisodeID || actual.ScriptContent != payload.ScriptContent {
		t.Fatalf("unexpected payload: %#v", actual)
	}
}
