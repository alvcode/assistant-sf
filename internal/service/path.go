package service

import (
	"assistant-sf/internal/dict"
	"os"
	"path/filepath"
	"runtime"
)

func GetAppPath() (string, error) {
	var path string
	switch runtime.GOOS {
	case "linux":
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(base, dict.AppName)

	case "windows":
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(base, dict.AppName)

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, "Library", "Application Support", dict.AppName)
	}

	return path, nil
}

func PathExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}
