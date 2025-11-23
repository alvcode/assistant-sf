package service

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/fatih/color"
	"io"
	"os"
)

func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			color.Red("cannot close file: %w", err)
		}
	}(f)

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
