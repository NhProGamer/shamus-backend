package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

func SaveConfig(config Config) error {

	// Marshal struct into YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// Write YAML data to file
	err = os.WriteFile("config.yaml", data, 0644)
	if err != nil {
		return err
	}
	return nil
}
