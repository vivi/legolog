package core

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ExperimentConfig struct {
	WritesPerUpdatePeriod  uint64        `yaml:"writes_per_update_period"`
	ServerAddr             string        `yaml:"server_addr"`
	AuditorAddr            string        `yaml:"auditor_addr"`
	NumVerificationPeriods uint64        `yaml:"num_verification_periods"`
	TestThroughputDuration time.Duration `yaml:"test_throughput_duration"`
}

func ParseExperimentConfig(path string) (c ExperimentConfig, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c)
	return
}
