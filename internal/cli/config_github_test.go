package cli

import (
	"strings"
	"testing"
)

func TestConfigGitHubSetGetUnsetShow(t *testing.T) {
	cfg := testConfigForAI(t)

	runCLI(t, cfg, "config", "github", "set", "token", "ghp_123456789")

	out := runCLI(t, cfg, "config", "github", "show")
	if !strings.Contains(out, "token: ghp****89") {
		t.Fatalf("show output = %q, want masked token", out)
	}
	if strings.Contains(out, "ghp_123456789") {
		t.Fatalf("show output should mask token, got %q", out)
	}

	got := strings.TrimSpace(runCLI(t, cfg, "config", "github", "get", "token"))
	if got != "ghp_123456789" {
		t.Fatalf("get token = %q, want %q", got, "ghp_123456789")
	}

	runCLI(t, cfg, "config", "github", "unset", "token")
	got = strings.TrimSpace(runCLI(t, cfg, "config", "github", "get", "token"))
	if got != "" {
		t.Fatalf("get token after unset = %q, want empty", got)
	}
}

func TestConfigGitHubInvalidKeyShouldFail(t *testing.T) {
	cfg := testConfigForAI(t)
	out, err := runCLIWithError(cfg, "config", "github", "set", "invalid", "x")
	if err == nil {
		t.Fatalf("set invalid key should fail, out=%q", out)
	}
	if !strings.Contains(out, "unsupported") {
		t.Fatalf("error output = %q, want contains unsupported", out)
	}
}
