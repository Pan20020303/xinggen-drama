package services

import (
	"encoding/json"

	"github.com/drama-generator/backend/pkg/usage"
)

const (
	JobTypeImageGeneration     = "image_generation.process"
	JobTypeVideoGeneration     = "video_generation.process"
	JobTypeVideoPollStatus     = "video_generation.poll_status"
	JobTypeStoryboard          = "storyboard_generation.process"
	JobTypeCharacterExtraction = "character_extraction.process"
	JobTypePropExtraction      = "prop_extraction.process"
)

type AsyncJob struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ImageGenerationJobPayload struct {
	ImageGenerationID uint `json:"image_generation_id"`
}

type VideoGenerationJobPayload struct {
	VideoGenerationID uint `json:"video_generation_id"`
}

type StoryboardGenerationJobPayload struct {
	UserID        uint   `json:"user_id"`
	TaskID        string `json:"task_id"`
	EpisodeID     string `json:"episode_id"`
	Model         string `json:"model"`
	ScriptContent string `json:"script_content"`
	CharacterList string `json:"character_list"`
	SceneList     string `json:"scene_list"`
}

type CharacterExtractionJobPayload struct {
	UserID    uint   `json:"user_id"`
	TaskID    string `json:"task_id"`
	EpisodeID uint   `json:"episode_id"`
}

type PropExtractionJobPayload struct {
	UserID    uint   `json:"user_id"`
	TaskID    string `json:"task_id"`
	EpisodeID uint   `json:"episode_id"`
}

type VideoPollStatusJobPayload struct {
	VideoGenerationID uint             `json:"video_generation_id"`
	TaskID            string           `json:"task_id"`
	RecordedUsage     usage.TokenUsage `json:"recorded_usage"`
	Attempt           int              `json:"attempt"`
}
