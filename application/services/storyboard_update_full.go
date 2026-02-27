package services

import (
	"fmt"
	"strconv"

	"github.com/drama-generator/backend/domain/models"
	"gorm.io/gorm"
)

// UpdateStoryboard 更新分镜的所有字段，并重新生成提示词
func (s *StoryboardService) UpdateStoryboard(storyboardID string, updates map[string]interface{}) error {
	// 查找分镜
	var storyboard models.Storyboard
	if err := s.db.First(&storyboard, storyboardID).Error; err != nil {
		return fmt.Errorf("storyboard not found: %w", err)
	}

	// 构建用于重新生成提示词的Storyboard结构
	sb := Storyboard{
		ShotNumber: storyboard.StoryboardNumber,
	}

	var (
		hasCharacterIDs bool
		characterIDs    []uint
	)

	// 从updates中提取字段并更新
	updateData := make(map[string]interface{})

	if val, ok := updates["title"].(string); ok && val != "" {
		updateData["title"] = val
		sb.Title = val
	}
	if val, ok := updates["shot_type"].(string); ok && val != "" {
		updateData["shot_type"] = val
		sb.ShotType = val
	}
	if val, ok := updates["angle"].(string); ok && val != "" {
		updateData["angle"] = val
		sb.Angle = val
	}
	if val, ok := updates["movement"].(string); ok && val != "" {
		updateData["movement"] = val
		sb.Movement = val
	}
	if val, ok := updates["location"].(string); ok && val != "" {
		updateData["location"] = val
		sb.Location = val
	}
	if val, ok := updates["time"].(string); ok && val != "" {
		updateData["time"] = val
		sb.Time = val
	}
	if val, ok := updates["action"].(string); ok && val != "" {
		updateData["action"] = val
		sb.Action = val
	}
	if val, ok := updates["dialogue"].(string); ok && val != "" {
		updateData["dialogue"] = val
		sb.Dialogue = val
	}
	if val, ok := updates["result"].(string); ok && val != "" {
		updateData["result"] = val
		sb.Result = val
	}
	if val, ok := updates["atmosphere"].(string); ok && val != "" {
		updateData["atmosphere"] = val
		sb.Atmosphere = val
	}
	if val, ok := updates["description"].(string); ok && val != "" {
		updateData["description"] = val
	}
	if val, ok := updates["video_prompt"].(string); ok {
		updateData["video_prompt"] = val
	}
	if val, ok := updates["bgm_prompt"].(string); ok && val != "" {
		updateData["bgm_prompt"] = val
		sb.BgmPrompt = val
	}
	if val, ok := updates["sound_effect"].(string); ok && val != "" {
		updateData["sound_effect"] = val
		sb.SoundEffect = val
	}
	if val, ok := updates["duration"].(float64); ok {
		updateData["duration"] = int(val)
		sb.Duration = int(val)
	}
	if val, ok := updates["scene_id"].(float64); ok {
		sceneID := uint(val)
		updateData["scene_id"] = sceneID
	}
	if val, ok := updates["character_ids"]; ok {
		ids, err := parseUintIDs(val)
		if err != nil {
			return fmt.Errorf("invalid character_ids: %w", err)
		}
		hasCharacterIDs = true
		characterIDs = ids
	}

	// 使用当前数据库值填充缺失字段（用于生成提示词）
	if sb.Title == "" && storyboard.Title != nil {
		sb.Title = *storyboard.Title
	}
	if sb.ShotType == "" && storyboard.ShotType != nil {
		sb.ShotType = *storyboard.ShotType
	}
	if sb.Angle == "" && storyboard.Angle != nil {
		sb.Angle = *storyboard.Angle
	}
	if sb.Movement == "" && storyboard.Movement != nil {
		sb.Movement = *storyboard.Movement
	}
	if sb.Location == "" && storyboard.Location != nil {
		sb.Location = *storyboard.Location
	}
	if sb.Time == "" && storyboard.Time != nil {
		sb.Time = *storyboard.Time
	}
	if sb.Action == "" && storyboard.Action != nil {
		sb.Action = *storyboard.Action
	}
	if sb.Dialogue == "" && storyboard.Dialogue != nil {
		sb.Dialogue = *storyboard.Dialogue
	}
	if sb.Result == "" && storyboard.Result != nil {
		sb.Result = *storyboard.Result
	}
	if sb.Atmosphere == "" && storyboard.Atmosphere != nil {
		sb.Atmosphere = *storyboard.Atmosphere
	}
	if sb.BgmPrompt == "" && storyboard.BgmPrompt != nil {
		sb.BgmPrompt = *storyboard.BgmPrompt
	}
	if sb.SoundEffect == "" && storyboard.SoundEffect != nil {
		sb.SoundEffect = *storyboard.SoundEffect
	}
	if sb.Duration == 0 {
		sb.Duration = storyboard.Duration
	}

	// video_prompt 优先使用前端显式传入值；未传入时自动重建。
	// image_prompt不自动更新，因为可能对应多张已生成的帧图片
	if _, hasManualVideoPrompt := updates["video_prompt"]; !hasManualVideoPrompt {
		videoPrompt := s.generateVideoPrompt(sb)
		updateData["video_prompt"] = videoPrompt
	}

	// 更新数据库与角色关联
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(updateData) > 0 {
			if err := tx.Model(&storyboard).Updates(updateData).Error; err != nil {
				return fmt.Errorf("failed to update storyboard: %w", err)
			}
		}

		if hasCharacterIDs {
			var characters []models.Character
			if len(characterIDs) > 0 {
				if err := tx.Where("id IN ?", characterIDs).Find(&characters).Error; err != nil {
					return fmt.Errorf("failed to load characters: %w", err)
				}
			}
			assoc := tx.Model(&storyboard).Association("Characters")
			if len(characterIDs) == 0 {
				if err := assoc.Clear(); err != nil {
					return fmt.Errorf("failed to clear storyboard characters: %w", err)
				}
			} else {
				if err := assoc.Replace(characters); err != nil {
					return fmt.Errorf("failed to replace storyboard characters: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.log.Infow("Storyboard updated successfully",
		"storyboard_id", storyboardID,
		"fields_updated", len(updateData),
		"has_character_ids", hasCharacterIDs,
		"character_ids_count", len(characterIDs))

	return nil
}

func parseUintIDs(value interface{}) ([]uint, error) {
	if value == nil {
		return []uint{}, nil
	}

	rawList, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected array")
	}

	ids := make([]uint, 0, len(rawList))
	seen := make(map[uint]struct{}, len(rawList))

	for _, raw := range rawList {
		var id uint64
		switch v := raw.(type) {
		case float64:
			if v < 0 || v != float64(uint64(v)) {
				return nil, fmt.Errorf("invalid id value: %v", v)
			}
			id = uint64(v)
		case int:
			if v < 0 {
				return nil, fmt.Errorf("invalid id value: %v", v)
			}
			id = uint64(v)
		case uint:
			id = uint64(v)
		case string:
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid id value: %v", v)
			}
			id = parsed
		default:
			return nil, fmt.Errorf("unsupported id type: %T", raw)
		}

		if id == 0 {
			continue
		}
		u := uint(id)
		if _, exists := seen[u]; exists {
			continue
		}
		seen[u] = struct{}{}
		ids = append(ids, u)
	}

	return ids, nil
}
