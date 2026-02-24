package cli

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"td/internal/ai"
	"td/internal/ai/openai"
	"td/internal/app/usecase"
	"td/internal/config"
)

const (
	defaultDeepSeekEndpoint = "https://api.deepseek.com/v1/chat/completions"
	defaultDeepSeekModel    = "deepseek-chat"
	defaultOpenAIEndpoint   = "https://api.openai.com/v1/chat/completions"
	defaultOpenAIModel      = "gpt-4o-mini"
)

func newAIParseTaskUseCase(cfg config.Config) *usecase.AIParseTaskUseCase {
	provider := newAIProviderFromConfig(cfg)
	if provider == nil {
		return nil
	}
	return &usecase.AIParseTaskUseCase{Provider: provider}
}

func newAIProviderFromConfig(cfg config.Config) ai.Provider {
	userCfg, _ := config.LoadUserConfig(cfg.ConfigToml)
	provider := strings.ToLower(strings.TrimSpace(userCfg.AI.Provider))
	providerFromEnv := false
	if fromEnv := strings.ToLower(strings.TrimSpace(os.Getenv("TD_AI_PROVIDER"))); fromEnv != "" {
		provider = fromEnv
		providerFromEnv = true
	}
	if provider == "" {
		switch {
		case strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")) != "":
			provider = "deepseek"
		case strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != "":
			provider = "openai"
		default:
			provider = "deepseek"
		}
	}

	apiKey := strings.TrimSpace(os.Getenv("TD_AI_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(userCfg.AI.APIKey)
	}
	if apiKey == "" {
		switch provider {
		case "openai":
			apiKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		default:
			apiKey = strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
		}
	}
	if apiKey == "" {
		return nil
	}

	endpoint := ""
	if !providerFromEnv {
		endpoint = strings.TrimSpace(userCfg.AI.BaseURL)
	}
	if fromEnv := strings.TrimSpace(os.Getenv("TD_AI_BASE_URL")); fromEnv != "" {
		endpoint = fromEnv
	}
	model := ""
	if !providerFromEnv {
		model = strings.TrimSpace(userCfg.AI.Model)
	}
	if fromEnv := strings.TrimSpace(os.Getenv("TD_AI_MODEL")); fromEnv != "" {
		model = fromEnv
	}
	switch provider {
	case "deepseek":
		if endpoint == "" {
			endpoint = defaultDeepSeekEndpoint
		}
		if model == "" {
			model = defaultDeepSeekModel
		}
	case "openai":
		if endpoint == "" {
			endpoint = defaultOpenAIEndpoint
		}
		if model == "" {
			model = defaultOpenAIModel
		}
	default:
		return nil
	}

	timeoutSec := 20
	if userCfg.AI.Timeout > 0 {
		timeoutSec = userCfg.AI.Timeout
	}
	if raw := strings.TrimSpace(os.Getenv("TD_AI_TIMEOUT")); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			timeoutSec = seconds
		}
	}
	timeout := time.Duration(timeoutSec) * time.Second

	return &openai.Client{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Model:    model,
		Cache:    ai.NewCache(),
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}
