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
}
