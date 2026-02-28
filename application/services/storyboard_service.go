package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoryboardService struct {
	db          *gorm.DB
	aiService   *AIService
	taskService *TaskService
	billing     *BillingService
	log         *logger.Logger
	config      *config.Config
	promptI18n  *PromptI18n
}

func NewStoryboardService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *StoryboardService {
	return &StoryboardService{
		db:          db,
		aiService:   NewAIService(db, log),
		taskService: NewTaskService(db, log),
		billing:     NewBillingService(db, cfg, log),
		log:         log,
		config:      cfg,
		promptI18n:  NewPromptI18n(cfg),
	}
}

type Storyboard struct {
	ShotNumber  int    `json:"shot_number"`
	Title       string `json:"title"`        // 镜头标题
	ShotType    string `json:"shot_type"`    // 景别
	Angle       string `json:"angle"`        // 镜头角度
	Time        string `json:"time"`         // 时间
	Location    string `json:"location"`     // 地点
	SceneID     *uint  `json:"scene_id"`     // 背景ID（AI直接返回，可为null）
	Movement    string `json:"movement"`     // 运镜
	Action      string `json:"action"`       // 动作
	Dialogue    string `json:"dialogue"`     // 对话/独白
	Result      string `json:"result"`       // 画面结果
	Atmosphere  string `json:"atmosphere"`   // 环境氛围
	Emotion     string `json:"emotion"`      // 情绪
	Duration    int    `json:"duration"`     // 时长（秒）
	BgmPrompt   string `json:"bgm_prompt"`   // 配乐提示词
	SoundEffect string `json:"sound_effect"` // 音效描述
	Characters  []uint `json:"characters"`   // 涉及的角色ID列表
	IsPrimary   bool   `json:"is_primary"`   // 是否主镜
}

type GenerateStoryboardResult struct {
	Storyboards []Storyboard `json:"storyboards"`
	Total       int          `json:"total"`
}

type storyboardCharacterLink struct {
	StoryboardID uint `gorm:"column:storyboard_id"`
	CharacterID  uint `gorm:"column:character_id"`
}

type storyboardSegmentResult struct {
	Index       int
	Storyboards []Storyboard
	Err         error
}

const (
	storyboardMinMaxTokens       = 4000
	storyboardMaxMaxTokens       = 32000
	storyboardSegmentTargetRunes = 900
	storyboardSegmentMinRunes    = 450
	storyboardSegmentMaxRunes    = 1200
	storyboardSegmentConcurrency = 3
)

func estimateStoryboardMaxTokens(scriptLength int) int {
	if scriptLength <= 0 {
		return storyboardMinMaxTokens
	}

	estimatedShots := max(1, scriptLength/200)
	maxTokens := estimatedShots*350 + 1000
	if maxTokens < storyboardMinMaxTokens {
		return storyboardMinMaxTokens
	}
	if maxTokens > storyboardMaxMaxTokens {
		return storyboardMaxMaxTokens
	}
	return maxTokens
}

func uniqueUintSlice(ids []uint) []uint {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func optionalStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func requiredStringPtr(value string) *string {
	v := value
	return &v
}

func isStoryboardSceneBoundary(paragraph string) bool {
	trimmed := strings.TrimSpace(paragraph)
	if trimmed == "" {
		return false
	}

	sceneMarkers := []string{
		"【场景", "【镜头", "场景：", "场景:", "地点：", "地点:", "时间：", "时间:",
		"INT.", "EXT.", "内景", "外景",
	}
	for _, marker := range sceneMarkers {
		if strings.HasPrefix(trimmed, marker) {
			return true
		}
	}

	return strings.Contains(trimmed, "场景切换") || strings.Contains(trimmed, "转场")
}

func splitOversizedStoryboardParagraph(paragraph string, maxRunes int) []string {
	trimmed := strings.TrimSpace(paragraph)
	if trimmed == "" {
		return nil
	}
	if utf8.RuneCountInString(trimmed) <= maxRunes {
		return []string{trimmed}
	}

	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == '。' || r == '！' || r == '？' || r == '；' || r == '\n'
	})

	if len(parts) <= 1 {
		runes := []rune(trimmed)
		var chunks []string
		for start := 0; start < len(runes); start += maxRunes {
			end := min(start+maxRunes, len(runes))
			chunks = append(chunks, strings.TrimSpace(string(runes[start:end])))
		}
		return chunks
	}

	var chunks []string
	var builder strings.Builder
	currentRunes := 0

	flush := func() {
		chunk := strings.TrimSpace(builder.String())
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		builder.Reset()
		currentRunes = 0
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		partRunes := utf8.RuneCountInString(part)
		if currentRunes > 0 && currentRunes+partRunes+1 > maxRunes {
			flush()
		}
		if currentRunes > 0 {
			builder.WriteString("。")
		}
		builder.WriteString(part)
		currentRunes += partRunes + 1
	}
	flush()
	return chunks
}

func appendStoryboardParagraph(paragraphs []string, paragraph string) []string {
	return append(paragraphs, strings.TrimSpace(paragraph))
}

func joinStoryboardParagraphs(paragraphs []string) string {
	return strings.Join(paragraphs, "\n\n")
}

