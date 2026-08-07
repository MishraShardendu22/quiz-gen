package openrouter

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/go-resty/resty/v2"
)

const (
	DefaultModel   = "inclusionai/ling-3.0-flash:free"
	DefaultBaseURL = "https://openrouter.ai/api/v1"
	DefaultTimeout = 60 * time.Second
)

type Client struct {
	restyClient *resty.Client
	model       string
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

// returns the singleton OpenRouter client instance
func GetClient() *Client {
	once.Do(func() {
		model := DefaultModel
		apiKey := os.Getenv("OPENROUTER_API_KEY")

		rc := resty.New()
		rc.SetBaseURL(DefaultBaseURL)
		rc.SetTimeout(DefaultTimeout)
		if apiKey != "" {
			rc.SetHeader("Authorization", "Bearer "+apiKey)
		}
		rc.SetHeader("Content-Type", "application/json")

		// Resty retry configuration
		rc.SetRetryCount(3)
		rc.SetRetryWaitTime(2 * time.Second)
		rc.SetRetryMaxWaitTime(10 * time.Second)

		rc.AddRetryCondition(func(r *resty.Response, err error) bool {
			// Network error
			if err != nil {
				return true
			}
			if r == nil {
				return false
			}
			status := r.StatusCode()

			/*
				Retry ONLY on 429, 500, 502, 503, 504. (There could be a possible automatic fix for these error)
				- 429: Too Many Requests (rate limiting)
				- 500: Internal Server Error
				- 502: Bad Gateway
				- 503: Service Unavailable
				- 504: Gateway Timeout

				Do NOT retry 400, 401, 403, 404, 422.
				- 400: Bad Request (invalid request)
				- 401: Unauthorized (invalid API key)
				- 403: Forbidden (insufficient permissions)
				- 404: Not Found (invalid endpoint)
				- 422: Unprocessable Entity (invalid input data)
			*/ 
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
			model:       model,
		}
	})
	return instance
}

// sends a prompt to OpenRouter API and returns the generated content and usage
func (c *Client) GenerateQuestions(ctx context.Context, prompt string) (string, *Usage, error) {
	reqBody := ChatRequest{
		Model: c.model,
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
