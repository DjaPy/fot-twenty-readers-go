package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	System struct {
		BaseUrl string `yaml:"baseUrl" env:"BASE_URL"`
	}
	Telegram struct {
		BotToken string `yaml:"botToken" env:"BOT_TOKEN"`
	}
}

func NewConfiguration() (*Config, error) {
	var cfg Config
	var err error
	if err = env.ParseWithOptions(&cfg, env.Options{Prefix: ""}); err != nil {
		return nil, fmt.Errorf("couldn't find conf in environment: %v", err)
	}
	return &cfg, nil
}