func (s *StoryboardService) splitScriptIntoSegments(scriptContent string) []string {
	normalized := strings.ReplaceAll(scriptContent, "\r\n", "\n")
	rawParagraphs := strings.Split(normalized, "\n\n")
	paragraphs := make([]string, 0, len(rawParagraphs))
	for _, paragraph := range rawParagraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		paragraphs = append(paragraphs, splitOversizedStoryboardParagraph(paragraph, storyboardSegmentTargetRunes)...)
	}

	if len(paragraphs) == 0 {
		trimmed := strings.TrimSpace(normalized)
		if trimmed == "" {
			return nil
		}
		return splitOversizedStoryboardParagraph(trimmed, storyboardSegmentMaxRunes)
	}

	segments := make([]string, 0)
	currentParagraphs := make([]string, 0)
	currentRunes := 0

	flush := func() {
		if len(currentParagraphs) == 0 {
			return
		}
		segments = append(segments, joinStoryboardParagraphs(currentParagraphs))
		currentParagraphs = currentParagraphs[:0]
		currentRunes = 0
	}

	for _, paragraph := range paragraphs {
		paragraphRunes := utf8.RuneCountInString(paragraph)
		shouldSplit := false
		if len(currentParagraphs) > 0 {
			if currentRunes+paragraphRunes > storyboardSegmentTargetRunes && currentRunes >= storyboardSegmentMinRunes {
				shouldSplit = true
			}
			if isStoryboardSceneBoundary(paragraph) && currentRunes >= storyboardSegmentMinRunes {
				shouldSplit = true
			}
			if currentRunes+paragraphRunes > storyboardSegmentMaxRunes {
				shouldSplit = true
			}
		}

		if shouldSplit {
			flush()
		}

		currentParagraphs = appendStoryboardParagraph(currentParagraphs, paragraph)
		currentRunes += paragraphRunes
	}
	flush()

	if len(segments) == 0 {
		return []string{strings.TrimSpace(normalized)}
	}
	return segments
}

func renumberStoryboards(storyboards []Storyboard, start int) {
	for i := range storyboards {
		storyboards[i].ShotNumber = start + i
	}
}

func maxConcurrentStoryboardSegments(total int) int {
	if total <= 1 {
		return 1
	}
	if total < storyboardSegmentConcurrency {
		return total
	}
	return storyboardSegmentConcurrency
}

func mergeStoryboardSegmentResults(results []storyboardSegmentResult) ([]Storyboard, error) {
	merged := make([]Storyboard, 0)
	for _, result := range results {
		if result.Err != nil {
			return nil, result.Err
		}
		renumberStoryboards(result.Storyboards, len(merged)+1)
		merged = append(merged, result.Storyboards...)
	}
	return merged, nil
}

func mergeAvailableStoryboardSegmentResults(results []storyboardSegmentResult) ([]Storyboard, error) {
	merged := make([]Storyboard, 0)
	for _, result := range results {
		if result.Err != nil {
			return nil, result.Err
		}
		if len(result.Storyboards) == 0 {
			continue
		}
		renumberStoryboards(result.Storyboards, len(merged)+1)
		merged = append(merged, result.Storyboards...)
	}
	return merged, nil
}

func buildPreviousScriptContext(segment string) string {
	trimmed := strings.TrimSpace(segment)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) > 280 {
		runes = runes[len(runes)-280:]
	}
	return strings.TrimSpace(string(runes))
}

func countReadyStoryboardSegments(ready []bool) int {
	count := 0
	for _, item := range ready {
		if !item {
			break
		}
		count++
	}
	return count
}

func buildStoryboardTaskResult(storyboards []Storyboard, readySegments, totalSegments int) gin.H {
	return gin.H{
		"storyboards":        storyboards,
		"total":              len(storyboards),
		"segments_completed": readySegments,
		"segment_total":      totalSegments,
		"is_partial":         readySegments < totalSegments,
	}
}

func storyboardGenerationPhaseMessage(totalSegments, completedSegments, concurrency int) string {
	if completedSegments <= 0 {
		return fmt.Sprintf("正在拆分剧本，共 %d 段，并发 %d 段生成", totalSegments, concurrency)
	}
	if completedSegments >= totalSegments {
		return fmt.Sprintf("分镜段已全部完成（%d/%d），准备合并结果", completedSegments, totalSegments)
	}
	return fmt.Sprintf("正在拆分分镜，已完成 %d/%d 段，并发 %d 段生成", completedSegments, totalSegments, concurrency)
}

func (s *StoryboardService) summarizeStoryboardContext(storyboards []Storyboard) string {
	if len(storyboards) == 0 {
		return ""
	}

	start := len(storyboards) - 2
	if start < 0 {
		start = 0
	}

	var lines []string
	for _, sb := range storyboards[start:] {
		lines = append(lines, fmt.Sprintf(
			"#%d %s | 场景:%s %s | 动作:%s | 结果:%s",
			sb.ShotNumber,
			strings.TrimSpace(sb.Title),
			strings.TrimSpace(sb.Location),
			strings.TrimSpace(sb.Time),
			strings.TrimSpace(sb.Action),
			strings.TrimSpace(sb.Result),
		))
	}
	return strings.Join(lines, "\n")
}

func (s *StoryboardService) buildStoryboardSegmentPrompt(scriptSegment, characterList, sceneList, previousContext string, segmentIndex, totalSegments int) string {
	scriptBody := scriptSegment
	if previousContext != "" {
		scriptBody = fmt.Sprintf("【上一段分镜摘要】\n%s\n\n【当前待拆分片段 %d/%d】\n%s", previousContext, segmentIndex, totalSegments, scriptSegment)
	}
	return s.buildStoryboardPrompt(scriptBody, characterList, sceneList)
}

func parseStoryboardGenerationResult(text string) (GenerateStoryboardResult, error) {
	var result GenerateStoryboardResult
	var storyboards []Storyboard
	if err := utils.SafeParseAIJSON(text, &storyboards); err == nil {
		result.Storyboards = storyboards
		result.Total = len(storyboards)
		return result, nil
	}
	if err := utils.SafeParseAIJSON(text, &result); err != nil {
		return GenerateStoryboardResult{}, err
	}
	result.Total = len(result.Storyboards)
	return result, nil
}

