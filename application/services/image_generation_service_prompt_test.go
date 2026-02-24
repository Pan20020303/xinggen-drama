package services

import "testing"

func TestShouldPrependStylePrompt(t *testing.T) {
	tests := []struct {
		name      string
		imageType string
		prompt    string
		want      bool
	}{
		{
			name:      "non-character always prepend",
			imageType: "scene",
			prompt:    "普通场景提示词",
			want:      true,
		},
		{
			name:      "character turnaround zh should not prepend",
			imageType: "character",
			prompt:    "角色三视图设定图，单张画布左中右布局",
			want:      false,
		},
		{
			name:      "character turnaround en should not prepend",
			imageType: "character",
			prompt:    "Character turnaround sheet with triptych layout, front view, side view, back view",
			want:      false,
		},
		{
			name:      "normal character prompt still prepend",
			imageType: "character",
			prompt:    "young warrior, cinematic portrait",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPrependStylePrompt(tt.imageType, tt.prompt)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
