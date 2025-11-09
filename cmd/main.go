package main

import (
	"assistant-sf/internal/controller"
	"assistant-sf/internal/logging"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logging.NewLogger()
	ctx = logging.ContextWithLogger(ctx, logger)

	rootCmd := &cobra.Command{
		Use:   "ast-sf",
		Short: "CLI commands",
	}

	controller.InitController(rootCmd, ctx)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
