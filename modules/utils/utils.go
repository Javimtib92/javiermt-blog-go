package utils

import (
	"embed"
	"regexp"
	"strings"
)

var StaticAssets embed.FS

const CSS_PATH = "web/static/css/styles.css"

func GetAccentBaseValue() string {
	// Read the file synchronously
	fileContent, err := StaticAssets.ReadFile(CSS_PATH)
	if err != nil {
		// Handle error, e.g., log or return an error value
		return ""
	}

	// Convert byte slice to string
	fileContentStr := string(fileContent)

	// Find the line containing --color-accent-base and extract its value
	re := regexp.MustCompile(`--color-accent-base:\s*([^;]+)`)
	match := re.FindStringSubmatch(fileContentStr)

	if len(match) < 2 {
		return ""
	}

	accentBaseValue := strings.TrimSpace(match[1])
	return accentBaseValue
}