func (s *StoryboardService) prepareStoryboardGenerationClient(userID uint, model, episodeID string, billingRefID *string) (ai.AIClient, string, error) {
	requestedModel := model
	if requestedModel == "" {
		_, actualModel, cfgErr := s.aiService.GetBillingConfig("text", "", userID)
		if cfgErr != nil {
			return nil, "", cfgErr
		}
		requestedModel = actualModel
	}

	cfg, actualModel, cfgErr := s.aiService.GetBillingConfig("text", requestedModel, userID)
	if cfgErr != nil {
		return nil, "", cfgErr
	}

	client, err := s.aiService.GetAIClientForModelWithUser("text", actualModel, userID)
	if err != nil && model != "" {
		s.log.Warnw("Failed to get client for specified model, fallback to default model", "model", model, "error", err)
		cfg, actualModel, cfgErr = s.aiService.GetBillingConfig("text", "", userID)
		if cfgErr != nil {
			return nil, "", cfgErr
		}
		client, err = s.aiService.GetAIClientForModelWithUser("text", actualModel, userID)
	}
	if err != nil {
		return nil, "", err
	}

	refID, reserveErr := s.billing.ReserveAI(userID, "text", actualModel, cfg.CreditCost, "storyboard_generation:"+episodeID)
	if reserveErr != nil {
		return nil, "", reserveErr
	}
	*billingRefID = refID
	return client, actualModel, nil
}

