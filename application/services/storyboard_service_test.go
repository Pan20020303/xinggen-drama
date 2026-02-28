package services

import (
	"strconv"
	"strings"
	"testing"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func newStoryboardServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return db
}

func newStoryboardServiceForTest(t *testing.T) (*StoryboardService, *gorm.DB) {
	t.Helper()
	db := newStoryboardServiceTestDB(t)
	cfg := &config.Config{}
	svc := NewStoryboardService(db, cfg, logger.NewLogger(true))
	return svc, db
}

func TestEstimateStoryboardMaxTokens(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int
	}{
		{name: "small script keeps safe floor", length: 200, want: 4000},
		{name: "medium script scales up", length: 2000, want: 4500},
		{name: "long script grows further", length: 6000, want: 11500},
		{name: "very long script capped", length: 30000, want: 32000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateStoryboardMaxTokens(tt.length)
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestBuildStoryboardPromptIsCompact(t *testing.T) {
	svc, _ := newStoryboardServiceForTest(t)

	prompt := svc.buildStoryboardPrompt(
		"苏晚推门而入，看到桌上的信，愣住了。",
		`[{"id":1,"name":"苏晚"}]`,
		`[{"id":10,"location":"书房","time":"深夜"}]`,
	)

	if !strings.Contains(prompt, "【剧本内容】") {
		t.Fatalf("expected prompt to include script label")
	}
	if !strings.Contains(prompt, `"storyboards"`) {
		t.Fatalf("expected prompt to include output schema")
	}
	if strings.Contains(prompt, "【分镜要素】每个镜头聚焦单一动作，描述要详尽具体") {
		t.Fatalf("expected verbose legacy prompt block to be removed")
	}
	if strings.Contains(prompt, `"title": "噩梦惊醒"`) {
		t.Fatalf("expected large example payload to be removed")
	}
}

func TestSaveStoryboardsPersistsBatchAndAssociations(t *testing.T) {
	svc, db := newStoryboardServiceForTest(t)

	drama := models.Drama{UserID: 1, Title: "测试项目", Style: "realistic", Status: "draft"}
	if err := db.Create(&drama).Error; err != nil {
		t.Fatalf("create drama: %v", err)
	}

	episode := models.Episode{UserID: 1, DramaID: drama.ID, EpisodeNum: 1, Title: "第1集", Status: "draft"}
	if err := db.Create(&episode).Error; err != nil {
		t.Fatalf("create episode: %v", err)
	}

	char1 := models.Character{UserID: 1, DramaID: drama.ID, Name: "苏晚"}
	char2 := models.Character{UserID: 1, DramaID: drama.ID, Name: "林深"}
	if err := db.Create(&char1).Error; err != nil {
		t.Fatalf("create char1: %v", err)
	}
	if err := db.Create(&char2).Error; err != nil {
		t.Fatalf("create char2: %v", err)
	}

	oldTitle := "旧镜头"
	oldLocation := "旧地点"
	oldTime := "深夜"
	oldAction := "旧动作"
	oldDescription := "旧描述"
	oldStoryboard := models.Storyboard{
		UserID:           1,
		EpisodeID:        episode.ID,
		StoryboardNumber: 99,
		Title:            &oldTitle,
		Location:         &oldLocation,
		Time:             &oldTime,
		Action:           &oldAction,
		Description:      &oldDescription,
	}
	if err := db.Create(&oldStoryboard).Error; err != nil {
		t.Fatalf("create old storyboard: %v", err)
	}

	oldImage := models.ImageGeneration{
		UserID:       1,
		StoryboardID: &oldStoryboard.ID,
		DramaID:      drama.ID,
		ImageType:    string(models.ImageTypeStoryboard),
		Provider:     "test",
		Prompt:       "old",
		Model:        "mock",
		Size:         "1024x1024",
		Quality:      "standard",
		Status:       models.ImageStatusCompleted,
	}
	if err := db.Create(&oldImage).Error; err != nil {
		t.Fatalf("create old image: %v", err)
	}

	err := svc.saveStoryboards(strconv.Itoa(int(episode.ID)), []Storyboard{
		{
			ShotNumber:  1,
			Title:       "推门",
			ShotType:    "中景",
			Angle:       "平视",
			Time:        "深夜，窗外有雨声",
			Location:    "书房内，灯光昏黄",
			Movement:    "固定镜头",
			Action:      "苏晚推门进入书房，视线落在桌面的信封上。",
			Dialogue:    "苏晚：这是谁留下的？",
			Result:      "她停在门口，气氛忽然安静下来。",
			Atmosphere:  "压抑安静，只有雨声。",
			Emotion:     "疑惑",
			Duration:    6,
			BgmPrompt:   "低沉钢琴",
			SoundEffect: "推门声，雨声",
			Characters:  []uint{char1.ID},
			IsPrimary:   true,
		},
		{
			ShotNumber:  2,
			Title:       "拆信",
			ShotType:    "近景",
			Angle:       "俯视",
			Time:        "深夜，台灯照亮桌面",
			Location:    "书房书桌前",
			Movement:    "推镜",
			Action:      "苏晚伸手拆开信封，林深站在她身后看着。",
			Dialogue:    "林深：先别急，看看里面写了什么。",
			Result:      "纸张被缓慢抽出，两人神色都紧绷起来。",
			Atmosphere:  "紧张克制，空气像凝住。",
			Emotion:     "紧张",
			Duration:    7,
			BgmPrompt:   "持续悬疑音",
			SoundEffect: "纸张摩擦声",
			Characters:  []uint{char1.ID, char2.ID},
			IsPrimary:   true,
		},
	})
	if err != nil {
		t.Fatalf("save storyboards: %v", err)
	}

	var saved []models.Storyboard
	if err := db.Preload("Characters").Order("storyboard_number asc").Find(&saved, "episode_id = ?", episode.ID).Error; err != nil {
		t.Fatalf("load saved storyboards: %v", err)
	}
	if len(saved) != 2 {
		t.Fatalf("expected 2 saved storyboards, got %d", len(saved))
	}
	if len(saved[0].Characters) != 1 || len(saved[1].Characters) != 2 {
		t.Fatalf("expected character associations to be preserved, got %d and %d", len(saved[0].Characters), len(saved[1].Characters))
	}

	var joinCount int64
	if err := db.Table("storyboard_characters").Count(&joinCount).Error; err != nil {
		t.Fatalf("count storyboard_characters: %v", err)
	}
	if joinCount != 3 {
		t.Fatalf("expected 3 join rows, got %d", joinCount)
	}

	var refreshedImage models.ImageGeneration
	if err := db.First(&refreshedImage, oldImage.ID).Error; err != nil {
		t.Fatalf("reload old image: %v", err)
	}
	if refreshedImage.StoryboardID != nil {
		t.Fatalf("expected old image storyboard_id to be cleared")
	}
}

func TestSplitScriptIntoSegmentsSplitsLongScript(t *testing.T) {
	svc, _ := newStoryboardServiceForTest(t)

	makeParagraph := func(title string) string {
		return title + strings.Repeat("这是一个用于分段测试的句子。", 120)
	}

	script := strings.Join([]string{
		makeParagraph("【场景一】"),
		makeParagraph("【场景二】"),
		makeParagraph("【场景三】"),
	}, "\n\n")

	segments := svc.splitScriptIntoSegments(script)
	if len(segments) < 2 {
		t.Fatalf("expected script to split into multiple segments, got %d", len(segments))
	}

	joined := strings.Join(segments, "\n")
	for _, marker := range []string{"【场景一】", "【场景二】", "【场景三】"} {
		if !strings.Contains(joined, marker) {
			t.Fatalf("expected joined segments to contain marker %s", marker)
		}
	}
}

func TestSplitScriptIntoSegmentsSplitsSingleLongParagraph(t *testing.T) {
	svc, _ := newStoryboardServiceForTest(t)

	script := strings.Repeat("苏晚抬头看向门口，心跳逐渐加快，却仍强作镇定地向前迈步。", 32)

	segments := svc.splitScriptIntoSegments(script)
	if len(segments) < 2 {
		t.Fatalf("expected single long paragraph to split into multiple segments, got %d", len(segments))
	}

	joined := strings.Join(segments, "")
	if !strings.Contains(joined, "苏晚抬头看向门口") {
		t.Fatalf("expected split segments to preserve original content")
	}
}

func TestSummarizeStoryboardContextUsesLastTwoShots(t *testing.T) {
	svc, _ := newStoryboardServiceForTest(t)

	summary := svc.summarizeStoryboardContext([]Storyboard{
		{ShotNumber: 1, Title: "开门", Location: "客厅", Time: "清晨", Action: "推门进入", Result: "看到桌上信件"},
		{ShotNumber: 2, Title: "停步", Location: "客厅", Time: "清晨", Action: "站定观察", Result: "神色紧张"},
		{ShotNumber: 3, Title: "拆信", Location: "书桌前", Time: "清晨", Action: "伸手拆信", Result: "纸张露出内容"},
	})

	if strings.Contains(summary, "开门") {
		t.Fatalf("expected summary to only keep the latest shots")
	}
	if !strings.Contains(summary, "停步") || !strings.Contains(summary, "拆信") {
		t.Fatalf("expected summary to include the latest two shots, got %s", summary)
	}
}

func TestRenumberStoryboardsSequentially(t *testing.T) {
	storyboards := []Storyboard{
		{ShotNumber: 7, Title: "A"},
		{ShotNumber: 12, Title: "B"},
		{ShotNumber: 99, Title: "C"},
	}

	renumberStoryboards(storyboards, 5)

	for i, sb := range storyboards {
		want := 5 + i
		if sb.ShotNumber != want {
			t.Fatalf("expected shot number %d, got %d", want, sb.ShotNumber)
		}
	}
}

func TestMaxConcurrentStoryboardSegments(t *testing.T) {
	tests := []struct {
		total int
		want  int
	}{
		{total: 0, want: 1},
		{total: 1, want: 1},
		{total: 2, want: 2},
		{total: 3, want: 3},
		{total: 8, want: 3},
	}

	for _, tt := range tests {
		if got := maxConcurrentStoryboardSegments(tt.total); got != tt.want {
			t.Fatalf("total=%d expected %d got %d", tt.total, tt.want, got)
		}
	}
}

func TestMergeStoryboardSegmentResultsKeepsOrderAndRenumbers(t *testing.T) {
	results := []storyboardSegmentResult{
		{
			Index: 0,
			Storyboards: []Storyboard{
				{ShotNumber: 10, Title: "A1"},
				{ShotNumber: 11, Title: "A2"},
			},
		},
		{
			Index: 1,
			Storyboards: []Storyboard{
				{ShotNumber: 99, Title: "B1"},
			},
		},
	}

	merged, err := mergeStoryboardSegmentResults(results)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if len(merged) != 3 {
		t.Fatalf("expected 3 merged storyboards, got %d", len(merged))
	}
	if merged[0].Title != "A1" || merged[1].Title != "A2" || merged[2].Title != "B1" {
		t.Fatalf("expected original segment order to be preserved, got %+v", merged)
	}
	for i, sb := range merged {
		if sb.ShotNumber != i+1 {
			t.Fatalf("expected shot number %d, got %d", i+1, sb.ShotNumber)
		}
	}
}

func TestCountReadyStoryboardSegments_StopsAtFirstGap(t *testing.T) {
	tests := []struct {
		name  string
		ready []bool
		want  int
	}{
		{name: "none ready", ready: []bool{false, false, false}, want: 0},
		{name: "prefix ready", ready: []bool{true, true, false, true}, want: 2},
		{name: "all ready", ready: []bool{true, true, true}, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countReadyStoryboardSegments(tt.ready); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestBuildStoryboardTaskResult_MarksPartialState(t *testing.T) {
	result := buildStoryboardTaskResult([]Storyboard{
		{ShotNumber: 1, Title: "开场"},
		{ShotNumber: 2, Title: "转场"},
	}, 2, 4)

	if result["total"] != 2 {
		t.Fatalf("expected total=2, got %#v", result["total"])
	}
	if result["segments_completed"] != 2 {
		t.Fatalf("expected segments_completed=2, got %#v", result["segments_completed"])
	}
	if result["segment_total"] != 4 {
		t.Fatalf("expected segment_total=4, got %#v", result["segment_total"])
	}
	if result["is_partial"] != true {
		t.Fatalf("expected is_partial=true, got %#v", result["is_partial"])
	}
}

func TestMergeAvailableStoryboardSegmentResults_PublishesCompletedSegmentsWithoutPrefix(t *testing.T) {
	results := []storyboardSegmentResult{
		{Index: 0},
		{
			Index: 1,
			Storyboards: []Storyboard{
				{ShotNumber: 20, Title: "第二段-1"},
			},
		},
		{
			Index: 2,
			Storyboards: []Storyboard{
				{ShotNumber: 30, Title: "第三段-1"},
				{ShotNumber: 31, Title: "第三段-2"},
			},
		},
	}

	merged, err := mergeAvailableStoryboardSegmentResults(results)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if len(merged) != 3 {
		t.Fatalf("expected 3 preview storyboards, got %d", len(merged))
	}
	if merged[0].Title != "第二段-1" || merged[1].Title != "第三段-1" || merged[2].Title != "第三段-2" {
		t.Fatalf("unexpected preview order: %+v", merged)
	}
	for i, sb := range merged {
		if sb.ShotNumber != i+1 {
			t.Fatalf("expected preview shot number %d, got %d", i+1, sb.ShotNumber)
		}
	}
}
