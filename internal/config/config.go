package config

import (
	"os"
	"time"

	"go.yaml.in/yaml/v2"
)

type Config struct {
	Port        int           `yaml:"port"`
	TargetURL   string        `yaml:"target_url"`
	ZstdLevel   int           `yaml:"zstd_level"`
	GzipLevel   int           `yaml:"gzip_level"`
	Ttl         time.Duration `yaml:"ttl"`
	CacheType   string        `yaml:"cache_type"`
	MaxMemoryMB int           `yaml:"max_memory_mb"`
	LogLevel    string        `yaml:"log_level"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