func (s *StoryboardService) executeStoryboardSegmentsConcurrently(userID uint, taskID, actualModel, characterList, sceneList string, segments []string) ([]Storyboard, error) {
	totalSegments := len(segments)
	if totalSegments == 0 {
		return nil, fmt.Errorf("no storyboard segments")
	}

	concurrency := maxConcurrentStoryboardSegments(totalSegments)
	statusMsg := storyboardGenerationPhaseMessage(totalSegments, 0, concurrency)
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 10, statusMsg); err != nil {
		s.log.Warnw("Failed to update concurrent storyboard status", "error", err, "task_id", taskID)
	}

	sem := make(chan struct{}, concurrency)
	resultsCh := make(chan storyboardSegmentResult, totalSegments)
	var wg sync.WaitGroup

	for idx, segment := range segments {
		wg.Add(1)
		sem <- struct{}{}

		go func(index int, segmentText string) {
			defer wg.Done()
			defer func() { <-sem }()

			client, err := s.aiService.GetAIClientForModelWithUser("text", actualModel, userID)
			if err != nil {
				resultsCh <- storyboardSegmentResult{Index: index, Err: err}
				return
			}

			previousContext := ""
			if index > 0 {
				previousContext = buildPreviousScriptContext(segments[index-1])
			}
			maxTokens := estimateStoryboardMaxTokens(utf8.RuneCountInString(segmentText))
			prompt := s.buildStoryboardSegmentPrompt(segmentText, characterList, sceneList, previousContext, index+1, totalSegments)

			s.log.Infow("Generating storyboard segment concurrently",
				"task_id", taskID,
				"segment_index", index+1,
				"segment_total", totalSegments,
				"segment_length", utf8.RuneCountInString(segmentText),
				"prompt_length", len(prompt),
				"max_tokens", maxTokens)

			text, err := client.GenerateTextStream(prompt, "", nil, ai.WithMaxTokens(maxTokens))
			if err != nil {
				resultsCh <- storyboardSegmentResult{Index: index, Err: err}
				return
			}

			parsed, err := parseStoryboardGenerationResult(text)
			if err != nil {
				s.log.Errorw("Failed to parse concurrent storyboard segment",
					"error", err,
					"task_id", taskID,
					"segment_index", index+1,
					"response", text[:min(500, len(text))])
				resultsCh <- storyboardSegmentResult{Index: index, Err: err}
				return
			}

			resultsCh <- storyboardSegmentResult{
				Index:       index,
				Storyboards: parsed.Storyboards,
			}
		}(idx, segment)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]storyboardSegmentResult, totalSegments)
	completed := 0
	publishedCompleted := 0
	var firstErr error

	for result := range resultsCh {
		if result.Err != nil && firstErr == nil {
			firstErr = result.Err
		}
		results[result.Index] = result
		if result.Err == nil {
			completed++
			progress := 10 + int(float64(completed)*45/float64(totalSegments))
			if progress > 55 {
				progress = 55
			}
			message := storyboardGenerationPhaseMessage(totalSegments, completed, concurrency)
			if completed > publishedCompleted {
				previewStoryboards, mergeErr := mergeAvailableStoryboardSegmentResults(results)
				if mergeErr != nil {
					if firstErr == nil {
						firstErr = mergeErr
					}
				} else {
					previewMessage := message
					if len(previewStoryboards) > 0 {
						previewMessage = fmt.Sprintf("%s，已可预览 %d 个镜头", message, len(previewStoryboards))
					}
					if err := s.taskService.UpdateTaskProgressResult(
						taskID,
						"processing",
						progress,
						previewMessage,
						buildStoryboardTaskResult(previewStoryboards, completed, totalSegments),
					); err != nil {
						s.log.Warnw("Failed to update concurrent storyboard preview", "error", err, "task_id", taskID)
					}
					publishedCompleted = completed
					continue
				}
			}
			if err := s.taskService.UpdateTaskStatus(taskID, "processing", progress, message); err != nil {
				s.log.Warnw("Failed to update concurrent storyboard progress", "error", err, "task_id", taskID)
			}
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return mergeStoryboardSegmentResults(results)
}

func (s *StoryboardService) buildStoryboardPrompt(scriptContent, characterList, sceneList string) string {
	systemPrompt := s.promptI18n.GetStoryboardSystemPrompt()
	scriptLabel := s.promptI18n.FormatUserPrompt("script_content_label")
	taskLabel := s.promptI18n.FormatUserPrompt("task_label")
	taskInstruction := s.promptI18n.FormatUserPrompt("task_instruction")
	charListLabel := s.promptI18n.FormatUserPrompt("character_list_label")
	charConstraint := s.promptI18n.FormatUserPrompt("character_constraint")
	sceneListLabel := s.promptI18n.FormatUserPrompt("scene_list_label")
	sceneConstraint := s.promptI18n.FormatUserPrompt("scene_constraint")

	var builder strings.Builder
	builder.Grow(len(systemPrompt) + len(scriptContent) + len(characterList) + len(sceneList) + 1024)
	builder.WriteString(systemPrompt)
	builder.WriteString("\n\n")
	builder.WriteString(taskLabel)
	builder.WriteString("\n")
	builder.WriteString(taskInstruction)
	builder.WriteString("\n\n")
	builder.WriteString(charListLabel)
	builder.WriteString("\n")
	builder.WriteString(characterList)
	builder.WriteString("\n")
	builder.WriteString(charConstraint)
	builder.WriteString("\n\n")
	builder.WriteString(sceneListLabel)
	builder.WriteString("\n")
	builder.WriteString(sceneList)
	builder.WriteString("\n")
	builder.WriteString(sceneConstraint)
	builder.WriteString("\n\n")
	builder.WriteString(scriptLabel)
	builder.WriteString("\n")
	builder.WriteString(scriptContent)

	if s.promptI18n.IsEnglish() {
		builder.WriteString("\n\n[Output Contract]\n")
		builder.WriteString(`Return a JSON object: {"storyboards":[...]}. Each storyboard must include shot_number, title, shot_type, angle, time, location, scene_id, movement, action, dialogue, result, atmosphere, emotion, duration, bgm_prompt, sound_effect, characters, is_primary.`)
		builder.WriteString("\n- Keep one independent action per shot; do not merge beats.\n")
		builder.WriteString("- Keep all descriptive fields concrete and visual, but avoid filler.\n")
		builder.WriteString("- duration must be an integer between 4 and 12.\n")
		builder.WriteString("- characters must be an array of numeric ids from the provided list.\n")
		builder.WriteString("- scene_id must be a numeric id from the provided scene list or null.\n")
		builder.WriteString(`- dialogue must stay faithful to the script. Use "" when no dialogue exists.`)
		builder.WriteString("\n- Return JSON only, without markdown or explanations.")
	} else {
		builder.WriteString("\n\n【输出约束】\n")
		builder.WriteString(`返回 JSON 对象：{"storyboards":[...]}` + "\n")
		builder.WriteString("- 每个镜头只保留一个独立动作单元，不要合并剧情。\n")
		builder.WriteString("- 每个 storyboard 必须包含：shot_number、title、shot_type、angle、time、location、scene_id、movement、action、dialogue、result、atmosphere、emotion、duration、bgm_prompt、sound_effect、characters、is_primary。\n")
		builder.WriteString("- 描述字段要具体、可视化，但不要堆砌重复措辞。\n")
		builder.WriteString("- duration 必须是 4-12 的整数。\n")
		builder.WriteString("- characters 只能填写上方角色列表中的数字 ID。\n")
		builder.WriteString("- scene_id 只能填写上方场景列表中的数字 ID，没有匹配则填 null。\n")
		builder.WriteString("- dialogue 必须忠于原剧本，无对白时填写空字符串。\n")
		builder.WriteString("- 只返回 JSON，不要 markdown，不要解释。")
	}

	return builder.String()
}

func (s *StoryboardService) GenerateStoryboard(userID uint, episodeID string, model string) (string, error) {
	// 从数据库获取剧集信息
	var episode struct {
		ID            string
		ScriptContent *string
		Description   *string
		DramaID       string
	}

	err := s.db.Table("episodes").
		Select("episodes.id, episodes.script_content, episodes.description, episodes.drama_id").
		Joins("INNER JOIN dramas ON dramas.id = episodes.drama_id").
		Where("episodes.id = ? AND dramas.user_id = ?", episodeID, userID).
		First(&episode).Error

	if err != nil {
		return "", fmt.Errorf("剧集不存在或无权限访问")
	}

	// 获取剧本内容
	var scriptContent string
	if episode.ScriptContent != nil && *episode.ScriptContent != "" {
		scriptContent = *episode.ScriptContent
	} else if episode.Description != nil && *episode.Description != "" {
		scriptContent = *episode.Description
	} else {
		return "", fmt.Errorf("剧本内容为空，请先生成剧集内容")
	}

	// 获取该剧本的所有角色
	var characters []models.Character
	if err := s.db.Where("drama_id = ?", episode.DramaID).Order("name ASC").Find(&characters).Error; err != nil {
		return "", fmt.Errorf("获取角色列表失败: %w", err)
	}

	// 构建角色列表字符串（包含ID和名称）
	characterList := "无角色"
	if len(characters) > 0 {
		var charInfoList []string
		for _, char := range characters {
			charInfoList = append(charInfoList, fmt.Sprintf(`{"id": %d, "name": "%s"}`, char.ID, char.Name))
		}
		characterList = fmt.Sprintf("[%s]", strings.Join(charInfoList, ", "))
	}

	// 获取该项目已提取的场景列表（项目级）
	var scenes []models.Scene
	if err := s.db.Where("drama_id = ?", episode.DramaID).Order("location ASC, time ASC").Find(&scenes).Error; err != nil {
		s.log.Warnw("Failed to get scenes", "error", err)
	}

	// 构建场景列表字符串（包含ID、地点、时间）
	sceneList := "无场景"
	if len(scenes) > 0 {
		var sceneInfoList []string
		for _, bg := range scenes {
			sceneInfoList = append(sceneInfoList, fmt.Sprintf(`{"id": %d, "location": "%s", "time": "%s"}`, bg.ID, bg.Location, bg.Time))
		}
		sceneList = fmt.Sprintf("[%s]", strings.Join(sceneInfoList, ", "))
	}

	segmentCount := len(s.splitScriptIntoSegments(scriptContent))
	if segmentCount == 0 {
		segmentCount = 1
	}

	// 创建异步任务（若存在同资源进行中的任务则复用，避免重复扣分）
	task, created, err := s.taskService.CreateOrGetActiveTask("storyboard_generation", episodeID)
	if err != nil {
		s.log.Errorw("Failed to create task", "error", err)
		return "", fmt.Errorf("创建任务失败: %w", err)
	}
	if !created {
		s.log.Infow("Reusing active storyboard generation task", "task_id", task.ID, "episode_id", episodeID)
		return task.ID, nil
	}

	s.log.Infow("Generating storyboard asynchronously",
		"task_id", task.ID,
		"episode_id", episodeID,
		"drama_id", episode.DramaID,
		"script_length", len(scriptContent),
		"segment_count", segmentCount,
		"character_count", len(characters),
		"characters", characterList,
		"scene_count", len(scenes),
		"scenes", sceneList)

	// 启动后台goroutine处理AI调用和后续逻辑
	go s.processStoryboardGeneration(userID, task.ID, episodeID, model, scriptContent, characterList, sceneList)

	// 立即返回任务ID
	return task.ID, nil
}

// processStoryboardGeneration 后台处理故事板生成
func (s *StoryboardService) processStoryboardGeneration(userID uint, taskID, episodeID, model, scriptContent, characterList, sceneList string) {
	// 更新任务状态为处理中
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 10, "开始生成分镜头..."); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	billingRefID := ""
	success := false
	defer func() {
		if !success && billingRefID != "" {
			_ = s.billing.RefundAI(billingRefID)
		}
	}()
	fail := func(e error, userMessage string) {
		if updateErr := s.taskService.UpdateTaskError(taskID, fmt.Errorf(userMessage+": %w", e)); updateErr != nil {
			s.log.Errorw("Failed to update task error", "error", updateErr, "task_id", taskID)
		}
	}

	_, actualModel, reserveErr := s.prepareStoryboardGenerationClient(userID, model, episodeID, &billingRefID)
	if reserveErr != nil {
		fail(reserveErr, "生成分镜头失败")
		return
	}

	segments := s.splitScriptIntoSegments(scriptContent)
	if len(segments) == 0 {
		fail(fmt.Errorf("empty script segments"), "生成分镜头失败")
		return
	}

	s.log.Infow("Processing storyboard generation with concurrent segments",
		"task_id", taskID,
		"episode_id", episodeID,
		"segment_count", len(segments),
		"model", actualModel)

	allStoryboards, err := s.executeStoryboardSegmentsConcurrently(userID, taskID, actualModel, characterList, sceneList, segments)
	if err != nil {
		s.log.Errorw("Failed to generate storyboard segments concurrently", "error", err, "task_id", taskID)
		fail(err, "生成分镜头失败")
		return
	}

	result := GenerateStoryboardResult{
		Storyboards: allStoryboards,
		Total:       len(allStoryboards),
	}

	// 计算总时长（所有分镜时长之和）
	totalDuration := 0
	for _, sb := range result.Storyboards {
		totalDuration += sb.Duration
	}

	s.log.Infow("Storyboard generated",
		"task_id", taskID,
		"episode_id", episodeID,
		"count", result.Total,
		"total_duration_seconds", totalDuration)

	// 更新任务进度
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 70, fmt.Sprintf("分镜生成完成，共 %d 个镜头，正在保存到项目", result.Total)); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	// 保存分镜头到数据库
	if err := s.saveStoryboards(episodeID, result.Storyboards); err != nil {
		s.log.Errorw("Failed to save storyboards", "error", err, "task_id", taskID)
		fail(err, "保存分镜头失败")
		return
	}

	// 更新任务进度
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 90, fmt.Sprintf("分镜已保存，正在更新章节时长（总计 %d 秒）", totalDuration)); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	// 更新剧集时长（秒转分钟，向上取整）
	durationMinutes := (totalDuration + 59) / 60
	if err := s.db.Model(&models.Episode{}).Where("id = ?", episodeID).Update("duration", durationMinutes).Error; err != nil {
		s.log.Errorw("Failed to update episode duration", "error", err, "task_id", taskID)
		// 不中断流程，只记录错误
	} else {
		s.log.Infow("Episode duration updated",
			"task_id", taskID,
			"episode_id", episodeID,
			"duration_seconds", totalDuration,
			"duration_minutes", durationMinutes)
	}

	// 更新任务结果
	resultData := buildStoryboardTaskResult(result.Storyboards, len(segments), len(segments))
	resultData["total_duration"] = totalDuration
	resultData["duration_minutes"] = durationMinutes

	if err := s.taskService.UpdateTaskResult(taskID, resultData); err != nil {
		s.log.Errorw("Failed to update task result", "error", err, "task_id", taskID)
		return
	}

	success = true
	s.log.Infow("Storyboard generation completed", "task_id", taskID, "episode_id", episodeID)
}

