package cli

import (
	"path/filepath"
	"testing"

	"td/internal/ai/openai"
	"td/internal/config"
)

func TestNewAIProviderFromEnvDeepSeekDefaults(t *testing.T) {
	cfg := config.Default()
	cfg.ConfigToml = filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("TD_AI_PROVIDER", "deepseek")
	t.Setenv("TD_AI_API_KEY", "sk-test")
	t.Setenv("TD_AI_BASE_URL", "")
	t.Setenv("TD_AI_MODEL", "")
	t.Setenv("TD_AI_TIMEOUT", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	provider := newAIProviderFromConfig(cfg)
	client, ok := provider.(*openai.Client)
	if !ok {
		t.Fatalf("provider type = %T, want *openai.Client", provider)
	}
	if client.Endpoint != defaultDeepSeekEndpoint {
		t.Fatalf("endpoint = %q, want %q", client.Endpoint, defaultDeepSeekEndpoint)
	}
	if client.Model != defaultDeepSeekModel {
		t.Fatalf("model = %q, want %q", client.Model, defaultDeepSeekModel)
	}
}

func TestNewAIProviderFromEnvShouldReturnNilWithoutAPIKey(t *testing.T) {
	cfg := config.Default()
	cfg.ConfigToml = filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("TD_AI_PROVIDER", "deepseek")
	t.Setenv("TD_AI_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	if provider := newAIProviderFromConfig(cfg); provider != nil {
		t.Fatalf("provider = %T, want nil", provider)
	}
}
