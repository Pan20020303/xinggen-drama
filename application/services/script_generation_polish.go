package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/ai"
	"gorm.io/gorm"
)

var ErrEpisodeNotFound = errors.New("episode not found")

const defaultPolishTextModel = "doubao-seed-1-8-251228"

func (s *ScriptGenerationService) PolishEpisodeScript(userID uint, episodeID uint, rawContent string, model string, skillName string) (string, string, error) {
	var episode models.Episode
	if err := s.db.Where("id = ? AND user_id = ?", episodeID, userID).First(&episode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrEpisodeNotFound
		}
		return "", "", err
	}

	content := strings.TrimSpace(rawContent)
	if content == "" && episode.ScriptContent != nil {
		content = strings.TrimSpace(*episode.ScriptContent)
	}
	if content == "" {
		return "", "", errors.New("empty content")
	}

	return s.polishContentWithSkill(
		userID,
		content,
		model,
		skillName,
		fmt.Sprintf("episode_script_polish:%d", episodeID),
	)
}

func (s *ScriptGenerationService) PolishScriptText(userID uint, content string, model string, skillName string) (string, string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", "", errors.New("empty content")
	}
	return s.polishContentWithSkill(
		userID,
		content,
		model,
		skillName,
		"script_polish",
	)
}

func (s *ScriptGenerationService) polishContentWithSkill(userID uint, content string, model string, skillName string, detailPrefix string) (string, string, error) {
	skillID, skill := resolveScriptPolishSkill(skillName)
	isEN := s.promptI18n != nil && s.promptI18n.IsEnglish()

	systemPrompt := skill.SystemPromptZH
	userTemplate := skill.UserPromptZH
	if isEN {
		systemPrompt = skill.SystemPromptEN
		userTemplate = skill.UserPromptEN
	}
	userPrompt := fmt.Sprintf(userTemplate, content)

	modelHint := s.resolvePolishModelHint(model)

	client, actualModel, billingRefID, err := reserveTextClient(
		s.aiService,
		s.billing,
		userID,
		modelHint,
		fmt.Sprintf("%s:%s", detailPrefix, skillID),
	)
	if err != nil && strings.TrimSpace(model) == "" && modelHint != "" {
		// 自动解析到的平台模型不可用时，回退到系统默认选择逻辑，避免因单模型异常导致润色不可用。
		client, actualModel, billingRefID, err = reserveTextClient(
			s.aiService,
			s.billing,
			userID,
			"",
			fmt.Sprintf("%s:%s", detailPrefix, skillID),
		)
	}
	if err != nil {
		return "", "", err
	}

	success := false
	defer func() {
		if !success && billingRefID != "" {
			_ = s.billing.RefundAI(billingRefID)
		}
	}()

	polished, err := client.GenerateText(
		userPrompt,
		systemPrompt,
		ai.WithTemperature(0.45),
		ai.WithMaxTokens(2600),
	)
	if err != nil {
		return "", "", err
	}

	polished = normalizePolishedScript(polished)
	if polished == "" {
		return "", "", errors.New("polished content is empty")
	}

	s.log.Infow("Script polished",
		"user_id", userID,
		"skill_name", skillID,
		"model", actualModel,
		"length", len([]rune(polished)))

	success = true
	return polished, skillID, nil
}

func normalizePolishedScript(raw string) string {
	out := strings.TrimSpace(raw)
	out = strings.TrimPrefix(out, "```text")
	out = strings.TrimPrefix(out, "```markdown")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)

	prefixes := []string{
		"润色后：",
		"润色结果：",
		"Polished version:",
		"Refined text:",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(out), strings.ToLower(p)) {
			out = strings.TrimSpace(strings.TrimPrefix(out, p))
			break
		}
	}
	return strings.TrimSpace(out)
}

func (s *ScriptGenerationService) resolvePolishModelHint(requestModel string) string {
	if m := strings.TrimSpace(requestModel); m != "" {
		return m
	}

	// 管理端配置（平台级）优先：使用文本服务的默认模型。
	if cfg, err := s.aiService.GetDefaultConfig("text"); err == nil && len(cfg.Model) > 0 {
		if model := strings.TrimSpace(cfg.Model[0]); model != "" {
			return model
		}
	}

	// 若管理端未配置，使用约定默认模型。
	return defaultPolishTextModel
}