// generateImagePrompt 生成专门用于图片生成的提示词（首帧静态画面）
func (s *StoryboardService) generateImagePrompt(sb Storyboard) string {
	var parts []string

	// 1. 完整的场景背景描述
	if sb.Location != "" {
		locationDesc := sb.Location
		if sb.Time != "" {
			locationDesc += ", " + sb.Time
		}
		parts = append(parts, locationDesc)
	}

	// 2. 角色初始静态姿态（去除动作过程，只保留起始状态）
	if sb.Action != "" {
		initialPose := extractInitialPose(sb.Action)
		if initialPose != "" {
			parts = append(parts, initialPose)
		}
	}

	// 3. 情绪氛围
	if sb.Emotion != "" {
		parts = append(parts, sb.Emotion)
	}

	// 4. 动漫风格
	parts = append(parts, "anime style, first frame")

	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	return "anime scene"
}

// extractInitialPose 提取初始静态姿态（去除动作过程）
func extractInitialPose(action string) string {
	// 去除动作过程关键词，保留初始状态描述
	processWords := []string{
		"然后", "接着", "接下来", "随后", "紧接着",
		"向下", "向上", "向前", "向后", "向左", "向右",
		"开始", "继续", "逐渐", "慢慢", "快速", "突然", "猛然",
	}

	result := action
	for _, word := range processWords {
		if idx := strings.Index(result, word); idx > 0 {
			// 在动作过程词之前截断
			result = result[:idx]
			break
		}
	}

	// 清理末尾标点
	result = strings.TrimRight(result, "，。,. ")
	return strings.TrimSpace(result)
}

