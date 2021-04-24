package bors

import "github.com/pelletier/go-toml"

type BorsConfig struct {
	DeleteMergedBranches bool     `toml:"delete-merged-branches"`
	Timeout              int      `toml:"timeout-seconds"`
	Status               []string `toml:"status"`
}

func (b *BorsConfig) ApplyDefaults() {
	b.DeleteMergedBranches = true
}

func (b *BorsConfig) Serialize() ([]byte, error) {
	return toml.Marshal(b)
}

func (b *BorsConfig) AddJob(jobName string) {
	b.Status = append(b.Status, jobName)
}