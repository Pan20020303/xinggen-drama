package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/ai"
	"gorm.io/gorm"
)

func (s *StoryboardService) OptimizeVideoPrompt(userID uint, storyboardID string, rawPrompt string, model string) (string, error) {
	var storyboard models.Storyboard
	if err := s.db.Where("id = ? AND user_id = ?", storyboardID, userID).First(&storyboard).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("storyboard not found")
		}
		return "", err
	}

	basePrompt := strings.TrimSpace(rawPrompt)
	if basePrompt == "" && storyboard.VideoPrompt != nil {
		basePrompt = strings.TrimSpace(*storyboard.VideoPrompt)
	}
	if basePrompt == "" {
		basePrompt = s.generateVideoPrompt(buildStoryboardSnapshot(storyboard))
	}
	if basePrompt == "" {
		return "", errors.New("empty prompt")
	}

	client, actualModel, billingRefID, err := reserveTextClient(
		s.aiService,
		s.billing,
		userID,
		model,
		"video_prompt_optimize:"+storyboardID,
	)
	if err != nil {
		return "", err
	}

	success := false
	defer func() {
		if !success && billingRefID != "" {
			_ = s.billing.RefundAI(billingRefID)
		}
	}()

	systemPrompt := s.videoPromptOptimizeSystemPrompt()
	userPrompt := s.videoPromptOptimizeUserPrompt(&storyboard, basePrompt)

	optimized, err := client.GenerateText(
		userPrompt,
		systemPrompt,
		ai.WithTemperature(0.4),
		ai.WithMaxTokens(800),
	)
	if err != nil {
		return "", err
	}

	optimized = normalizeOptimizedPrompt(optimized)
	if optimized == "" {
		return "", errors.New("optimized prompt is empty")
	}

	// 与视频接口约束一致，避免后续提交被拦截（min=5,max=2000）
	if len([]rune(optimized)) > 2000 {
		optimized = string([]rune(optimized)[:2000])
	}
	if len([]rune(strings.TrimSpace(optimized))) < 5 {
		return "", errors.New("optimized prompt too short")
	}

	if err := s.db.Model(&storyboard).Update("video_prompt", optimized).Error; err != nil {
		return "", fmt.Errorf("failed to save optimized prompt: %w", err)
	}

	s.log.Infow("Video prompt optimized",
		"storyboard_id", storyboardID,
		"user_id", userID,
		"model", actualModel,
		"length", len([]rune(optimized)))
	success = true
	return optimized, nil
}

func buildStoryboardSnapshot(storyboard models.Storyboard) Storyboard {
	sb := Storyboard{
		ShotNumber:  storyboard.StoryboardNumber,
		Duration:    storyboard.Duration,
		SceneID:     storyboard.SceneID,
		Characters:  nil,
		IsPrimary:   false,
		Title:       derefString(storyboard.Title),
		ShotType:    derefString(storyboard.ShotType),
		Angle:       derefString(storyboard.Angle),
		Time:        derefString(storyboard.Time),
		Location:    derefString(storyboard.Location),
		Movement:    derefString(storyboard.Movement),
		Action:      derefString(storyboard.Action),
		Dialogue:    derefString(storyboard.Dialogue),
		Result:      derefString(storyboard.Result),
		Atmosphere:  derefString(storyboard.Atmosphere),
		Emotion:     "",
		BgmPrompt:   derefString(storyboard.BgmPrompt),
		SoundEffect: derefString(storyboard.SoundEffect),
	}
	return sb
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func normalizeOptimizedPrompt(raw string) string {
	out := strings.TrimSpace(raw)
	out = strings.TrimPrefix(out, "```text")
	out = strings.TrimPrefix(out, "```markdown")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)

	prefixes := []string{
		"优化后的提示词：",
		"优化后提示词：",
		"Optimized prompt:",
		"Optimized video prompt:",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(out), strings.ToLower(p)) {
			out = strings.TrimSpace(strings.TrimPrefix(out, p))
			break
		}
	}
	return strings.TrimSpace(out)
}

func (s *StoryboardService) videoPromptOptimizeSystemPrompt() string {
	if s.promptI18n != nil && s.promptI18n.IsEnglish() {
		return "You are a professional image-to-video prompt engineer. Rewrite the input into one production-ready prompt for video generation with optional reference images. Keep character identity, scene continuity and camera logic consistent. Keep output concise, vivid, and directly usable by a video model. Return plain text only."
	}
	return "你是专业的图生视频提示词工程师。请把输入重写为一条可直接用于视频生成模型的高质量提示词。必须保持角色一致性、场景连续性、运镜逻辑和动作节奏。输出精炼但信息完整，直接返回纯文本提示词，不要任何解释。"
}

func (s *StoryboardService) videoPromptOptimizeUserPrompt(storyboard *models.Storyboard, basePrompt string) string {
	var context []string
	context = append(context, fmt.Sprintf("当前提示词：%s", basePrompt))

	if storyboard.Title != nil && strings.TrimSpace(*storyboard.Title) != "" {
		context = append(context, fmt.Sprintf("镜头标题：%s", strings.TrimSpace(*storyboard.Title)))
	}
	if storyboard.Action != nil && strings.TrimSpace(*storyboard.Action) != "" {
		context = append(context, fmt.Sprintf("动作：%s", strings.TrimSpace(*storyboard.Action)))
	}
	if storyboard.Dialogue != nil && strings.TrimSpace(*storyboard.Dialogue) != "" {
		context = append(context, fmt.Sprintf("对话：%s", strings.TrimSpace(*storyboard.Dialogue)))
	}
	if storyboard.Movement != nil && strings.TrimSpace(*storyboard.Movement) != "" {
		context = append(context, fmt.Sprintf("运镜：%s", strings.TrimSpace(*storyboard.Movement)))
	}
	if storyboard.ShotType != nil && strings.TrimSpace(*storyboard.ShotType) != "" {
		context = append(context, fmt.Sprintf("景别：%s", strings.TrimSpace(*storyboard.ShotType)))
	}
	if storyboard.Angle != nil && strings.TrimSpace(*storyboard.Angle) != "" {
		context = append(context, fmt.Sprintf("角度：%s", strings.TrimSpace(*storyboard.Angle)))
	}
	if storyboard.Location != nil && strings.TrimSpace(*storyboard.Location) != "" {
		context = append(context, fmt.Sprintf("场景：%s", strings.TrimSpace(*storyboard.Location)))
	}
	if storyboard.Time != nil && strings.TrimSpace(*storyboard.Time) != "" {
		context = append(context, fmt.Sprintf("时间：%s", strings.TrimSpace(*storyboard.Time)))
	}
	if storyboard.Atmosphere != nil && strings.TrimSpace(*storyboard.Atmosphere) != "" {
		context = append(context, fmt.Sprintf("氛围：%s", strings.TrimSpace(*storyboard.Atmosphere)))
	}
	if storyboard.Result != nil && strings.TrimSpace(*storyboard.Result) != "" {
		context = append(context, fmt.Sprintf("结果：%s", strings.TrimSpace(*storyboard.Result)))
	}
	if storyboard.Duration > 0 {
		context = append(context, fmt.Sprintf("时长：%d秒", storyboard.Duration))
	}

	context = append(context, "要求：1) 输出仅一段可直接投喂视频模型的提示词；2) 禁止解释说明；3) 长度控制在80-400字（英文约40-220词）；4) 保留或强化动作与镜头连贯性。")
	return strings.Join(context, "\n")
}
