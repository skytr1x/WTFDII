package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/skytr1x/wtfdii/internal/models"
)

const (
	ConfigFileName = ".wtfdiirc"
)

func Load(projectPath string) (*models.Config, error) {
	configPath := filepath.Join(projectPath, ConfigFileName)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return models.DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Save(projectPath string, config *models.Config) error {
	configPath := filepath.Join(projectPath, ConfigFileName)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
