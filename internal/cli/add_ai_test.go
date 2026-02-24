package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"td/internal/config"
)

func TestAddClipAIShouldUseConfiguredProvider(t *testing.T) {
	var (
		gotPath string
		gotAuth string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"title\":\"from-ai\"}"}}]}`)
	}))
	defer server.Close()

	t.Setenv("TD_AI_PROVIDER", "deepseek")
	t.Setenv("TD_AI_API_KEY", "sk-test")
	t.Setenv("TD_AI_BASE_URL", server.URL+"/v1")
	t.Setenv("TD_AI_MODEL", "deepseek-chat")

	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "add", "--clip", "--ai", "raw title from clip")
	if !strings.Contains(out, "from-ai") {
		t.Fatalf("add output = %q, want contains %q", out, "from-ai")
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("request path = %q, want %q", gotPath, "/v1/chat/completions")
	}
	if gotAuth != "Bearer sk-test" {
		t.Fatalf("authorization = %q, want %q", gotAuth, "Bearer sk-test")
	}
}

func TestAddClipAIShouldWriteDueAndTodo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"title\":\"完成在线推理功能\",\"due\":\"2026-02-25 13:00\"}"}}]}`)
	}))
	defer server.Close()

	t.Setenv("TD_AI_PROVIDER", "deepseek")
	t.Setenv("TD_AI_API_KEY", "sk-test")
	t.Setenv("TD_AI_BASE_URL", server.URL+"/v1")
	t.Setenv("TD_AI_MODEL", "deepseek-chat")

	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "add", "--clip", "--ai", "明天下午一点完成在线推理功能")
	matched := regexp.MustCompile(`created #(\d+)`).FindStringSubmatch(out)
	if len(matched) != 2 {
		t.Fatalf("cannot parse created id from output: %q", out)
	}
	id, err := strconv.ParseInt(matched[1], 10, 64)
	if err != nil {
		t.Fatalf("parse created id: %v", err)
	}

	ls := runCLI(t, cfg, "ls")
	if !strings.Contains(ls, strconv.FormatInt(id, 10)+"\t[todo]\t完成在线推理功能") {
		t.Fatalf("ls output = %q, want todo row", ls)
	}
	if !strings.Contains(ls, "2026-02-25 13:00") {
		t.Fatalf("ls output = %q, want due", ls)
	}
}
