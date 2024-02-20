package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// GetDirectoriesInPath retrieves the list of directories in the specified path.
//
// It takes a path string as a parameter and returns a slice of strings.
func GetDirectoriesInPath(path string) (sites []string) {
	dirs, _ := os.ReadDir(path)
	for _, entry := range dirs {
		if entry.IsDir() && entry.Name()[0] != '.' {
			sites = append(sites, entry.Name())
		}
	}
	return
}

// GeneratePassword generates a password of the specified length.
//
// Parameter(s):
// length int - the desired length of the password
// Return type(s):
// string - the generated password
// error - an error, if any
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

// ReplaceDashWithUnderscore replaces dashes with underscores in the given string.
//
// s: a string to be processed
// string: the modified string with dashes replaced by underscores
func ReplaceDashWithUnderscore(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// ReplaceSpacesWithDashes replaces spaces with dashes in the given string.
//
// It takes a string parameter and returns a string.
func ReplaceSpacesWithDashes(s string) string {
	return strings.ReplaceAll(s, " ", "-")
}

// AppendToFile appends the given content to the specified file.
//
// fileName: the name of the file to which the content will be appended
// content: the content to be appended to the file
// error: returns an error if the operation fails
func AppendToFile(fileName, content string) error {
	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a writer that appends to the file
	writer := bufio.NewWriter(file)

	// Write the content followed by a newline
	_, err = fmt.Fprintln(writer, content)
	if err != nil {
		return err
	}

	// Flush the writer to ensure that the content is written to the file
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

// ValidateNotEmpty checks if the given string is not empty.
//
// s string
// error
func ValidateNotEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("search string cannot be empty")
	}
	return nil
}
