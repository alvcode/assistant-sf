package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/service"
	"context"
	"github.com/fatih/color"
)

func SyncRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		color.Red("Folder does not exist. Create a folder and specify the path to it in the configuration")
		return nil
	}

	color.Green("folder exists")

	err := syncRecursive(cnf.AssistantURL, cnf.FolderPath, nil)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	return nil
}

func syncRecursive(domain string, localPath string, parentID *int) error {

	return nil
}
