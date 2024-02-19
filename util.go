package main

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strings"
)

// gets list of directories in path
func GetDirectoriesInPath(path string) (sites []string) {
	dirs, _ := os.ReadDir(path)
	for _, entry := range dirs {
		if entry.IsDir() && entry.Name()[0] != '.' {
			sites = append(sites, entry.Name())
		}
	}
	return
}

func GeneratePassword(length int) (string, error) {
	// Calculate the required byte length based on the desired password length
	byteLength := (length * 3) / 4

	// Generate random bytes
	randomBytes := make([]byte, byteLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode random bytes to base64 to get a readable password
	password := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Trim the password to the desired length
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}

func ReplaceDashWithUnderscore(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func ReplaceSpacesWithDashes(s string) string {
	return strings.ReplaceAll(s, " ", "-")
}
