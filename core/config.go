package core

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	VerificationPeriod time.Duration `yaml:"verification_period"`
	UpdatePeriod       time.Duration `yaml:"update_period"`
	Partitions         uint64        `yaml:"partitions"`
	Verifier           string        `yaml:"verifier"`
	AggHistory         bool          `yaml:"agg_history"`
	AggHistoryDepth    uint32        `yaml:"agg_history_depth"`
}

func ParseConfig(path string) (c Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c)
	return
}
