package main

import (
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type ciConfig struct {
	NoPublish bool `yaml:"noPublish"`
}

func loadCiConfig(root string) (ciConfig, error) {
	configPath := path.Join(root, "ci/config.yaml")
	if !checkPathExists(configPath) {
		log.Printf("config not found at %s, using default", configPath)
		return ciConfig{}, nil
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return ciConfig{}, err
	}
	config := ciConfig{}
	err = yaml.Unmarshal(configData, &config)
	return config, err
}
