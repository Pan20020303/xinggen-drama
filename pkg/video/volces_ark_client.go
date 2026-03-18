package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/usage"
)

// VolcesArkClient 火山引擎ARK视频生成客户端
type VolcesArkClient struct {
	BaseURL       string
	APIKey        string
	Model         string
	Endpoint      string
	QueryEndpoint string
	HTTPClient    *http.Client
	lastUsage     usage.TokenUsage
}

func isSeedance15ProModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(m, "seedance-1-5-pro") ||
		strings.Contains(m, "seedance-1.5-pro") ||
		strings.Contains(m, "seedance1.5pro") ||
		strings.Contains(m, "seedance15pro")
}

func normalizeSeedance15ProModel(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "doubao-seedance-1-5-pro" || m == "seedance-1-5-pro" || m == "seedance-1.5-pro" {
		return "doubao-seedance-1-5-pro-251215"
	}
	return model
}

type VolcesArkContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	ImageURL map[string]interface{} `json:"image_url,omitempty"`
	Role     string                 `json:"role,omitempty"`
}

type VolcesArkRequest struct {
	Model         string             `json:"model"`
	TaskType      string             `json:"task_type,omitempty"`
	Content       []VolcesArkContent `json:"content"`
	GenerateAudio bool               `json:"generate_audio,omitempty"`
}

type VolcesArkResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Status  string `json:"status"`
	Content struct {
		VideoURL string `json:"video_url"`
	} `json:"content"`
	Usage struct {
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	CreatedAt             int64       `json:"created_at"`
	UpdatedAt             int64       `json:"updated_at"`
	Seed                  int         `json:"seed"`
	Resolution            string      `json:"resolution"`
	Ratio                 string      `json:"ratio"`
	Duration              int         `json:"duration"`
	FramesPerSecond       int         `json:"framespersecond"`
	ServiceTier           string      `json:"service_tier"`
	ExecutionExpiresAfter int         `json:"execution_expires_after"`
	GenerateAudio         bool        `json:"generate_audio"`
	Error                 interface{} `json:"error,omitempty"`
}

func NewVolcesArkClient(baseURL, apiKey, model, endpoint, queryEndpoint string) *VolcesArkClient {
	if endpoint == "" {
		endpoint = "/api/v3/contents/generations/tasks"
	}
	if queryEndpoint == "" {
		queryEndpoint = endpoint
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}

	return &VolcesArkClient{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Model:         model,
		Endpoint:      endpoint,
		QueryEndpoint: queryEndpoint,
		HTTPClient: &http.Client{
			Timeout:   300 * time.Second,
			Transport: transport,
		},
	}
}

func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}

	if netErr, ok := err.(net.Error); ok && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	msg := strings.ToLower(err.Error())
	retryableFragments := []string{
		"tls handshake timeout",
		"timeout awaiting response headers",
		"i/o timeout",
		"connection reset by peer",
		"unexpected eof",
		"eof",
	}

	for _, frag := range retryableFragments {
		if strings.Contains(msg, frag) {
			return true
		}
	}

	return false
}

