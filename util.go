package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/charmbracelet/huh/spinner"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// Check if new version is available and update + exit if it is
func CheckForUpdate() {
	var latest *selfupdate.Release
	var found bool
	var err error
	currentVersion := semver.MustParse(VERSION)
	spinner.New().Title("Checking for update...").Action(func() {
		latest, found, err = selfupdate.DetectLatest("BOOST-Creative/boost-server-cli")
	}).Run()
	checkError(err, "Failed to check for updates.")

	if !found || latest.Version.LTE(currentVersion) {
		return
	}

	printInBox(fmt.Sprintf("Update available: %s -> %s", VERSION, latest.Version))

	var binaryPath string
	spinner.New().Title(fmt.Sprintf("Updating to %s...", latest.Version)).Action(func() {
		binaryPath, err = os.Executable()
		checkError(err, "Could not locate executable path")
		err = selfupdate.UpdateTo(latest.AssetURL, binaryPath)
	}).Run()
	if err != nil {
		checkError(err, "Error occurred while updating binary:\n\n"+err.Error()+"\n\nIf the error is permission based, try running with sudo.")
	}
	printInBox(fmt.Sprintf("Successfully updated: %s -> %s\n\nRelease note:\n%s", VERSION, latest.Version, strings.TrimSpace(latest.ReleaseNotes)))
	os.Exit(0)
}

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

func DownloadFile(url, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the output file
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the contents to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// fmt.Printf("File downloaded to: %s\n", destination)
	return nil
}

func ReplaceTextInFile(filePath, oldText, newText string) error {
	// Read the content of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Perform the replacement
	newContent := strings.Replace(string(content), oldText, newText, -1)

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetHostsFromSSHConfig(configPath string, sudo bool) ([]string, error) {
	var content []byte
	var err error
	if sudo {
		content, err = exec.Command("sudo", "cat", configPath).Output()
	} else {
		content, err = os.ReadFile(configPath)
	}
	if err != nil {
		return nil, err
	}

	// Define a regular expression to match Host entries in the SSH config file
	re := regexp.MustCompile(`Host\s([a-zA-Z0-9\.\-]+)`)
	var hosts []string

	// Search for matches in the content
	matches := re.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		hosts = append(hosts, match[1])
	}

	return hosts, nil
}

// finds the last modified file with the specified extension in the given directory.
func FindLastModifiedFile(dirPath string, ext string) (fs.FileInfo, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var matchedFiles []os.FileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			info, err := file.Info()
			if err != nil {
				return nil, err
			}
			matchedFiles = append(matchedFiles, info)
		}
	}

	if len(matchedFiles) == 0 {
		return nil, errors.New("no .sql files found in the directory")
	}

	sort.Slice(matchedFiles, func(i, j int) bool {
		return matchedFiles[i].ModTime().After(matchedFiles[j].ModTime())
	})

	return matchedFiles[0], nil
}

// update constant values in a php file
func UpdateDefineValues(filePath string, updates map[string]string) error {
	fileData, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	// Regular expression to match any define variation
	re := regexp.MustCompile(`define\s*\(\s*['"](?P<key>[^'\"]+)['"]\s*,\s*['"](?P<value>[^'\"]+)['"]\s*\);`)

	// Replace all matches and create new data
	newData := re.ReplaceAllStringFunc(string(fileData), func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) > 2 {
			// fmt.Println("key:", matches[1], "value:", matches[2])
			key := matches[1]
			if newValue, ok := updates[key]; ok {
				return fmt.Sprintf("define('%s', '%s');", key, newValue)
			}
		}
		return match // No match or no update for key, return original string
	})

	// store current permission
	cmd := exec.Command("stat", "-c", "%a", filePath)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting permission: %w", err)
	}
	perm := strings.TrimSpace(string(out))

	// set write permission with sudo
	cmd = exec.Command("sudo", "chmod", "666", filePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error setting write permission: %w", err)
	}

	// Write back the modified data to the file
	err = os.WriteFile(filePath, []byte(newData), 0644)
	if err != nil {
		return err
	}

	// restore permission
	cmd = exec.Command("sudo", "chmod", perm, filePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error restoring permission: %w", err)
	}

	return nil
}
