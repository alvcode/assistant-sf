package service

import (
	"assistant-sf/internal/dto"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func Authentication(domain string, login string, password string) (*dto.LoginSuccessResponse, error) {
	url := domain + "/api/auth/login"

	reqBody := dto.LoginRequest{
		Login:    login,
		Password: password,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	// ошибка по статусу
	if resp.StatusCode != http.StatusOK {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return nil, errors.New("login failed. bad response")
		}
		return nil, errors.New(er.Message)
	}

	// успех
	var sr dto.LoginSuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	err = KeyringSaveTokens(sr.Token, sr.RefreshToken)
	if err != nil {
		return nil, errors.New("failed save token")
	}

	return &sr, nil
}