func (c *VolcesArkClient) doRequestWithRetry(method, endpoint string, jsonBody []byte) (*http.Response, error) {
	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var bodyReader io.Reader
		if len(jsonBody) > 0 {
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequest(method, endpoint, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		if len(jsonBody) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)

		resp, err := c.HTTPClient.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt == maxAttempts || !isRetryableNetworkError(err) {
			break
		}

		backoff := time.Duration(attempt) * time.Second
		fmt.Printf("[VolcesARK] Request failed (attempt %d/%d): %v, retrying in %s\n", attempt, maxAttempts, err, backoff)
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("send request: %w", lastErr)
}

// GenerateVideo 生成视频（支持首帧、首尾帧、参考图等多种模式）
func (c *VolcesArkClient) GenerateVideo(imageURL, prompt string, opts ...VideoOption) (*VideoResult, error) {
	c.lastUsage = usage.TokenUsage{}
	options := &VideoOptions{
		Duration:    5,
		AspectRatio: "adaptive",
	}

	for _, opt := range opts {
		opt(options)
	}

	model := c.Model
	if options.Model != "" {
		model = options.Model
	}
	model = normalizeSeedance15ProModel(model)
	modelLower := strings.ToLower(model)

	// Seedance 1.5 Pro 文档约束：多图参考模式需 1-4 张参考图
	if isSeedance15ProModel(modelLower) && len(options.ReferenceImageURLs) > 4 {
		return nil, fmt.Errorf("seedance-1-5-pro supports 1-4 reference images, got %d", len(options.ReferenceImageURLs))
	}

	// 构建prompt文本（包含duration和ratio参数）
	promptText := prompt
	if options.AspectRatio != "" {
		promptText += fmt.Sprintf("  --ratio %s", options.AspectRatio)
	}
	if options.Duration > 0 {
		promptText += fmt.Sprintf("  --dur %d", options.Duration)
	}

	content := []VolcesArkContent{
		{
			Type: "text",
			Text: promptText,
		},
	}

	// 处理不同的图片模式
	// 1. 多图模式
	if len(options.ReferenceImageURLs) > 0 {
		for _, refURL := range options.ReferenceImageURLs {
			item := VolcesArkContent{
				Type: "image_url",
				ImageURL: map[string]interface{}{
					"url": refURL,
				},
				Role: "reference_image",
			}
			content = append(content, VolcesArkContent{
				Type:     item.Type,
				ImageURL: item.ImageURL,
				Role:     item.Role,
			})
		}
	} else if options.FirstFrameURL != "" && options.LastFrameURL != "" {
		// 2. 首尾帧模式
		content = append(content, VolcesArkContent{
			Type: "image_url",
			ImageURL: map[string]interface{}{
				"url": options.FirstFrameURL,
			},
			Role: "first_frame",
		})
		content = append(content, VolcesArkContent{
			Type: "image_url",
			ImageURL: map[string]interface{}{
				"url": options.LastFrameURL,
			},
			Role: "last_frame",
		})
	} else if imageURL != "" {
		// 3. 单图模式（默认）
		content = append(content, VolcesArkContent{
			Type: "image_url",
			ImageURL: map[string]interface{}{
				"url": imageURL,
			},
			// 单图模式不需要role
		})
	} else if options.FirstFrameURL != "" {
		// 4. 只有首帧
		content = append(content, VolcesArkContent{
			Type: "image_url",
			ImageURL: map[string]interface{}{
				"url": options.FirstFrameURL,
			},
			Role: "first_frame",
		})
	}

	// 只有 seedance-1-5-pro 模型支持 generate_audio 参数
	generateAudio := false
	if isSeedance15ProModel(modelLower) {
		generateAudio = true
	}

	reqBody := VolcesArkRequest{
		Model:         model,
		Content:       content,
		GenerateAudio: generateAudio,
	}
	if isSeedance15ProModel(modelLower) {
		// 对 seedance-1.5-pro 显式声明任务类型，避免网关按内容误判为 r2v。
		if len(content) > 1 {
			reqBody.TaskType = "i2v"
		} else {
			reqBody.TaskType = "t2v"
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := c.BaseURL + c.Endpoint
	fmt.Printf("[VolcesARK] Generating video - Endpoint: %s, FullURL: %s, Model: %s\n", c.Endpoint, endpoint, model)
	fmt.Printf("[VolcesARK] Request body: %s\n", string(jsonData))

	resp, err := c.doRequestWithRetry(http.MethodPost, endpoint, jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("[VolcesARK] Response status: %d, body: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// 兼容处理：部分网关在多图参考时会误判为 r2v（该模型不支持），
		// 这里自动降级为单图 i2v 重试一次，保证任务可继续执行。
		if isSeedance15ProModel(modelLower) &&
			len(options.ReferenceImageURLs) > 1 &&
			strings.Contains(string(body), "task_type r2v does not support model") {
			fallbackReq := VolcesArkRequest{
				Model: model,
				TaskType: "i2v",
				Content: []VolcesArkContent{
					{
						Type: "text",
						Text: promptText,
					},
					{
						Type: "image_url",
						ImageURL: map[string]interface{}{
							"url": options.ReferenceImageURLs[0],
						},
					},
				},
				GenerateAudio: generateAudio,
			}

			fallbackJSON, mErr := json.Marshal(fallbackReq)
			if mErr != nil {
				return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
			}

			fmt.Printf("[VolcesARK] Retrying with single-image i2v fallback for seedance-1-5-pro\n")
			fmt.Printf("[VolcesARK] Fallback request body: %s\n", string(fallbackJSON))

			retryResp, rErr := c.doRequestWithRetry(http.MethodPost, endpoint, fallbackJSON)
			if rErr != nil {
				return nil, fmt.Errorf("fallback request failed: %w", rErr)
			}
			defer retryResp.Body.Close()

			retryBody, rErr := io.ReadAll(retryResp.Body)
			if rErr != nil {
				return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
			}

			fmt.Printf("[VolcesARK] Fallback response status: %d, body: %s\n", retryResp.StatusCode, string(retryBody))

			if retryResp.StatusCode == http.StatusOK || retryResp.StatusCode == http.StatusCreated {
				body = retryBody
			} else {
				return nil, fmt.Errorf("API error (status %d): %s", retryResp.StatusCode, string(retryBody))
			}
		} else {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
	}

	var result VolcesArkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	c.lastUsage = usage.TokenUsage{
		PromptTokens:     0,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}

	fmt.Printf("[VolcesARK] Video generation initiated - TaskID: %s, Status: %s\n", result.ID, result.Status)

	if result.Error != nil {
		errorMsg := fmt.Sprintf("%v", result.Error)
		return nil, fmt.Errorf("volces error: %s", errorMsg)
	}

	videoResult := &VideoResult{
		TaskID:    result.ID,
		Status:    result.Status,
		Completed: result.Status == "completed" || result.Status == "succeeded",
		Duration:  result.Duration,
		Usage:     c.lastUsage,
	}

	if result.Content.VideoURL != "" {
		videoResult.VideoURL = result.Content.VideoURL
		videoResult.Completed = true
	}

	return videoResult, nil
}

func (c *VolcesArkClient) GetTaskStatus(taskID string) (*VideoResult, error) {
	// 替换占位符{taskId}、{task_id}或直接拼接
	queryPath := c.QueryEndpoint
	if strings.Contains(queryPath, "{taskId}") {
		queryPath = strings.ReplaceAll(queryPath, "{taskId}", taskID)
	} else if strings.Contains(queryPath, "{task_id}") {
		queryPath = strings.ReplaceAll(queryPath, "{task_id}", taskID)
	} else {
		queryPath = queryPath + "/" + taskID
	}

	endpoint := c.BaseURL + queryPath
	fmt.Printf("[VolcesARK] Querying task status - TaskID: %s, QueryEndpoint: %s, FullURL: %s\n", taskID, c.QueryEndpoint, endpoint)

	resp, err := c.doRequestWithRetry(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("[VolcesARK] Response body: %s\n", string(body))

	var result VolcesArkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	fmt.Printf("[VolcesARK] Parsed result - ID: %s, Status: %s, VideoURL: %s\n", result.ID, result.Status, result.Content.VideoURL)

	videoResult := &VideoResult{
		TaskID:    result.ID,
		Status:    result.Status,
		Completed: result.Status == "completed" || result.Status == "succeeded",
		Duration:  result.Duration,
		Usage: usage.TokenUsage{
			PromptTokens:     0,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
	}
	c.lastUsage = videoResult.Usage

	if result.Error != nil {
		videoResult.Error = fmt.Sprintf("%v", result.Error)
	}

	if result.Content.VideoURL != "" {
		videoResult.VideoURL = result.Content.VideoURL
		videoResult.Completed = true
	}

	return videoResult, nil
}

func (c *VolcesArkClient) GetLastUsage() usage.TokenUsage {
	return c.lastUsage
}
