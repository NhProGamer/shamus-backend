package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func LoadConfig() (Config, error) {
	exist, err := doesExistFile("config.yaml")
	if err != nil {
		return Config{}, err
	}
	if !exist {
		log.Println("No config file found ! Writing a config file...")
		err = SaveConfig(defaultConfig)
		if err != nil {
			return Config{}, err
		}
	}

	// Read YAML file
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return Config{}, err
	}

	// Unmarshal YAML data into Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
