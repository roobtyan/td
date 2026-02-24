package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"td/internal/config"
)

type aiField string

const (
	aiFieldProvider aiField = "provider"
	aiFieldAPIKey   aiField = "api_key"
	aiFieldBaseURL  aiField = "base_url"
	aiFieldModel    aiField = "model"
	aiFieldTimeout  aiField = "timeout"
)

func newConfigCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage td config",
	}
	cmd.AddCommand(newConfigAICmd(cfg))
	return cmd
}

func newConfigAICmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Manage AI config",
	}
	cmd.AddCommand(newConfigAISetCmd(cfg))
	cmd.AddCommand(newConfigAIGetCmd(cfg))
	cmd.AddCommand(newConfigAIUnsetCmd(cfg))
	cmd.AddCommand(newConfigAIShowCmd(cfg))
	return cmd
}

func newConfigAISetCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set AI config item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			field, err := parseAIField(args[0])
			if err != nil {
				return err
			}
			userCfg, err := config.LoadUserConfig(cfg.ConfigToml)
			if err != nil {
				return err
			}
			if err := setAIField(&userCfg.AI, field, args[1]); err != nil {
				return err
			}
			if err := config.SaveUserConfig(cfg.ConfigToml, userCfg); err != nil {
				return err
			}
			cmd.Printf("saved ai.%s\n", field)
			return nil
		},
	}
}

func newConfigAIGetCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get AI config item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			field, err := parseAIField(args[0])
			if err != nil {
				return err
			}
			userCfg, err := config.LoadUserConfig(cfg.ConfigToml)
			if err != nil {
				return err
			}
			cmd.Println(getAIField(userCfg.AI, field))
			return nil
		},
	}
}

func newConfigAIUnsetCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset AI config item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			field, err := parseAIField(args[0])
			if err != nil {
				return err
			}
			userCfg, err := config.LoadUserConfig(cfg.ConfigToml)
			if err != nil {
				return err
			}
			unsetAIField(&userCfg.AI, field)
			if err := config.SaveUserConfig(cfg.ConfigToml, userCfg); err != nil {
				return err
			}
			cmd.Printf("unset ai.%s\n", field)
			return nil
		},
	}
}

func newConfigAIShowCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show AI config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			userCfg, err := config.LoadUserConfig(cfg.ConfigToml)
			if err != nil {
				return err
			}
			cmd.Printf("provider: %s\n", fallbackDash(userCfg.AI.Provider))
			cmd.Printf("api_key: %s\n", maskSecret(userCfg.AI.APIKey))
			cmd.Printf("base_url: %s\n", fallbackDash(userCfg.AI.BaseURL))
			cmd.Printf("model: %s\n", fallbackDash(userCfg.AI.Model))
			if userCfg.AI.Timeout > 0 {
				cmd.Printf("timeout: %d\n", userCfg.AI.Timeout)
			} else {
				cmd.Printf("timeout: -\n")
			}
			return nil
		},
	}
}

func parseAIField(raw string) (aiField, error) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.ReplaceAll(key, "-", "_")
	switch key {
	case "provider":
		return aiFieldProvider, nil
	case "api_key", "apikey", "key":
		return aiFieldAPIKey, nil
	case "base_url", "baseurl", "url", "endpoint":
		return aiFieldBaseURL, nil
	case "model":
		return aiFieldModel, nil
	case "timeout":
		return aiFieldTimeout, nil
	default:
		return "", fmt.Errorf("unsupported ai key: %s", raw)
	}
}

func setAIField(aiCfg *config.AIConfig, field aiField, value string) error {
	switch field {
	case aiFieldProvider:
		provider := strings.ToLower(strings.TrimSpace(value))
		if provider != "deepseek" && provider != "openai" {
			return fmt.Errorf("provider must be deepseek or openai")
		}
		aiCfg.Provider = provider
		if provider == "openai" {
			aiCfg.BaseURL = defaultOpenAIEndpoint
			aiCfg.Model = defaultOpenAIModel
		} else {
			aiCfg.BaseURL = defaultDeepSeekEndpoint
			aiCfg.Model = defaultDeepSeekModel
		}
	case aiFieldAPIKey:
		aiCfg.APIKey = strings.TrimSpace(value)
	case aiFieldBaseURL:
		aiCfg.BaseURL = strings.TrimSpace(value)
	case aiFieldModel:
		aiCfg.Model = strings.TrimSpace(value)
	case aiFieldTimeout:
		timeout, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || timeout <= 0 {
			return fmt.Errorf("timeout must be a positive integer")
		}
		aiCfg.Timeout = timeout
	default:
		return fmt.Errorf("unsupported ai key: %s", field)
	}
	return nil
}

func getAIField(aiCfg config.AIConfig, field aiField) string {
	switch field {
	case aiFieldProvider:
		return aiCfg.Provider
	case aiFieldAPIKey:
		return aiCfg.APIKey
	case aiFieldBaseURL:
		return aiCfg.BaseURL
	case aiFieldModel:
		return aiCfg.Model
	case aiFieldTimeout:
		if aiCfg.Timeout <= 0 {
			return ""
		}
		return strconv.Itoa(aiCfg.Timeout)
	default:
		return ""
	}
}

func unsetAIField(aiCfg *config.AIConfig, field aiField) {
	switch field {
	case aiFieldProvider:
		aiCfg.Provider = ""
	case aiFieldAPIKey:
		aiCfg.APIKey = ""
	case aiFieldBaseURL:
		aiCfg.BaseURL = ""
	case aiFieldModel:
		aiCfg.Model = ""
	case aiFieldTimeout:
		aiCfg.Timeout = 0
	}
}

func maskSecret(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "-"
	}
	if len(text) <= 6 {
		return "***"
	}
	return text[:3] + "****" + text[len(text)-2:]
}

func fallbackDash(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "-"
	}
	return text
}
