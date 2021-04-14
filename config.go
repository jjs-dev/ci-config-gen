package main

import (
	"gopkg.in/yaml.v2"
	"os"
	"path"
)

type ciConfig struct {
	Publish bool
}

func newDefaultCiConfig() ciConfig {
	return ciConfig{
		Publish: true,
	}
}

func loadCiConfig(root string) (ciConfig, error) {
	configPath := path.Join(root, "ci/config.yaml")
	if !checkPathExists(configPath) {
		return newDefaultCiConfig(), nil
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return ciConfig{}, err
	}
	config := ciConfig{}
	err = yaml.Unmarshal(configData, &config)
	return config, err
}
