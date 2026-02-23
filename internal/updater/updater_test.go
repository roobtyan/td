package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{name: "latest greater", current: "v0.1.0", latest: "v0.2.0", want: true},
		{name: "same version", current: "v0.2.0", latest: "v0.2.0", want: false},
		{name: "latest smaller", current: "v0.2.1", latest: "v0.2.0", want: false},
		{name: "current invalid latest valid", current: "dev", latest: "v0.2.0", want: true},
		{name: "latest invalid", current: "v0.2.0", latest: "latest", want: false},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := IsNewer(tt.current, tt.latest)
			if got != tt.want {
				t.Fatalf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "darwin", goarch: "amd64", want: "td-darwin-amd64"},
		{goos: "darwin", goarch: "arm64", want: "td-darwin-arm64"},
		{goos: "linux", goarch: "amd64", want: "td-linux-amd64"},
		{goos: "linux", goarch: "arm64", want: "td-linux-arm64"},
	}

	for _, tt := range cases {
		got, err := AssetName(tt.goos, tt.goarch)
		if err != nil {
			t.Fatalf("AssetName(%q, %q) error = %v", tt.goos, tt.goarch, err)
		}
		if got != tt.want {
			t.Fatalf("AssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestCheckDetectsUpdate(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/roobtyan/td/releases/latest" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `{
  "tag_name": "v0.2.0",
  "assets": [
    {"name": "td-darwin-arm64", "browser_download_url": "https://example.com/td-darwin-arm64"},
    {"name": "sha256sum.txt", "browser_download_url": "https://example.com/sha256sum.txt"}
  ]
}`)
	}))
	defer api.Close()

	u := New(Options{
		Owner:          "roobtyan",
		Repo:           "td",
		CurrentVersion: "v0.1.0",
		APIBaseURL:     api.URL,
		HTTPClient:     api.Client(),
		GOOS:           "darwin",
		GOARCH:         "arm64",
	})

	result, err := u.Check(context.Background())
	if err != nil {
		t.Fatalf("Check error = %v", err)
	}
	if !result.HasUpdate {
		t.Fatalf("result.HasUpdate = false, want true")
	}
	if result.LatestVersion != "v0.2.0" {
		t.Fatalf("result.LatestVersion = %q, want %q", result.LatestVersion, "v0.2.0")
	}
	if result.AssetName != "td-darwin-arm64" {
		t.Fatalf("result.AssetName = %q, want %q", result.AssetName, "td-darwin-arm64")
	}
}

func TestUpgradeReplacesBinary(t *testing.T) {
	newBinary := []byte("new-td-binary")
	sum := sha256.Sum256(newBinary)
	sumHex := hex.EncodeToString(sum[:])

	var api *httptest.Server
	api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/roobtyan/td/releases/latest":
			fmt.Fprintf(w, `{
  "tag_name": "v0.2.0",
  "assets": [
    {"name": "td-darwin-arm64", "browser_download_url": %q},
    {"name": "sha256sum.txt", "browser_download_url": %q}
  ]
}`, api.URL+"/download/td-darwin-arm64", api.URL+"/download/sha256sum.txt")
		case "/download/td-darwin-arm64":
			_, _ = w.Write(newBinary)
		case "/download/sha256sum.txt":
			_, _ = w.Write([]byte(sumHex + "  td-darwin-arm64\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()

	targetPath := filepath.Join(t.TempDir(), "td")
	if err := os.WriteFile(targetPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}

	u := New(Options{
		Owner:          "roobtyan",
		Repo:           "td",
		CurrentVersion: "v0.1.0",
		APIBaseURL:     api.URL,
		HTTPClient:     api.Client(),
		GOOS:           "darwin",
		GOARCH:         "arm64",
		TargetPath:     targetPath,
	})

	result, err := u.Upgrade(context.Background())
	if err != nil {
		t.Fatalf("Upgrade error = %v", err)
	}
	if !result.Updated {
		t.Fatalf("result.Updated = false, want true")
	}

	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("ReadFile target: %v", err)
	}
	if string(got) != string(newBinary) {
		t.Fatalf("target binary content mismatch, got %q, want %q", string(got), string(newBinary))
	}
}

func TestUpgradeFailsOnChecksumMismatch(t *testing.T) {
	var api *httptest.Server
	api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/roobtyan/td/releases/latest":
			fmt.Fprintf(w, `{
  "tag_name": "v0.2.0",
  "assets": [
    {"name": "td-darwin-arm64", "browser_download_url": %q},
    {"name": "sha256sum.txt", "browser_download_url": %q}
  ]
}`, api.URL+"/download/td-darwin-arm64", api.URL+"/download/sha256sum.txt")
		case "/download/td-darwin-arm64":
			_, _ = w.Write([]byte("new-td-binary"))
		case "/download/sha256sum.txt":
			_, _ = w.Write([]byte(strings.Repeat("0", 64) + "  td-darwin-arm64\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()

	targetPath := filepath.Join(t.TempDir(), "td")
	if err := os.WriteFile(targetPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}

	u := New(Options{
		Owner:          "roobtyan",
		Repo:           "td",
		CurrentVersion: "v0.1.0",
		APIBaseURL:     api.URL,
		HTTPClient:     api.Client(),
		GOOS:           "darwin",
		GOARCH:         "arm64",
		TargetPath:     targetPath,
	})

	_, err := u.Upgrade(context.Background())
	if err == nil {
		t.Fatalf("Upgrade error = nil, want checksum mismatch error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "checksum") {
		t.Fatalf("Upgrade error = %v, want contains checksum", err)
	}
}
