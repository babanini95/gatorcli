package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (c *Config, err error) {
	configPath, err := getConfigFilePath()
	configContent, err := os.Open(configPath)
	if err != nil {
		return &Config{}, fmt.Errorf("can not get config because: %v", err)
	}
	defer configContent.Close()

	// Read the config json file and decode to Config struct
	decoder := json.NewDecoder(configContent)
	err = decoder.Decode(&c)
	if err != nil {
		return &Config{}, fmt.Errorf("can not read config because: %v", err)
	}

	return c, nil
}

func (c *Config) SetUser(userName string) error {
	// Update config
	c.CurrentUserName = userName

	// Write json
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("can't get config file path because: %v", err)
	}

	configFile, _ := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer configFile.Close()
	encoder := json.NewEncoder(configFile)
	encoder.Encode(&c)
	return nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("can not get the user's home directory because: %v", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}
