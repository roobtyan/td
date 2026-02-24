package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"td/internal/ai/openai"
	"td/internal/config"
)

func TestConfigAISetGetUnsetShow(t *testing.T) {
	cfg := testConfigForAI(t)

	runCLI(t, cfg, "config", "ai", "set", "provider", "deepseek")
	runCLI(t, cfg, "config", "ai", "set", "api-key", "sk-123456")
	runCLI(t, cfg, "config", "ai", "set", "base-url", "https://api.deepseek.com/v1")
	runCLI(t, cfg, "config", "ai", "set", "model", "deepseek-chat")
	runCLI(t, cfg, "config", "ai", "set", "timeout", "30")

	out := runCLI(t, cfg, "config", "ai", "show")
	if !strings.Contains(out, "provider: deepseek") {
		t.Fatalf("show output = %q, want provider", out)
	}
	if !strings.Contains(out, "base_url: https://api.deepseek.com/v1") {
		t.Fatalf("show output = %q, want base_url", out)
	}
	if !strings.Contains(out, "model: deepseek-chat") {
		t.Fatalf("show output = %q, want model", out)
	}
	if !strings.Contains(out, "timeout: 30") {
		t.Fatalf("show output = %q, want timeout", out)
	}
	if strings.Contains(out, "sk-123456") {
		t.Fatalf("show output should mask api key, got %q", out)
	}

	gotModel := runCLI(t, cfg, "config", "ai", "get", "model")
	if strings.TrimSpace(gotModel) != "deepseek-chat" {
		t.Fatalf("get model = %q, want %q", gotModel, "deepseek-chat")
	}

	runCLI(t, cfg, "config", "ai", "unset", "model")
	gotModel = runCLI(t, cfg, "config", "ai", "get", "model")
	if strings.TrimSpace(gotModel) != "" {
		t.Fatalf("get model after unset = %q, want empty", gotModel)
	}
}

func TestConfigAISetInvalidProviderShouldFail(t *testing.T) {
	cfg := testConfigForAI(t)
	out, err := runCLIWithError(cfg, "config", "ai", "set", "provider", "invalid")
	if err == nil {
		t.Fatalf("set invalid provider should fail, out=%q", out)
	}
	if !strings.Contains(out, "provider") {
		t.Fatalf("error output = %q, want contains provider", out)
	}
}

func TestConfigAISetProviderShouldApplyDefaultBaseURLAndModel(t *testing.T) {
	cfg := testConfigForAI(t)

	runCLI(t, cfg, "config", "ai", "set", "base-url", "https://custom.example/v1")
	runCLI(t, cfg, "config", "ai", "set", "model", "custom-model")

	runCLI(t, cfg, "config", "ai", "set", "provider", "openai")
	openaiURL := strings.TrimSpace(runCLI(t, cfg, "config", "ai", "get", "base-url"))
	openaiModel := strings.TrimSpace(runCLI(t, cfg, "config", "ai", "get", "model"))
	if openaiURL != defaultOpenAIEndpoint {
		t.Fatalf("openai base-url = %q, want %q", openaiURL, defaultOpenAIEndpoint)
	}
	if openaiModel != defaultOpenAIModel {
		t.Fatalf("openai model = %q, want %q", openaiModel, defaultOpenAIModel)
	}

	runCLI(t, cfg, "config", "ai", "set", "provider", "deepseek")
	deepseekURL := strings.TrimSpace(runCLI(t, cfg, "config", "ai", "get", "base-url"))
	deepseekModel := strings.TrimSpace(runCLI(t, cfg, "config", "ai", "get", "model"))
	if deepseekURL != defaultDeepSeekEndpoint {
		t.Fatalf("deepseek base-url = %q, want %q", deepseekURL, defaultDeepSeekEndpoint)
	}
	if deepseekModel != defaultDeepSeekModel {
		t.Fatalf("deepseek model = %q, want %q", deepseekModel, defaultDeepSeekModel)
	}
}

func TestAIProviderShouldReadFromConfigAndEnvPriority(t *testing.T) {
	cfg := testConfigForAI(t)
	runCLI(t, cfg, "config", "ai", "set", "provider", "openai")
	runCLI(t, cfg, "config", "ai", "set", "api-key", "sk-openai")

	provider := newAIProviderFromConfig(cfg)
	client, ok := provider.(*openai.Client)
	if !ok {
		t.Fatalf("provider type = %T, want *openai.Client", provider)
	}
	if client.APIKey != "sk-openai" {
		t.Fatalf("api key = %q, want %q", client.APIKey, "sk-openai")
	}
	if client.Model != defaultOpenAIModel {
		t.Fatalf("model = %q, want %q", client.Model, defaultOpenAIModel)
	}

	t.Setenv("TD_AI_PROVIDER", "deepseek")
	t.Setenv("TD_AI_API_KEY", "sk-deepseek")
	provider = newAIProviderFromConfig(cfg)
	client, ok = provider.(*openai.Client)
	if !ok {
		t.Fatalf("env override provider type = %T, want *openai.Client", provider)
	}
	if client.APIKey != "sk-deepseek" {
		t.Fatalf("env api key = %q, want %q", client.APIKey, "sk-deepseek")
	}
	if client.Model != defaultDeepSeekModel {
		t.Fatalf("env model = %q, want %q", client.Model, defaultDeepSeekModel)
	}
}

func testConfigForAI(t *testing.T) config.Config {
	t.Helper()
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")
	cfg.ConfigToml = filepath.Join(tdHome, "config.toml")
	return cfg
}

func runCLIWithError(cfg config.Config, args ...string) (string, error) {
	cmd := NewRootCmd(cfg)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}
