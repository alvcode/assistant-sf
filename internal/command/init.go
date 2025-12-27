package command

import (
	"assistant-sf/internal/service"
	"fmt"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	AssistantURL string `yaml:"assistant_url"`
	FolderPath   string `yaml:"folder_path"`
}

func InitRun() error {
	appPath, err := service.GetAppPath()
	if err != nil {
		return fmt.Errorf("error get path application: %v", err)
	}
	path := filepath.Join(appPath, "main.yaml")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("error create directory: %v", err)
	}

	cfg := Config{
		AssistantURL: "https://you-domain.com",
		FolderPath:   "/path/to/sync/folder",
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshall config: %v", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("error write config: %v", err)
	}

	color.Green(fmt.Sprintf("%s %s \n\n%s", "Created:", path, "Please go to configuration and specify the settings"))
	return nil
}
