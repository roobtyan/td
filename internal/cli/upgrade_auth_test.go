package cli

import "testing"

func TestResolveGitHubTokenShouldFallbackToConfig(t *testing.T) {
	cfg := testConfigForAI(t)
	runCLI(t, cfg, "config", "github", "set", "token", "cfg-token-123")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	got := resolveGitHubToken(cfg)
	if got != "cfg-token-123" {
		t.Fatalf("token = %q, want %q", got, "cfg-token-123")
	}
}

func TestResolveGitHubTokenShouldPreferGHToken(t *testing.T) {
	cfg := testConfigForAI(t)
	runCLI(t, cfg, "config", "github", "set", "token", "cfg-token-123")
	t.Setenv("GH_TOKEN", "gh-token-1")
	t.Setenv("GITHUB_TOKEN", "github-token-2")

	got := resolveGitHubToken(cfg)
	if got != "gh-token-1" {
		t.Fatalf("token = %q, want %q", got, "gh-token-1")
	}
}

func TestResolveGitHubTokenShouldFallbackToGitHubTokenEnv(t *testing.T) {
	cfg := testConfigForAI(t)
	runCLI(t, cfg, "config", "github", "set", "token", "cfg-token-123")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "github-token-2")

	got := resolveGitHubToken(cfg)
	if got != "github-token-2" {
		t.Fatalf("token = %q, want %q", got, "github-token-2")
	}
}
