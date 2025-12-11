package controller

import (
	"assistant-sf/internal/command"
	"context"
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
}
