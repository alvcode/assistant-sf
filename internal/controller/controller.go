package controller

import (
	"assistant-sf/internal/command"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"io"
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

			err := command.FromDiskRun(ctx, debugFlag)
			if err != nil {
				color.Red(err.Error())
				return nil
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

			err := command.ToDiskRun(ctx, debugFlag)
			if err != nil {
				color.Red(err.Error())
				return nil
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

			err := command.SyncRun(ctx, headFlag, debugFlag)
			if err != nil {
				color.Red(err.Error())
				return nil
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

	cryptCommand := &cobra.Command{
		Use:   "crypt",
		Short: "Two-way synchronization",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
			plaintext := []byte("exampleplaintext")

			block, err := aes.NewCipher(key)
			if err != nil {
				panic(err.Error())
			}

			aesgcm, err := cipher.NewGCM(block)
			if err != nil {
				panic(err.Error())
			}

			// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
			nonce := make([]byte, aesgcm.NonceSize())
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				panic(err.Error())
			}

			ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
			fmt.Printf("nonce: %x\n", nonce)
			fmt.Printf("ciphertext: %x\n", ciphertext)

			// ========================= decrypt

			key, _ = hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
			//ciphertext, _ = hex.DecodeString(string(ciphertext))
			//nonce, _ = hex.DecodeString(string(nonce))

			block, err = aes.NewCipher(key)
			if err != nil {
				panic(err.Error())
			}

			aesgcm, err = cipher.NewGCM(block)
			if err != nil {
				panic(err.Error())
			}

			plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("%s\n", plaintext)

			return nil
		}}
	rootCmd.AddCommand(cryptCommand)
}
