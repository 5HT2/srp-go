package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

var (
	hash = sha256.New()
)

func GetFileHash(path string) (string, error) {
	hash.Reset()

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
