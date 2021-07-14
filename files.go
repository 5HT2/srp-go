package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
)

var (
	hash = sha256.New()
)

func GetFileHash(path string) string {
	hash.Reset()

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := io.Copy(hash, f); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}
