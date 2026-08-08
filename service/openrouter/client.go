package openrouter

import (
	"context"
	"fmt"
	"sync"

	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	restyClient *resty.Client
}

var (
	instance *Client
	once     sync.Once
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatChoice struct {
	Message ChatMessage `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

type OpenRouterErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// GetClient returns the singleton OpenRouter client instance configured via util.Config.
// sync.Once is being used to make GetClient() initialize the OpenRouter client exactly once, even if multiple goroutines call it concurrently
// reuses the HTTP client's connection pool, which can reuse existing TCP/TLS connections when possible.
func GetClient() *Client {
	once.Do(func() {
		apiKey := util.Config.OpenRouterAPIKey
		baseURL := util.Config.OpenRouterBaseURL
		timeout := util.Config.HTTPTimeout

		rc := resty.New()
		rc.SetBaseURL(baseURL)
		rc.SetTimeout(timeout)
		if apiKey != "" {
			rc.SetHeader("Authorization", "Bearer "+apiKey)
		}
		rc.SetHeader("Content-Type", "application/json")

		// Resty retry configuration using util.Config
		rc.SetRetryCount(util.Config.RetryCount)
		rc.SetRetryWaitTime(util.Config.RetryWaitTime)
		rc.SetRetryMaxWaitTime(util.Config.RetryMaxWaitTime)

		rc.AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}
			if r == nil {
				return false
			}
			status := r.StatusCode()
			return status == 429 || status == 500 || status == 502 || status == 503 || status == 504
		})

		rc.AddRetryHook(func(r *resty.Response, err error) {
			attempt := 0
			status := 0
			reason := ""
			if r != nil {
				if r.Request != nil {
					attempt = r.Request.Attempt
				}
				status = r.StatusCode()
				reason = r.Status()
			}
			if err != nil {
				reason = err.Error()
			}
			util.Warn("OpenRouter HTTP retry", "attempt", attempt, "status", status, "reason", reason)
		})

		instance = &Client{
			restyClient: rc,
		}
	})
	return instance
}

// GetModel returns the configured model name from util.Config
func (c *Client) GetModel() string {
	return util.Config.ModelName
}

// GenerateQuestions sends a prompt to OpenRouter API
func (c *Client) GenerateQuestions(ctx context.Context, prompt string) (string, *Usage, error) {
	reqBody := ChatRequest{
		Model: c.GetModel(),
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	var result ChatResponse
	var errResp OpenRouterErrorResponse

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetResult(&result).
		SetError(&errResp).
		Post("/chat/completions")

	if err != nil {
		return "", nil, fmt.Errorf("openrouter request failed: %w", err)
	}

	if resp.IsError() {
		errMsg := errResp.Error.Message
		if errMsg == "" {
			errMsg = resp.Status()
		}
		return "", nil, fmt.Errorf("openrouter API error (status %d): %s", resp.StatusCode(), errMsg)
	}

	if len(result.Choices) == 0 {
		return "", nil, fmt.Errorf("openrouter returned empty choices")
	}

	return result.Choices[0].Message.Content, &result.Usage, nil
}
