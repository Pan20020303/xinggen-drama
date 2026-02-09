package ai

// StreamCallback 流式输出回调函数类型
// chunk: 本次收到的文本片段
// totalChars: 累计收到的总字符数
// estimatedProgress: 估算的生成进度 (0.0 - 1.0)
type StreamCallback func(chunk string, totalChars int, estimatedProgress float64)

// AIClient 定义文本生成客户端接口
type AIClient interface {
	GenerateText(prompt string, systemPrompt string, options ...func(*ChatCompletionRequest)) (string, error)
	// GenerateTextStream 流式生成文本，通过 callback 实时返回生成内容
	GenerateTextStream(prompt string, systemPrompt string, callback StreamCallback, options ...func(*ChatCompletionRequest)) (string, error)
	GenerateImage(prompt string, size string, n int) ([]string, error)
	TestConnection() error
}
