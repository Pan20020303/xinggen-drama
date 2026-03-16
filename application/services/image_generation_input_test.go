package services

import (
	"testing"

	"github.com/drama-generator/backend/domain/models"
)

func TestResolveCharacterGenerationInput_UsesAppearanceAndLocalPath(t *testing.T) {
	appearance := "红斗篷小女孩，童话绘本风"
	description := "这条描述不应优先使用"
	localPath := "/tmp/reference-character.png"

	prompt, resolvedLocalPath := resolveCharacterGenerationInput(
		models.Character{
			Name:        "小红帽",
			Appearance:  &appearance,
			Description: &description,
			LocalPath:   &localPath,
		},
		models.Drama{Style: "fantasy"},
		nil,
	)

	if prompt != "红斗篷小女孩，童话绘本风, fantasy" {
		t.Fatalf("unexpected prompt: %s", prompt)
	}
	if resolvedLocalPath == nil || *resolvedLocalPath != localPath {
		t.Fatalf("expected local path %s, got %#v", localPath, resolvedLocalPath)
	}
}

func TestResolveCharacterGenerationInput_AllowsClearingLocalPath(t *testing.T) {
	appearance := "红斗篷小女孩，童话绘本风"
	localPath := "/tmp/reference-character.png"
	disableLocalPath := ""

	_, resolvedLocalPath := resolveCharacterGenerationInput(
		models.Character{
			Name:       "小红帽",
			Appearance: &appearance,
			LocalPath:  &localPath,
		},
		models.Drama{},
		&disableLocalPath,
	)

	if resolvedLocalPath != nil {
		t.Fatalf("expected local path to be cleared, got %#v", resolvedLocalPath)
	}
}

func TestResolveSceneGenerationInput_UsesStoredPromptAndLocalPath(t *testing.T) {
	localPath := "/tmp/reference-scene.png"
	scene := models.Scene{
		Location:  "木屋厨房",
		Time:      "清晨",
		Prompt:    "童话木屋厨房，暖色晨光",
		LocalPath: &localPath,
	}

	prompt, resolvedLocalPath := resolveSceneGenerationInput(scene, "", nil)
	if prompt != "童话木屋厨房，暖色晨光" {
		t.Fatalf("unexpected prompt: %s", prompt)
	}
	if resolvedLocalPath == nil || *resolvedLocalPath != localPath {
		t.Fatalf("expected local path %s, got %#v", localPath, resolvedLocalPath)
	}
}

func TestResolveSceneGenerationInput_AllowsClearingLocalPath(t *testing.T) {
	localPath := "/tmp/reference-scene.png"
	disableLocalPath := ""
	scene := models.Scene{
		Location:  "木屋厨房",
		Time:      "清晨",
		Prompt:    "童话木屋厨房，暖色晨光",
		LocalPath: &localPath,
	}

	_, resolvedLocalPath := resolveSceneGenerationInput(scene, "", &disableLocalPath)
	if resolvedLocalPath != nil {
		t.Fatalf("expected local path to be cleared, got %#v", resolvedLocalPath)
	}
}
