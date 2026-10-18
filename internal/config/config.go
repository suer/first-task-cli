package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	APIKey       string `json:"apiKey"`
	ProjectID    string `json:"projectId"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type Credentials struct {
	UID          string    `json:"uid"`
	Email        string    `json:"email"`
	IDToken      string    `json:"idToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

func Dir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "firsttask"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "firsttask"), nil
}

func Load() (Config, error) {
	var cfg Config
	dir, err := Dir()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config.json: %w", err)
		}
	}

	override(&cfg.APIKey, "FIRSTTASK_API_KEY")
	override(&cfg.ProjectID, "FIRSTTASK_PROJECT_ID")
	override(&cfg.ClientID, "FIRSTTASK_CLIENT_ID")
	override(&cfg.ClientSecret, "FIRSTTASK_CLIENT_SECRET")
	return cfg, nil
}

func override(dst *string, env string) {
	if v := os.Getenv(env); v != "" {
		*dst = v
	}
}

func (c Config) Validate() error {
	var missing []string
	if c.APIKey == "" {
		missing = append(missing, "apiKey")
	}
	if c.ProjectID == "" {
		missing = append(missing, "projectId")
	}
	if c.ClientID == "" {
		missing = append(missing, "clientId")
	}
	if c.ClientSecret == "" {
		missing = append(missing, "clientSecret")
	}
	if len(missing) > 0 {
		dir, _ := Dir()
		return fmt.Errorf("missing config %v: set them in %s or via FIRSTTASK_* env vars", missing, filepath.Join(dir, "config.json"))
	}
	return nil
}

func LoadCredentials() (Credentials, error) {
	var cred Credentials
	dir, err := Dir()
	if err != nil {
		return cred, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "credentials.json"))
	if errors.Is(err, os.ErrNotExist) {
		return cred, errors.New("not logged in: run `firsttask login`")
	}
	if err != nil {
		return cred, err
	}
	if err := json.Unmarshal(data, &cred); err != nil {
		return cred, fmt.Errorf("parse credentials.json: %w", err)
	}
	return cred, nil
}

func SaveCredentials(cred Credentials) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0o600)
}

func DeleteCredentials() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(dir, "credentials.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
