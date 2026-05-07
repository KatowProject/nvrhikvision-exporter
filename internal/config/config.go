package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type NVRConfig struct {
	IP       string `yaml:"ip"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type Config struct {
	NVRs   []NVRConfig  `yaml:"nvrs"`
	Server ServerConfig `yaml:"server"`
}

func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	log.Printf("Loaded config from %s", filePath)
	return &config, nil
}

func LoadConfigFromFlags(ip, user, pass, port string) *Config {
	config := Config{
		NVRs: []NVRConfig{
			{
				IP:       ip,
				Username: user,
				Password: pass,
				Name:     "NVR-Legacy",
			},
		},
		Server: ServerConfig{
			Port: port,
		},
	}
	return &config
}
