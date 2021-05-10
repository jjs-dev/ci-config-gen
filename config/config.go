package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type CiConfig struct {
	NoPublish                bool     `yaml:"noPublish"`
	NoE2e                    bool     `yaml:"noE2e"`
	Codegen                  bool     `yaml:"codegen"`
	DockerImages             []string `yaml:"dockerImages"`
	BuildTimeout             int      `yaml:"buildTimeoutMinutes"`
	JobTimeout               int      `yaml:"jobTimeoutMinutes"`
	InternalHackForGenerator bool     `yaml:"internalHackForGenerator"`
}

func Load(root string) (CiConfig, error) {
	configPath := path.Join(root, "ci/config.yaml")
	_, err := os.Stat(configPath)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatalf("config not exists at %s", configPath)
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return CiConfig{}, err
	}
	config := CiConfig{}
	err = yaml.Unmarshal(configData, &config)

	if err != nil {
		return CiConfig{}, err
	}

	if !config.NoPublish {
		if len(config.DockerImages) == 0 {
			return CiConfig{}, fmt.Errorf("publish enabled, but no images listed")
		}
	}
	if config.BuildTimeout == 0 {
		return CiConfig{}, fmt.Errorf("build timeout not specified")
	}
	if config.JobTimeout == 0 {
		config.JobTimeout = config.BuildTimeout
	}

	return config, nil
}