// extractSimpleLocation 提取简化的场景地点（去除详细描述）
func extractSimpleLocation(location string) string {
	// 在"·"符号处截断，只保留主场景名称
	if idx := strings.Index(location, "·"); idx > 0 {
		return strings.TrimSpace(location[:idx])
	}

	// 如果有逗号，只保留第一部分
	if idx := strings.Index(location, "，"); idx > 0 {
		return strings.TrimSpace(location[:idx])
	}
	if idx := strings.Index(location, ","); idx > 0 {
		return strings.TrimSpace(location[:idx])
	}

	// 限制长度不超过15个字符
	maxLen := 15
	if len(location) > maxLen {
		return strings.TrimSpace(location[:maxLen])
	}

	return strings.TrimSpace(location)
}

// extractSimplePose 提取简单的核心姿态关键词（不超过10个字）
func extractSimplePose(action string) string {
	// 只提取前面最多10个字符作为核心姿态
	runes := []rune(action)
	maxLen := 10
	if len(runes) > maxLen {
		// 在标点符号处截断
		truncated := runes[:maxLen]
		for i := maxLen - 1; i >= 0; i-- {
			if truncated[i] == '，' || truncated[i] == '。' || truncated[i] == ',' || truncated[i] == '.' {
				truncated = runes[:i]
				break
			}
		}
		return strings.TrimSpace(string(truncated))
	}
	return strings.TrimSpace(action)
}

// extractFirstFramePose 从动作描述中提取首帧静态姿态
func extractFirstFramePose(action string) string {
	// 去除表示动作过程的关键词，保留初始状态
	processWords := []string{
		"然后", "接着", "向下", "向前", "走向", "冲向", "转身",
		"开始", "继续", "逐渐", "慢慢", "快速", "突然",
	}

	pose := action
	for _, word := range processWords {
		// 简单处理：在这些词之前截断
		if idx := strings.Index(pose, word); idx > 0 {
			pose = pose[:idx]
			break
		}
	}

	// 清理末尾标点
	pose = strings.TrimRight(pose, "，。,.")
	return strings.TrimSpace(pose)
}

// extractCompositionType 从镜头类型中提取构图类型（去除运镜）
func extractCompositionType(shotType string) string {
	// 去除运镜相关描述
	cameraMovements := []string{
		"晃动", "摇晃", "推进", "拉远", "跟随", "环绕",
		"运镜", "摄影", "移动", "旋转",
	}

	comp := shotType
	for _, movement := range cameraMovements {
		comp = strings.ReplaceAll(comp, movement, "")
	}

	// 清理多余的标点和空格
	comp = strings.ReplaceAll(comp, "··", "·")
	comp = strings.ReplaceAll(comp, "·", " ")
	comp = strings.TrimSpace(comp)

	return comp
}

// generateVideoPrompt 生成专门用于视频生成的提示词（包含运镜和动态元素）
func (s *StoryboardService) generateVideoPrompt(sb Storyboard) string {
	var parts []string
	videoRatio := "16:9"
	// 1. 人物动作
	if sb.Action != "" {
		parts = append(parts, fmt.Sprintf("Action: %s", sb.Action))
	}

	// 2. 对话
	if sb.Dialogue != "" {
		parts = append(parts, fmt.Sprintf("Dialogue: %s", sb.Dialogue))
	}

	// 3. 镜头运动（视频特有）
	if sb.Movement != "" {
		parts = append(parts, fmt.Sprintf("Camera movement: %s", sb.Movement))
	}

	// 4. 镜头类型和角度
	if sb.ShotType != "" {
		parts = append(parts, fmt.Sprintf("Shot type: %s", sb.ShotType))
	}
	if sb.Angle != "" {
		parts = append(parts, fmt.Sprintf("Camera angle: %s", sb.Angle))
	}

	// 5. 场景环境
	if sb.Location != "" {
		locationDesc := sb.Location
		if sb.Time != "" {
			locationDesc += ", " + sb.Time
		}
		parts = append(parts, fmt.Sprintf("Scene: %s", locationDesc))
	}

	// 6. 环境氛围
	if sb.Atmosphere != "" {
		parts = append(parts, fmt.Sprintf("Atmosphere: %s", sb.Atmosphere))
	}

	// 7. 情绪和结果
	if sb.Emotion != "" {
		parts = append(parts, fmt.Sprintf("Mood: %s", sb.Emotion))
	}
	if sb.Result != "" {
		parts = append(parts, fmt.Sprintf("Result: %s", sb.Result))
	}

	// 8. 音频元素
	if sb.BgmPrompt != "" {
		parts = append(parts, fmt.Sprintf("BGM: %s", sb.BgmPrompt))
	}
	if sb.SoundEffect != "" {
		parts = append(parts, fmt.Sprintf("Sound effects: %s", sb.SoundEffect))
	}

	// 9. 视频比例
	parts = append(parts, fmt.Sprintf("=VideoRatio: %s", videoRatio))
	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}
	return "Anime style video scene"
}

