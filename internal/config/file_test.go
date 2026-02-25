package config

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadUserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	in := UserConfig{
		AI: AIConfig{
			Provider: "deepseek",
			APIKey:   "sk-test",
			BaseURL:  "https://api.deepseek.com/v1",
			Model:    "deepseek-chat",
			Timeout:  20,
		},
		GitHub: GitHubConfig{
			Token: "ghp_testtoken",
		},
	}

	if err := SaveUserConfig(path, in); err != nil {
		t.Fatalf("save config: %v", err)
	}
	out, err := LoadUserConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if out.AI.Provider != in.AI.Provider {
		t.Fatalf("provider = %q, want %q", out.AI.Provider, in.AI.Provider)
	}
	if out.AI.APIKey != in.AI.APIKey {
		t.Fatalf("api_key = %q, want %q", out.AI.APIKey, in.AI.APIKey)
	}
	if out.AI.BaseURL != in.AI.BaseURL {
		t.Fatalf("base_url = %q, want %q", out.AI.BaseURL, in.AI.BaseURL)
	}
	if out.AI.Model != in.AI.Model {
		t.Fatalf("model = %q, want %q", out.AI.Model, in.AI.Model)
	}
	if out.AI.Timeout != in.AI.Timeout {
		t.Fatalf("timeout = %d, want %d", out.AI.Timeout, in.AI.Timeout)
	}
	if out.GitHub.Token != in.GitHub.Token {
		t.Fatalf("github.token = %q, want %q", out.GitHub.Token, in.GitHub.Token)
	}
}

func TestLoadUserConfigNotExistsShouldReturnEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	cfg, err := LoadUserConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.AI.Provider != "" || cfg.AI.APIKey != "" || cfg.AI.Timeout != 0 || cfg.GitHub.Token != "" {
		t.Fatalf("cfg = %#v, want empty", cfg)
	}
}
