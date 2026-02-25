package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AIConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  int
}

type GitHubConfig struct {
	Token string
}

type UserConfig struct {
	AI     AIConfig
	GitHub GitHubConfig
}

func LoadUserConfig(path string) (UserConfig, error) {
	var out UserConfig
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return out, nil
		}
		return out, err
	}

	section := ""
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return out, fmt.Errorf("invalid config line %d", lineNo)
		}

		key := normalizeConfigKey(parts[0])
		val := strings.TrimSpace(parts[1])
		switch section {
		case "ai":
			switch key {
			case "provider":
				out.AI.Provider = parseConfigString(val)
			case "api_key":
				out.AI.APIKey = parseConfigString(val)
			case "base_url":
				out.AI.BaseURL = parseConfigString(val)
			case "model":
				out.AI.Model = parseConfigString(val)
			case "timeout":
				raw := parseConfigString(val)
				if strings.TrimSpace(raw) == "" {
					out.AI.Timeout = 0
					continue
				}
				n, err := strconv.Atoi(raw)
				if err != nil {
					return out, fmt.Errorf("invalid ai.timeout at line %d", lineNo)
				}
				out.AI.Timeout = n
			}
		case "github":
			switch key {
			case "token":
				out.GitHub.Token = parseConfigString(val)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func SaveUserConfig(path string, cfg UserConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("[ai]\n")
	b.WriteString(`provider = ` + strconv.Quote(cfg.AI.Provider) + "\n")
	b.WriteString(`api_key = ` + strconv.Quote(cfg.AI.APIKey) + "\n")
	b.WriteString(`base_url = ` + strconv.Quote(cfg.AI.BaseURL) + "\n")
	b.WriteString(`model = ` + strconv.Quote(cfg.AI.Model) + "\n")
	if cfg.AI.Timeout > 0 {
		b.WriteString(fmt.Sprintf("timeout = %d\n", cfg.AI.Timeout))
	} else {
		b.WriteString("timeout = 0\n")
	}
	b.WriteString("\n")
	b.WriteString("[github]\n")
	b.WriteString(`token = ` + strconv.Quote(cfg.GitHub.Token) + "\n")

	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func normalizeConfigKey(raw string) string {
	text := strings.TrimSpace(strings.ToLower(raw))
	text = strings.ReplaceAll(text, "-", "_")
	return text
}

func parseConfigString(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	if strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`) {
		if unquoted, err := strconv.Unquote(text); err == nil {
			return unquoted
		}
	}
	if strings.HasPrefix(text, `'`) && strings.HasSuffix(text, `'`) && len(text) >= 2 {
		return text[1 : len(text)-1]
	}
	return text
}
