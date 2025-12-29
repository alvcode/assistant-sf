package controller

import (
	"assistant-sf/internal/command"
	"assistant-sf/internal/service"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func InitController(rootCmd *cobra.Command, ctx context.Context) {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := command.InitRun()
			if err != nil {
				color.Red(err.Error())
			}
			return nil
		}}
	rootCmd.AddCommand(initCommand)

	authCommand := &cobra.Command{
		Use:   "auth",
		Short: "Authorization",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := command.AuthRun(ctx)
			if err != nil {
				color.Red(err.Error())
			}
			return nil
		}}
	rootCmd.AddCommand(authCommand)

	fromDiskCommand := &cobra.Command{
		Use:   "from-disk",
		Short: "Brings the folder to a cloud-like state",
		RunE: func(cmd *cobra.Command, args []string) error {
			debugFlag, _ := cmd.Flags().GetBool("debug")

			stopSpinner := func() {}
			if !debugFlag {
				stopSpinner = service.StartSpinner("Processing...")
			}

			err := command.FromDiskRun(ctx, debugFlag)
			if err != nil {
				if !debugFlag {
					stopSpinner()
				}
				color.Red(err.Error())
				return nil
			}
			if !debugFlag {
				stopSpinner()
			}
			return nil
		}}
	rootCmd.AddCommand(fromDiskCommand)
	fromDiskCommand.PersistentFlags().Bool("debug", false, "Enable debug mode")

	toDiskCommand := &cobra.Command{
		Use:   "to-disk",
		Short: "Brings the cloud into a folder-like state",
		RunE: func(cmd *cobra.Command, args []string) error {
			debugFlag, _ := cmd.Flags().GetBool("debug")

			stopSpinner := func() {}
			if !debugFlag {
				stopSpinner = service.StartSpinner("Processing...")
			}

			err := command.ToDiskRun(ctx, debugFlag)
			if err != nil {
				if !debugFlag {
					stopSpinner()
				}
				color.Red(err.Error())
				return nil
			}
			if !debugFlag {
				stopSpinner()
			}
			return nil
		}}
	rootCmd.AddCommand(toDiskCommand)
	toDiskCommand.PersistentFlags().Bool("debug", false, "Enable debug mode")

	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Two-way synchronization",
		RunE: func(cmd *cobra.Command, args []string) error {
			headFlag, _ := cmd.Flags().GetString("head")
			debugFlag, _ := cmd.Flags().GetBool("debug")

			if headFlag != "server" && headFlag != "local" {
				color.Red("The --head flag can take the following values: server, local")
				return nil
			}

			stopSpinner := func() {}
			if !debugFlag {
				stopSpinner = service.StartSpinner("Processing...")
			}

			err := command.SyncRun(ctx, headFlag, debugFlag)
			if err != nil {
				if !debugFlag {
					stopSpinner()
				}
				color.Red(err.Error())
				return nil
			}
			if !debugFlag {
				stopSpinner()
			}
			return nil
		}}
	rootCmd.AddCommand(syncCommand)
	syncCommand.PersistentFlags().String(
		"head",
		"",
		fmt.Sprintf(
			"%s %s",
			"If files exist in the folder and on the server, but they differ: with the \"server\" option, the files",
			"from the server will be downloaded and replace the local ones. With the \"local\" option, these files will be uploaded to the server.",
		),
	)
	syncCommand.PersistentFlags().Bool("debug", false, "Enable debug mode")
}
