package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/service"
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
)

func AuthRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	color.Blue("Enter your Assistant account login and password")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Login: ")
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return fmt.Errorf("failed to read login")
	}
	login := scanner.Text()
	if login == "" {
		return fmt.Errorf("login cannot be empty")
	}

	fmt.Print("Password: ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	password := string(passBytes)
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	_, err = service.Authentication(cnf.AssistantURL, login, password)
	if err != nil {
		return err
	}

	color.Green("Authentication successful")
	return nil
}
