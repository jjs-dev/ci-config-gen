package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
)

type ciConfig struct {
	NoPublish               bool     `yaml:"noPublish"`
	NoE2e                   bool     `yaml:"noE2e"`
	DockerImages            []string `yaml:"dockerImages"`
	BuildTimeout            int      `yaml:"buildTimeoutMinutes"`
	JobTimeout              int      `yaml:"jobTimeoutMinutes"`
	InternalHackForGenerator bool     `yaml:"internalHackForGenerator"`
}

func loadCiConfig(root string) (ciConfig, error) {
	configPath := path.Join(root, "ci/config.yaml")
	if !checkPathExists(configPath) {
		log.Fatalf("config not exists at %s", configPath)
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return ciConfig{}, err
	}
	config := ciConfig{}
	err = yaml.Unmarshal(configData, &config)

	if err != nil {
		return ciConfig{}, err
	}

	if !config.NoPublish {
		if len(config.DockerImages) == 0 {
			return ciConfig{}, fmt.Errorf("publish enabled, but no images listed")
		}
	}
	if config.BuildTimeout == 0 {
		return ciConfig{}, fmt.Errorf("build timeout not specified")
	}
	if config.JobTimeout == 0 {
		config.JobTimeout = config.BuildTimeout
	}

	return config, nil
}