func (s *StoryboardService) saveStoryboards(episodeID string, storyboards []Storyboard) error {
	// 验证 episodeID
	epID, err := strconv.ParseUint(episodeID, 10, 32)
	if err != nil {
		s.log.Errorw("Invalid episode ID", "episode_id", episodeID, "error", err)
		return fmt.Errorf("无效的章节ID: %s", episodeID)
	}

	// 防御性检查：如果AI返回的分镜数量为0，不应该删除旧分镜
	if len(storyboards) == 0 {
		s.log.Errorw("AI返回的分镜数量为0，拒绝保存以避免删除现有分镜", "episode_id", episodeID)
		return fmt.Errorf("AI生成分镜失败：返回的分镜数量为0")
	}

	s.log.Infow("开始保存分镜头",
		"episode_id", episodeID,
		"episode_id_uint", uint(epID),
		"storyboard_count", len(storyboards))

	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 验证该章节是否存在
		var episode models.Episode
		if err := tx.First(&episode, epID).Error; err != nil {
			s.log.Errorw("Episode not found", "episode_id", episodeID, "error", err)
			return fmt.Errorf("章节不存在: %s", episodeID)
		}

		s.log.Infow("找到章节信息",
			"episode_id", episode.ID,
			"episode_number", episode.EpisodeNum,
			"drama_id", episode.DramaID,
			"title", episode.Title)

		// 获取该剧集所有的分镜ID（使用 uint 类型）
		var storyboardIDs []uint
		if err := tx.Model(&models.Storyboard{}).
			Where("episode_id = ?", uint(epID)).
			Pluck("id", &storyboardIDs).Error; err != nil {
			return err
		}

		s.log.Infow("查询到现有分镜",
			"episode_id_string", episodeID,
			"episode_id_uint", uint(epID),
			"existing_storyboard_count", len(storyboardIDs),
			"storyboard_ids", storyboardIDs)

		// 如果有分镜，先清理关联的image_generations的storyboard_id
		if len(storyboardIDs) > 0 {
			if err := tx.Model(&models.ImageGeneration{}).
				Where("storyboard_id IN ?", storyboardIDs).
				Update("storyboard_id", nil).Error; err != nil {
				return err
			}
			s.log.Infow("已清理关联的图片生成记录", "count", len(storyboardIDs))
		}

		// 删除该剧集已有的分镜头（使用 uint 类型确保类型匹配）
		s.log.Warnw("准备删除分镜数据",
			"episode_id_string", episodeID,
			"episode_id_uint", uint(epID),
			"episode_id_from_db", episode.ID,
			"will_delete_count", len(storyboardIDs))

		result := tx.Where("episode_id = ?", uint(epID)).Delete(&models.Storyboard{})
		if result.Error != nil {
			s.log.Errorw("删除旧分镜失败", "episode_id", uint(epID), "error", result.Error)
			return result.Error
		}

		s.log.Infow("已删除旧分镜头",
			"episode_id", uint(epID),
			"deleted_count", result.RowsAffected)

		// 注意：不删除背景，因为背景是在分镜拆解前就提取好的
		// AI会直接返回scene_id，不需要在这里做字符串匹配

		scenesToCreate := make([]models.Storyboard, 0, len(storyboards))
		storyboardCharacterIDs := make([][]uint, 0, len(storyboards))
		allCharacterIDs := make([]uint, 0)

		for _, sb := range storyboards {
			description := fmt.Sprintf("【镜头类型】%s\n【运镜】%s\n【动作】%s\n【对话】%s\n【结果】%s\n【情绪】%s",
				sb.ShotType, sb.Movement, sb.Action, sb.Dialogue, sb.Result, sb.Emotion)
			imagePrompt := s.generateImagePrompt(sb)
			videoPrompt := s.generateVideoPrompt(sb)
			characterIDs := uniqueUintSlice(sb.Characters)
			allCharacterIDs = append(allCharacterIDs, characterIDs...)

			if sb.SceneID != nil {
				s.log.Infow("Background ID from AI",
					"shot_number", sb.ShotNumber,
					"scene_id", *sb.SceneID)
			}

			scene := models.Storyboard{
				UserID:           episode.UserID,
				EpisodeID:        uint(epID),
				SceneID:          sb.SceneID,
				StoryboardNumber: sb.ShotNumber,
				Title:            optionalStringPtr(sb.Title),
				Location:         requiredStringPtr(sb.Location),
				Time:             requiredStringPtr(sb.Time),
				ShotType:         optionalStringPtr(sb.ShotType),
				Angle:            optionalStringPtr(sb.Angle),
				Movement:         optionalStringPtr(sb.Movement),
				Description:      requiredStringPtr(description),
				Action:           requiredStringPtr(sb.Action),
				Result:           optionalStringPtr(sb.Result),
				Atmosphere:       optionalStringPtr(sb.Atmosphere),
				Dialogue:         optionalStringPtr(sb.Dialogue),
				ImagePrompt:      requiredStringPtr(imagePrompt),
				VideoPrompt:      requiredStringPtr(videoPrompt),
				BgmPrompt:        optionalStringPtr(sb.BgmPrompt),
				SoundEffect:      optionalStringPtr(sb.SoundEffect),
				Duration:         sb.Duration,
			}

			scenesToCreate = append(scenesToCreate, scene)
			storyboardCharacterIDs = append(storyboardCharacterIDs, characterIDs)
		}

		if err := tx.CreateInBatches(&scenesToCreate, 50).Error; err != nil {
			s.log.Errorw("Failed to batch create storyboards", "error", err, "count", len(scenesToCreate))
			return err
		}

		validCharacterIDSet := make(map[uint]struct{})
		uniqueCharacterIDs := uniqueUintSlice(allCharacterIDs)
		if len(uniqueCharacterIDs) > 0 {
			var validCharacterIDs []uint
			if err := tx.Model(&models.Character{}).
				Where("drama_id = ? AND id IN ?", episode.DramaID, uniqueCharacterIDs).
				Pluck("id", &validCharacterIDs).Error; err != nil {
				s.log.Warnw("Failed to load characters for storyboard associations", "error", err, "character_ids", uniqueCharacterIDs)
			} else {
				for _, id := range validCharacterIDs {
					validCharacterIDSet[id] = struct{}{}
				}
			}
		}

		links := make([]storyboardCharacterLink, 0)
		for index, scene := range scenesToCreate {
			for _, characterID := range storyboardCharacterIDs[index] {
				if _, ok := validCharacterIDSet[characterID]; !ok {
					continue
				}
				links = append(links, storyboardCharacterLink{
					StoryboardID: scene.ID,
					CharacterID:  characterID,
				})
			}
		}

		if len(links) > 0 {
			if err := tx.Table("storyboard_characters").CreateInBatches(links, 200).Error; err != nil {
				s.log.Warnw("Failed to batch associate storyboard characters", "error", err, "count", len(links))
				return err
			}
		}

		s.log.Infow("Storyboards saved successfully", "episode_id", episodeID, "count", len(storyboards))
		return nil
	})
}

