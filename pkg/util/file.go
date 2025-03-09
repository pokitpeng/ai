package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsTextFile checks if a file is a text file
func IsTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	// Common text file extensions
	textExtensions := map[string]bool{
		// Code files
		".go":    true,
		".py":    true,
		".js":    true,
		".ts":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".php":   true,
		".rb":    true,
		".swift": true,
		".kt":    true,
		".rs":    true,
		".scala": true,
		".sh":    true,
		".bash":  true,
		".pl":    true,
		".r":     true,

		// Markup languages
		".html": true,
		".htm":  true,
		".xml":  true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".md":   true,
		".rst":  true,
		".tex":  true,
		".css":  true,
		".scss": true,
		".sass": true,
		".less": true,

		// Configuration files
		".conf":   true,
		".config": true,
		".ini":    true,
		".toml":   true,
		".env":    true,

		// Other text files
		".txt": true,
		".log": true,
		".csv": true,
		".tsv": true,
	}

	return textExtensions[ext]
}

// ReadTextFile reads the content of a text file
func ReadTextFile(filename string) (string, error) {
	// Check if the file is a text file
	if !IsTextFile(filename) {
		return "", fmt.Errorf("unsupported file type: %s", filename)
	}

	// Ensure the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", err)
	}

	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %s", err)
	}

	return string(content), nil
}

// GetFileInfo gets file information
func GetFileInfo(filename string) (string, string, error) {
	// Read file content
	content, err := ReadTextFile(filename)
	if err != nil {
		return "", "", err
	}

	// Get language type
	language := detectLanguage(filename)

	return content, language, nil
}

// detectLanguage detects programming language based on file extension
func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Simple mapping from extension to language
	langMap := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".java":  "Java",
		".c":     "C",
		".cpp":   "C++",
		".cs":    "C#",
		".php":   "PHP",
		".rb":    "Ruby",
		".html":  "HTML",
		".css":   "CSS",
		".rs":    "Rust",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
		".r":     "R",
		".sh":    "Shell",
		".bash":  "Bash",
		".json":  "JSON",
		".yaml":  "YAML",
		".yml":   "YAML",
		".md":    "Markdown",
		".xml":   "XML",
		".sql":   "SQL",
		".pl":    "Perl",
		".txt":   "Text",
	}

	lang, ok := langMap[ext]
	if !ok {
		// Default to plain text
		return "Text"
	}

	return lang
}
