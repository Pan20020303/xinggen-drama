package services

import "testing"

func TestEstimateFramePromptCallCount(t *testing.T) {
	tests := []struct {
		name       string
		frameType  FrameType
		panelCount int
		want       int
		wantErr    bool
	}{
		{name: "first", frameType: FrameTypeFirst, panelCount: 0, want: 1},
		{name: "key", frameType: FrameTypeKey, panelCount: 0, want: 1},
		{name: "last", frameType: FrameTypeLast, panelCount: 0, want: 1},
		{name: "action", frameType: FrameTypeAction, panelCount: 0, want: 1},
		{name: "panel default 3", frameType: FrameTypePanel, panelCount: 0, want: 3},
		{name: "panel 3", frameType: FrameTypePanel, panelCount: 3, want: 3},
		{name: "panel 4", frameType: FrameTypePanel, panelCount: 4, want: 4},
		{name: "panel unsupported count fallback 3", frameType: FrameTypePanel, panelCount: 9, want: 3},
		{name: "unsupported type", frameType: FrameType("unknown"), panelCount: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := estimateFramePromptCallCount(tt.frameType, tt.panelCount)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

