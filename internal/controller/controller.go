package controller

import (
	"assistant-sf/internal/command"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var head string

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

	//rootCmd.Flags().StringVar(
	//	&head,
	//	"head",
	//	"local",
	//	"Head mode: server | local",
	//)

	syncCommand := &cobra.Command{
		Use:   "sync",
		Short: "Two-way synchronization",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("run sync")
			headFlag, _ := cmd.Flags().GetString("head")

			if headFlag == "server" {
				fmt.Println("head:", headFlag)
			} else if headFlag == "local" {
				fmt.Println("head:", headFlag)
			} else {
				color.Red("unknown head flag")
			}

			return nil
		}}
	rootCmd.AddCommand(syncCommand)

	//syncCommand.Flags().Count("head-server", "If a file exists in the folder and on the server, but they are different, the file will be downloaded from the server and replace the local file.")
	//syncCommand.Flags().Count("head-local", "If a file exists in the folder and on the server, but they are different, the local file will be sent to the server.")
	//syncCommand.PersistentFlags().String("head", "", "If a file exists in the folder and on the server, but they are different, the file will be downloaded from the server and replace the local file.")
	//syncCommand.PersistentFlags().String("head", "", "If a file exists in the folder and on the server, but they are different, the local file will be sent to the server.")

	syncCommand.PersistentFlags().String("head", "", "If files exist in the folder and on the server, but they differ: with the \"server\" option, the files from the server will be downloaded and replace the local ones. With the \"local\" option, these files will be uploaded to the server.")
}
