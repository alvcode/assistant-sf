package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/service"
	"context"
	"github.com/fatih/color"
)

func FromDiskRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		color.Red("Folder does not exist. Create a folder and specify the path to it in the configuration")
		return nil
	}

	color.Green("folder exists")
	return nil
}