// CreateStoryboardRequest 创建分镜请求
type CreateStoryboardRequest struct {
	EpisodeID        uint    `json:"episode_id"`
	SceneID          *uint   `json:"scene_id"`
	StoryboardNumber int     `json:"storyboard_number"`
	Title            *string `json:"title"`
	Location         *string `json:"location"`
	Time             *string `json:"time"`
	ShotType         *string `json:"shot_type"`
	Angle            *string `json:"angle"`
	Movement         *string `json:"movement"`
	Description      *string `json:"description"`
	Action           *string `json:"action"`
	Result           *string `json:"result"`
	Atmosphere       *string `json:"atmosphere"`
	Dialogue         *string `json:"dialogue"`
	BgmPrompt        *string `json:"bgm_prompt"`
	SoundEffect      *string `json:"sound_effect"`
	Duration         int     `json:"duration"`
	Characters       []uint  `json:"characters"`
}

// CreateStoryboard 创建单个分镜
func (s *StoryboardService) CreateStoryboard(req *CreateStoryboardRequest) (*models.Storyboard, error) {
	var episode models.Episode
	if err := s.db.Where("id = ?", req.EpisodeID).First(&episode).Error; err != nil {
		return nil, fmt.Errorf("episode not found")
	}

	// 构建Storyboard对象
	sb := Storyboard{
		ShotNumber:  req.StoryboardNumber,
		ShotType:    getString(req.ShotType),
		Angle:       getString(req.Angle),
		Time:        getString(req.Time),
		Location:    getString(req.Location),
		SceneID:     req.SceneID,
		Movement:    getString(req.Movement),
		Action:      getString(req.Action),
		Dialogue:    getString(req.Dialogue),
		Result:      getString(req.Result),
		Atmosphere:  getString(req.Atmosphere),
		Emotion:     "", // 可以后续添加
		Duration:    req.Duration,
		BgmPrompt:   getString(req.BgmPrompt),
		SoundEffect: getString(req.SoundEffect),
		Characters:  req.Characters,
	}
	if req.Title != nil {
		sb.Title = *req.Title
	}

	// 生成提示词
	imagePrompt := s.generateImagePrompt(sb)
	videoPrompt := s.generateVideoPrompt(sb)

	// 构建 description
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	modelSB := &models.Storyboard{
		UserID:           episode.UserID,
		EpisodeID:        req.EpisodeID,
		SceneID:          req.SceneID,
		StoryboardNumber: req.StoryboardNumber,
		Title:            req.Title,
		Location:         req.Location,
		Time:             req.Time,
		ShotType:         req.ShotType,
		Angle:            req.Angle,
		Movement:         req.Movement,
		Description:      &desc,
		Action:           req.Action,
		Result:           req.Result,
		Atmosphere:       req.Atmosphere,
		Dialogue:         req.Dialogue,
		ImagePrompt:      &imagePrompt,
		VideoPrompt:      &videoPrompt,
		BgmPrompt:        req.BgmPrompt,
		SoundEffect:      req.SoundEffect,
		Duration:         req.Duration,
	}

	if err := s.db.Create(modelSB).Error; err != nil {
		return nil, fmt.Errorf("failed to create storyboard: %w", err)
	}

	// 关联角色
	if len(req.Characters) > 0 {
		var characters []models.Character
		if err := s.db.Where("id IN ?", req.Characters).Find(&characters).Error; err != nil {
			s.log.Warnw("Failed to find characters for new storyboard", "error", err)
		} else if len(characters) > 0 {
			s.db.Model(modelSB).Association("Characters").Append(characters)
		}
	}

	s.log.Infow("Storyboard created", "id", modelSB.ID, "episode_id", req.EpisodeID)
	return modelSB, nil
}

// DeleteStoryboard 删除分镜
func (s *StoryboardService) DeleteStoryboard(storyboardID uint) error {
	result := s.db.Where("id = ? ", storyboardID).Delete(&models.Storyboard{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("storyboard not found")
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
