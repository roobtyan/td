package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const defaultAPIBaseURL = "https://api.github.com"

var semverPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

type Options struct {
	Owner          string
	Repo           string
	CurrentVersion string
	APIBaseURL     string
	TargetPath     string
	GOOS           string
	GOARCH         string
	HTTPClient     *http.Client
}

type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
	Updated        bool
	AssetName      string
	AssetURL       string
	ChecksumURL    string
}

type Updater struct {
	owner          string
	repo           string
	currentVersion string
	apiBaseURL     string
	targetPath     string
	goos           string
	goarch         string
	httpClient     *http.Client
}

type releaseResponse struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func New(opts Options) *Updater {
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	apiBaseURL := strings.TrimRight(opts.APIBaseURL, "/")
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}

	targetPath := opts.TargetPath
	if targetPath == "" {
		targetPath, _ = os.Executable()
	}

	goos := opts.GOOS
	if goos == "" {
		goos = runtimeGOOS()
	}

	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtimeGOARCH()
	}

	return &Updater{
		owner:          defaultString(opts.Owner, "roobtyan"),
		repo:           defaultString(opts.Repo, "td"),
		currentVersion: opts.CurrentVersion,
		apiBaseURL:     apiBaseURL,
		targetPath:     targetPath,
		goos:           goos,
		goarch:         goarch,
		httpClient:     client,
	}
}

func (u *Updater) Check(ctx context.Context) (CheckResult, error) {
	assetName, err := AssetName(u.goos, u.goarch)
	if err != nil {
		return CheckResult{}, err
	}

	release, err := u.fetchLatestRelease(ctx)
	if err != nil {
		return CheckResult{}, err
	}

	assetURL := ""
	checksumURL := ""
	for _, asset := range release.Assets {
		switch asset.Name {
		case assetName:
			assetURL = asset.BrowserDownloadURL
		case "sha256sum.txt":
			checksumURL = asset.BrowserDownloadURL
		}
	}
	if assetURL == "" {
		return CheckResult{}, fmt.Errorf("latest release missing asset %q", assetName)
	}

	return CheckResult{
		CurrentVersion: u.currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      IsNewer(u.currentVersion, release.TagName),
		AssetName:      assetName,
		AssetURL:       assetURL,
		ChecksumURL:    checksumURL,
	}, nil
}

func (u *Updater) Upgrade(ctx context.Context) (CheckResult, error) {
	checkResult, err := u.Check(ctx)
	if err != nil {
		return CheckResult{}, err
	}
	if !checkResult.HasUpdate {
		return checkResult, nil
	}

	binaryData, err := u.getBytes(ctx, checkResult.AssetURL)
	if err != nil {
		return CheckResult{}, err
	}

	if checkResult.ChecksumURL != "" {
		sumData, err := u.getBytes(ctx, checkResult.ChecksumURL)
		if err != nil {
			return CheckResult{}, err
		}
		if err := verifyChecksum(sumData, checkResult.AssetName, binaryData); err != nil {
			return CheckResult{}, err
		}
	}

	if err := writeFileAtomic(u.targetPath, binaryData); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return CheckResult{}, fmt.Errorf("no permission to replace %q: %w", u.targetPath, err)
		}
		return CheckResult{}, err
	}

	checkResult.Updated = true
	return checkResult, nil
}

func (u *Updater) fetchLatestRelease(ctx context.Context) (releaseResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", u.apiBaseURL, u.owner, u.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return releaseResponse{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "td-updater")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return releaseResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return releaseResponse{}, fmt.Errorf("fetch latest release failed: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return releaseResponse{}, err
	}
	if strings.TrimSpace(release.TagName) == "" {
		return releaseResponse{}, fmt.Errorf("latest release tag is empty")
	}
	return release, nil
}

func (u *Updater) getBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "td-updater")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("download failed: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}

func verifyChecksum(sumData []byte, assetName string, data []byte) error {
	want, ok := FindChecksum(sumData, assetName)
	if !ok {
		return fmt.Errorf("checksum for %q not found", assetName)
	}
	hash := sha256.Sum256(data)
	got := hex.EncodeToString(hash[:])
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum mismatch for %q", assetName)
	}
	return nil
}

func writeFileAtomic(targetPath string, data []byte) error {
	if strings.TrimSpace(targetPath) == "" {
		return fmt.Errorf("target path is empty")
	}

	mode := os.FileMode(0o755)
	if info, err := os.Stat(targetPath); err == nil {
		mode = info.Mode()
	}

	dir := filepath.Dir(targetPath)
	tmpFile, err := os.CreateTemp(dir, ".td-upgrade-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Chmod(mode); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, targetPath)
}

func AssetName(goos, goarch string) (string, error) {
	switch {
	case goos == "darwin" && goarch == "amd64":
		return "td-darwin-amd64", nil
	case goos == "darwin" && goarch == "arm64":
		return "td-darwin-arm64", nil
	case goos == "linux" && goarch == "amd64":
		return "td-linux-amd64", nil
	case goos == "linux" && goarch == "arm64":
		return "td-linux-arm64", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}
}

func FindChecksum(sumData []byte, assetName string) (string, bool) {
	lines := strings.Split(string(sumData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		baseName := strings.ReplaceAll(name, "\\", "/")
		if idx := strings.LastIndex(baseName, "/"); idx >= 0 {
			baseName = baseName[idx+1:]
		}
		if name == assetName || baseName == assetName {
			return fields[0], true
		}
	}
	return "", false
}

func IsNewer(current, latest string) bool {
	cur, okCur := parseSemver(current)
	lat, okLat := parseSemver(latest)
	if !okLat {
		return false
	}
	if !okCur {
		return true
	}

	for i := 0; i < len(cur); i++ {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	return false
}

func parseSemver(version string) ([3]int, bool) {
	match := semverPattern.FindStringSubmatch(strings.TrimSpace(version))
	if len(match) != 4 {
		return [3]int{}, false
	}

	var out [3]int
	for i := 1; i <= 3; i++ {
		value, err := strconv.Atoi(match[i])
		if err != nil {
			return [3]int{}, false
		}
		out[i-1] = value
	}
	return out, true
}

func runtimeGOOS() string {
	return runtime.GOOS
}

func runtimeGOARCH() string {
	return runtime.GOARCH
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
