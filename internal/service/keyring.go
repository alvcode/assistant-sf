package service

import (
	"assistant-sf/internal/dict"
	"github.com/zalando/go-keyring"
)

func KeyringSaveTokens(token string, refreshToken string) error {
	err := keyring.Set(dict.AppName, "auth_token", token)
	if err != nil {
		return err
	}
	err = keyring.Set(dict.AppName, "refresh_token", refreshToken)
	if err != nil {
		return err
	}
	return nil
}

func KeyringGetAuthToken() (string, error) {
	val, err := keyring.Get(dict.AppName, "auth_token")
	if err != nil {
		return "", err
	}
	return val, nil
}

func KeyringGetRefreshToken() (string, error) {
	val, err := keyring.Get(dict.AppName, "refresh_token")
	if err != nil {
		return "", err
	}
	return val, nil
}
