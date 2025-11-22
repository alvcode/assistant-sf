package service

import (
	"assistant-sf/internal/dto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const endpointDriveTree = "api/drive/tree"

func GetTree(domain string, parentId *int) ([]*dto.DriveTree, error) {
	return getTreeInternal(domain, parentId, false)
}

func getTreeInternal(domain string, parentId *int, retry bool) ([]*dto.DriveTree, error) {
	var url string
	if parentId == nil {
		url = fmt.Sprintf("%s/%s", domain, endpointDriveTree)
	} else {
		url = fmt.Sprintf("%s/%s?parentId=%d", domain, endpointDriveTree, parentId)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	token, err := KeyringGetAuthToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if retry {
			return nil, errors.New("unauthorized after refresh")
		}
		if err := RefreshToken(domain); err != nil {
			return nil, fmt.Errorf("refresh token failed: %w", err)
		}
		return getTreeInternal(domain, parentId, true)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return nil, errors.New("failed. bad response")
		}
		return nil, errors.New(er.Message)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []*dto.DriveTree
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
