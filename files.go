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

func ReadFileUnsafe(file string) string {
	content, err := os.ReadFile(file)

	if err != nil {
		panic(err)
	}

	return string(content)
}

func ReadDirsUnsafe(dirs ...string) map[string][]os.DirEntry {
	entries := make(map[string][]os.DirEntry, 0)

	for _, dir := range dirs {
		dirEntries, err := os.ReadDir(dir)

		if err != nil {
			panic(err)
		}

		entries[dir] = dirEntries
	}

	return entries
}
