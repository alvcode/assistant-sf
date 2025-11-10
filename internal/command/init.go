package command

import (
	"assistant-sf/internal/service"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
)

func InitRun() error {
	appPath, err := service.GetAppPath()
	if err != nil {
		return err
	}
	path := filepath.Join(appPath, "main.yaml")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte("assistant_url: \"https://you-domain.com\"\n"), 0o644); err != nil {
		return err
	}

	color.Green(fmt.Sprintf("%s %s \n\n%s", "Created:", path, "Please go to the configuration and specify the URL of your ASSISTANT server."))
	return nil
}
