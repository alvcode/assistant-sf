package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/service"
	"context"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
)

func AuthRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	fmt.Println("Enter your Assistant account login and password")
	var login string
	fmt.Print("Login: ")
	_, err := fmt.Scanln(&login)
	if err != nil {
		return err
	}

	fmt.Print("Password: ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	fmt.Println()

	password := string(passBytes)

	_, err = service.Authentication(cnf.AssistantURL, login, password)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	color.Green("Authentication successful")
	return nil
}
