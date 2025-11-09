package config

import (
	"assistant-sf/internal/logging"
	"assistant-sf/internal/service"
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"path/filepath"
)

type Config struct {
	AssistantURL string `yaml:"assistant_url" env:"ASSISTANT_URL" env-required:"true"`
	AppName      string `yaml:"app_name" env:"APP_NAME" env-default:"ast-sync-folder"`
}

func MustLoad(ctx context.Context) *Config {
	var cfg Config

	appPath, err := service.GetAppPath()
	if err != nil {
		logging.GetLogger(ctx).Fatalln(err)
	}

	configPath := filepath.Join(appPath, "main.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logging.GetLogger(ctx).Fatal("config file main.yaml does not exist")
	}

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		logging.GetLogger(ctx).Fatalf("error reading main.yaml file: %s", err)
	}
	return &cfg
}
