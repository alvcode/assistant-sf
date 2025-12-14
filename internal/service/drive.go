package service

import (
	"assistant-sf/internal/dict"
	"assistant-sf/internal/dto"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
	url := fmt.Sprintf("%s/api/drive/files/%d/chunks-info", domain, structID)

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

func GetChunk(domain string, structID int, chunkNumber int) ([]byte, error) {
	return getChunkInternal(domain, structID, chunkNumber, false)
}

func getChunkInternal(domain string, structID int, chunkNumber int, retry bool) ([]byte, error) {
	url := fmt.Sprintf("%s/api/drive/files/%d/chunks/%d", domain, structID, chunkNumber)

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
		return getChunkInternal(domain, structID, chunkNumber, true)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return nil, errors.New("failed. bad response")
		}
		return nil, errors.New(er.Message)
	}

	return io.ReadAll(resp.Body)
}

func UpdateHash(domain string, structID int, hash string) error {
	return updateHashInternal(domain, structID, hash, false)
}

func updateHashInternal(domain string, structID int, hash string, retry bool) error {
	url := fmt.Sprintf("%s/api/drive/files/%d/sha256/%s", domain, structID, hash)

	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		return err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if retry {
			return errors.New("unauthorized after refresh")
		}
		if err := RefreshToken(domain); err != nil {
			return fmt.Errorf("refresh token failed: %w", err)
		}
		return updateHashInternal(domain, structID, hash, true)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return errors.New("failed. bad response")
		}
		return errors.New(er.Message)
	}
	return nil
}

func DeleteStruct(domain string, structID int) error {
	return deleteStructInternal(domain, structID, false)
}

func deleteStructInternal(domain string, structID int, retry bool) error {
	url := fmt.Sprintf("%s/api/drive/%d", domain, structID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if retry {
			return errors.New("unauthorized after refresh")
		}
		if err := RefreshToken(domain); err != nil {
			return fmt.Errorf("refresh token failed: %w", err)
		}
		return deleteStructInternal(domain, structID, true)
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return errors.New("failed. bad response")
		}
		return errors.New(er.Message)
	}
	return nil
}

func CreateDir(domain string, name string, parentID *int) ([]*dto.DriveTree, error) {
	return createDirInternal(domain, name, parentID, false)
}

func createDirInternal(domain string, name string, parentID *int, retry bool) ([]*dto.DriveTree, error) {
	url := fmt.Sprintf("%s/api/drive/directories", domain)

	requestBody := map[string]interface{}{
		"name":      name,
		"parent_id": nil,
	}
	if parentID != nil {
		requestBody["parent_id"] = *parentID
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

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
		return createDirInternal(domain, name, parentID, true)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
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

func UploadFile(domain string, filePath string, parentID *int) ([]*dto.DriveTree, error) {
	return uploadFileInternal(domain, filePath, parentID, false)
}

func uploadFileInternal(domain string, filePath string, parentID *int, retry bool) ([]*dto.DriveTree, error) {
	var url string
	if parentID == nil {
		url = fmt.Sprintf("%s/%s", domain, "api/drive/upload-file")
	} else {
		url = fmt.Sprintf("%s/%s?parentId=%d", domain, "api/drive/upload-file", *parentID)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return nil, err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
		return uploadFileInternal(domain, filePath, parentID, true)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
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

func UploadFileByChunks(domain string, filePath string, parentID *int) ([]*dto.DriveTree, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	newStructID, err := uploadChunkPrepare(domain, info.Name(), info.Size(), parentID, false)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, dict.ChunkSize)
	chunkIndex := 0

	for {
		n, err := file.Read(buf)
		if n > 0 {
			if err := uploadFileByChunksInternal(domain, buf[:n], parentID, newStructID, chunkIndex, false); err != nil {
				return nil, err
			}
			chunkIndex++
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			errDel := deleteStructInternal(domain, newStructID, false)
			if errDel != nil {
				return nil, fmt.Errorf("delete struct failed: %w; after error upload chunk: %w", errDel, err)
			}
			return nil, fmt.Errorf("upload chunk failed: %w", err)
		}
	}

	err = chunkEndInternal(domain, newStructID, false)
	if err != nil {
		errDel := deleteStructInternal(domain, newStructID, false)
		if errDel != nil {
			return nil, fmt.Errorf("delete struct failed: %w; after error upload chunk: %w", errDel, err)
		}
		return nil, fmt.Errorf("chunk end failed: %w", err)
	}

	tree, err := GetTree(domain, parentID)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

/** Завершение загрузки чанками */
func chunkEndInternal(domain string, structID int, retry bool) error {
	url := fmt.Sprintf("%s/%s", domain, "api/drive/chunk-end")

	requestBody := map[string]interface{}{
		"struct_id": structID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if retry {
			return errors.New("unauthorized after refresh")
		}
		if err := RefreshToken(domain); err != nil {
			return fmt.Errorf("refresh token failed: %w", err)
		}
		return chunkEndInternal(domain, structID, true)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return errors.New("failed. bad response")
		}
		return errors.New(er.Message)
	}

	return nil
}

/** Загрузка чанками */
func uploadFileByChunksInternal(domain string, chunk []byte, parentID *int, structID int, chunkNumber int, retry bool) error {
	var url string
	if parentID == nil {
		url = fmt.Sprintf(
			"%s/%s?structId=%d&chunkNumber=%d",
			domain,
			"api/drive/upload-chunk",
			structID,
			chunkNumber,
		)
	} else {
		url = fmt.Sprintf(
			"%s/%s?structId=%d&chunkNumber=%d&parentId=%d",
			domain,
			"api/drive/upload-chunk",
			structID,
			chunkNumber,
			*parentID,
		)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "chunk.bin")
	if err != nil {
		return err
	}

	if _, err = part.Write(chunk); err != nil {
		return err
	}

	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		if retry {
			return errors.New("unauthorized after refresh")
		}
		if err := RefreshToken(domain); err != nil {
			return fmt.Errorf("refresh token failed: %w", err)
		}
		return uploadFileByChunksInternal(domain, chunk, parentID, structID, chunkNumber, true)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
		var er dto.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
			return errors.New("failed. bad response")
		}
		return errors.New(er.Message)
	}

	return nil
}

/** Подготовка к загрузке чанками, получение structID */
func uploadChunkPrepare(domain string, filename string, size int64, parentID *int, retry bool) (int, error) {
	url := fmt.Sprintf("%s/%s", domain, "api/drive/chunk-prepare")

	requestBody := map[string]interface{}{
		"filename":  filename,
		"full_size": size,
		"parent_id": nil,
		"sha256":    nil,
	}
	if parentID != nil {
		requestBody["parent_id"] = *parentID
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, err
	}
	token, err := KeyringGetAuthToken()
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

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
		return uploadChunkPrepare(domain, filename, size, parentID, true)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
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

	var result *dto.ChunkPrepareResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	return result.StructID, nil
}
