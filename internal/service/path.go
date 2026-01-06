package service

import (
	"assistant-sf/internal/dict"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var ErrForbiddenPath = errors.New("a dangerous directory is specified in the folder_path setting in the main.yaml file. Please change it to avoid data loss")

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

func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

var (
	// Корень
	reFsRoot = regexp.MustCompile(`^([A-Za-z]:|/)$`)

	// Каталог пользователей
	reUsersRoot = regexp.MustCompile(`^(/home|/Users|[A-Za-z]:/Users)$`)

	// Домашняя папка
	reUserHome = regexp.MustCompile(
		`^(/home/[^/\\]+|/Users/[^/\\]+|[A-Za-z]:/Users/[^/\\]+)$`,
	)
)

func ValidateSyncPath(path string) error {
	if path == "" {
		return ErrForbiddenPath
	}

	p := normalizePath(path)

	fmt.Println("link: ", p)

	if reFsRoot.MatchString(p) {
		fmt.Println("yes root")
		return ErrForbiddenPath
	}

	if reUsersRoot.MatchString(p) {
		fmt.Println("yes user")
		return ErrForbiddenPath
	}

	if reUserHome.MatchString(p) {
		fmt.Println("yes home")
		return ErrForbiddenPath
	}

	return nil
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)

	p = strings.ReplaceAll(p, `\`, `/`)

	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}

	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
	}

	return p
}
