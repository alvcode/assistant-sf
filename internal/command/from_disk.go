package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/service"
	"context"
	"fmt"
	"github.com/fatih/color"
)

func FromDiskRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		color.Red("Folder does not exist. Create a folder and specify the path to it in the configuration")
		return nil
	}

	color.Green("folder exists")

	tree, err := service.GetTree(cnf.AssistantURL, nil)
	if err != nil {
		color.Red(err.Error())
		return nil
	}
	fmt.Println(tree)

	return nil
}
