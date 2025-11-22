package service

import (
	"assistant-sf/internal/dto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func GetTree(domain string, parentID *int) ([]*dto.DriveTree, error) {
	return getTreeInternal(domain, parentID, false)
}

func getTreeInternal(domain string, parentID *int, retry bool) ([]*dto.DriveTree, error) {
	var url string
	if parentID == nil {
		url = fmt.Sprintf("%s/%s", domain, "api/drive/tree")
	} else {
		url = fmt.Sprintf("%s/%s?parentId=%d", domain, "api/drive/tree", *parentID)
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
		return getTreeInternal(domain, parentID, true)
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

func GetFullFile(domain string, structID int) ([]byte, error) {
	return getFullFileInternal(domain, structID, false)
}

func getFullFileInternal(domain string, structID int, retry bool) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%d", domain, "api/drive/files", structID)

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
		return getFullFileInternal(domain, structID, true)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return nil, errors.New("download failed. bad response")
		}
		return nil, errors.New(er.Message)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetMaxChunk(domain string, structID int) (int, error) {
	return getMaxChunkInternal(domain, structID, false)
}

func getMaxChunkInternal(domain string, structID int, retry bool) (int, error) {
	url := fmt.Sprintf("%s/%s/%d/chunks-info", domain, "api/drive/files", structID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if retry {
			return 0, errors.New("unauthorized after refresh")
		}
		if err := RefreshToken(domain); err != nil {
			return 0, fmt.Errorf("refresh token failed: %w", err)
		}
		return getMaxChunkInternal(domain, structID, true)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return 0, errors.New("failed. bad response")
		}
		return 0, errors.New(er.Message)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result dto.ChunkInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	return result.EndNumber, nil
}

/**
func getChunk(domain string, fileID, chunkNumber int) ([]byte, error) {
    url := fmt.Sprintf("%s/api/drive/files/%d/chunks/%d", domain, fileID, chunkNumber)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("error chunk %d: %s", chunkNumber, string(b))
    }

    return io.ReadAll(resp.Body)
}
*/
