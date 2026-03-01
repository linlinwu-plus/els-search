package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	RateLimit    RateLimitConfig    `yaml:"rate_limit"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type ElasticsearchConfig struct {
	Hosts []string `yaml:"hosts"`
}

type RateLimitConfig struct {
	Global GlobalRateLimitConfig `yaml:"global"`
	Search SearchRateLimitConfig `yaml:"search"`
}

type GlobalRateLimitConfig struct {
	RPS int `yaml:"rps"`
}

type SearchRateLimitConfig struct {
	RPS   int `yaml:"rps"`
	Burst int `yaml:"burst"`
}

func Load() (*Config, error) {
	// 从环境变量读取配置文件路径
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = filepath.Join("config", "config.yaml")
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
