package services

import (
	"strings"
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
)

func TestBuildCharacterTurnaroundPrompt_ZH(t *testing.T) {
	svc := &CharacterLibraryService{
		promptI18n: NewPromptI18n(&config.Config{
			App: config.AppConfig{Language: "zh"},
		}),
	}
	appearance := "黑色短发，军装外套，年轻男性"
	desc := "不会被选中"
	character := &models.Character{
		Name:        "林野",
		Appearance:  &appearance,
		Description: &desc,
	}

	got := svc.buildCharacterTurnaroundPrompt(character, "赛博电影感")

	mustContain := []string{
		"单张画布",
		"正视图、侧视图、背视图",
		"左中右",
		appearance,
		"赛博电影感",
		"禁止出现其他人物",
	}
	for _, item := range mustContain {
		if !strings.Contains(got, item) {
			t.Fatalf("prompt should contain %q, got: %s", item, got)
		}
	}
	if strings.Contains(got, desc) {
		t.Fatalf("prompt should prioritize appearance over description")
	}
}

func TestBuildCharacterTurnaroundPrompt_EN(t *testing.T) {
	svc := &CharacterLibraryService{
		promptI18n: NewPromptI18n(&config.Config{
			App: config.AppConfig{Language: "en"},
		}),
	}
	description := "female scout with short brown hair and utility vest"
	character := &models.Character{
		Name:        "Su Xiao",
		Description: &description,
	}

	got := svc.buildCharacterTurnaroundPrompt(character, "")

	mustContain := []string{
		"Single canvas only",
		"front view, side view, back view",
		"triptych layout",
		description,
		"no additional people",
	}
	for _, item := range mustContain {
		if !strings.Contains(got, item) {
			t.Fatalf("prompt should contain %q, got: %s", item, got)
		}
	}
}
