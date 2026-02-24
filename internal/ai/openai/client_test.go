package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"td/internal/ai"
)

func TestParseTaskShouldCallCompatibleAPIAndReturnMessageContent(t *testing.T) {
	var (
		gotPath  string
		gotAuth  string
		gotModel string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		defer r.Body.Close()

		var payload struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gotModel = payload.Model

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"title\":\"AI task\",\"priority\":\"P2\"}"}}]}`)
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL + "/v1",
		APIKey:     "sk-test",
		Model:      "deepseek-chat",
		HTTPClient: server.Client(),
	}

	raw, err := client.ParseTask(context.Background(), "buy milk")
	if err != nil {
		t.Fatalf("ParseTask error = %v, want nil", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q, want %q", gotPath, "/v1/chat/completions")
	}
	if gotAuth != "Bearer sk-test" {
		t.Fatalf("authorization = %q, want %q", gotAuth, "Bearer sk-test")
	}
	if gotModel != "deepseek-chat" {
		t.Fatalf("model = %q, want %q", gotModel, "deepseek-chat")
	}
	if strings.TrimSpace(raw) != `{"title":"AI task","priority":"P2"}` {
		t.Fatalf("raw = %q, want compact json content", raw)
	}
}

func TestParseTaskShouldHitCacheForSameInput(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"title\":\"cached\"}"}}]}`)
	}))
	defer server.Close()

	client := &Client{
		Endpoint:   server.URL + "/v1/chat/completions",
		APIKey:     "sk-test",
		Model:      "deepseek-chat",
		Cache:      ai.NewCache(),
		HTTPClient: server.Client(),
	}

	first, err := client.ParseTask(context.Background(), "same input")
	if err != nil {
		t.Fatalf("first ParseTask error = %v, want nil", err)
	}
	second, err := client.ParseTask(context.Background(), "same input")
	if err != nil {
		t.Fatalf("second ParseTask error = %v, want nil", err)
	}

	if first != second {
		t.Fatalf("cache mismatch, first=%q second=%q", first, second)
	}
	if hits != 1 {
		t.Fatalf("server hits = %d, want 1", hits)
	}
}
