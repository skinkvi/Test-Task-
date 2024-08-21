package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
	SSLMode  string `yaml:"sslmode"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
}

func ParseConfig(file string) (*Config, error) {
	logrus.Debugf("Parsing config file: %s", file)

	data, err := os.ReadFile(file)
	if err != nil {
		logrus.Errorf("Error reading config file: %v", err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		logrus.Errorf("Error unmarshaling config file: %v", err)
		return nil, err
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logrus.Errorf("Error marshaling config: %v", err)
		return nil, err
	}
	logrus.Debugf("Parsed config: %s", configJSON)

	return &config, nil
}
