package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	System struct {
		BaseUrl string `yaml:"base_url" env:"SYSTEM_BASE_URL"`
	}
	Telegram struct {
		BotToken   string `yaml:"bot_token" env:"TELEGRAM_BOT_TOKEN"`
		NumWorkers int8   `yaml:"num_workers" env:"TELEGRAM_NUM_WORKERS" envDefault:"10"`
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
