package controller

import (
	"assistant-sf/internal/command"
	"assistant-sf/internal/service"
	"context"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func InitController(rootCmd *cobra.Command, ctx context.Context) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.InitRun()
		}})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Authorization",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.AuthRun(ctx)
		}})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "from-disk",
		Short: "Brings the folder to a cloud-like state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.FromDiskRun(ctx)
		}})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "to-disk",
		Short: "Brings the cloud into a folder-like state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.ToDiskRun(ctx)
		}})

	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Two-way synchronization",
		RunE: func(cmd *cobra.Command, args []string) error {
			headFlag, _ := cmd.Flags().GetString("head")

			stopSpinner := service.StartSpinner("Processing...")

			if headFlag != "server" && headFlag != "local" {
				stopSpinner()
				color.Red("The --head flag can take the following values: server, local")
				return nil
			}

			err := command.SyncRun(ctx, headFlag)
			if err != nil {
				stopSpinner()
				color.Red(err.Error())
			}
			return nil
		}}
	rootCmd.AddCommand(syncCommand)
	syncCommand.PersistentFlags().String("head", "", "If files exist in the folder and on the server, but they differ: with the \"server\" option, the files from the server will be downloaded and replace the local ones. With the \"local\" option, these files will be uploaded to the server.")
}
