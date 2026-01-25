package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

var globalConfig *config

func loadEnvs() (*config, error) {
	cfg := &config{
		Web: Web{
			ListenAddr: os.Getenv("WEB_LISTEN_ADDR"),
			BaseURL:    os.Getenv("WEB_BASE_URL"),
			SessionKey: os.Getenv("WEB_SESSION_KEY"),
		},
		Auth: Auth{
			Type: strings.ToLower(os.Getenv("AUTH_TYPE")),
		},
	}

	switch cfg.Auth.Type {
	case "basic_auth":
		cfg.Auth.Basic = &AuthBasic{
			Users: strings.Split(os.Getenv("AUTH_BASIC_USERS"), ","),
		}
	case "oidc":
		cfg.Auth.OIDC = &AuthOIDC{
			ClientID:     os.Getenv("AUTH_OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("AUTH_OIDC_CLIENT_SECRET"),
			Issuer:       os.Getenv("AUTH_OIDC_ISSUER"),
		}
	default:
		return nil, fmt.Errorf("unknown AUTH_TYPE '%s'", cfg.Auth.Type)
	}

	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("set defaults: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *config) error {
	validate := validator.New()
	validate.RegisterValidation("listenaddr", validateListenAddr)
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	return nil
}

// Load loads configuration from environment variables and validates it.
func Load() error {
	cfg, err := loadEnvs()
	if err != nil {
		return err
	}

	err = validateConfig(cfg)
	if err != nil {
		return err
	}

	globalConfig = cfg
	return nil
}

// GetConfig returns the loaded global config (loading it on first use).
func GetConfig() *config {
	if globalConfig == nil {
		if err := Load(); err != nil {
			panic(err)
		}
	}
	return globalConfig
}

// IsSecureConnection reports whether the configured base URL uses HTTPS.
func IsSecureConnection() bool {
	return strings.HasPrefix(GetConfig().Web.BaseURL, "https")
}
