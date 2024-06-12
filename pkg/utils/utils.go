package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var FILE_FORMAT = map[string]string{
	"C++":        ".cpp",
	"Python3":    ".py",
	"Python":     ".py",
	"MySQL":      ".sql",
	"Go":         ".go",
	"Java":       ".java",
	"C":          ".c",
	"JavaScript": ".js",
	"TypeScript": ".ts",
	"PHP":        ".php",
	"C#":         ".cs",
	"Ruby":       ".rb",
	"Swift":      ".swift",
	"Scala":      ".scl",
	"Kotlin":     ".kt",
	"Rust":       ".rs",
}

// generatePathWithDir generates a path with a directory for the given problem details.
func GeneratePathWithDir(problemID, problemTitle, submissionLanguage, outputDir string) (string, error) {
	var pathname, filename string
	if isDigit(problemID[0]) {
		problemIDInt, err := strconv.Atoi(problemID)
		if err != nil {
			return "", err
		}
		pathname = filepath.Join(outputDir, fmt.Sprintf("%04d.%s", problemIDInt, problemTitle))
		filename = fmt.Sprintf("%d-%s%s", problemIDInt, problemTitle, FILE_FORMAT[submissionLanguage])
	} else {
		pathname = filepath.Join(outputDir, fmt.Sprintf("%s.%s", problemID, problemTitle))
		filename = fmt.Sprintf("%s-%s%s", problemID, problemTitle, FILE_FORMAT[submissionLanguage])
	}

	if _, err := os.Stat(pathname); os.IsNotExist(err) {
		if err := os.MkdirAll(pathname, os.ModePerm); err != nil {
			return "", err
		}
	}

	return filepath.Join(pathname, filename), nil
}

// generatePath generates a path for the given problem details.
func GeneratePath(problemID, problemTitle, submissionLanguage, outputDir string) (string, error) {
	var pathname, filename string
	if isDigit(problemID[0]) {
		problemIDInt, err := strconv.Atoi(problemID)
		if err != nil {
			return "", err
		}
		pathname = outputDir
		filename = fmt.Sprintf("%04d.%s%s", problemIDInt, problemTitle, FILE_FORMAT[submissionLanguage])
	} else {
		pathname = outputDir
		filename = fmt.Sprintf("%s.%s%s", problemID, problemTitle, FILE_FORMAT[submissionLanguage])
	}

	if _, err := os.Stat(pathname); os.IsNotExist(err) {
		if err := os.MkdirAll(pathname, os.ModePerm); err != nil {
			return "", err
		}
	}

	return filepath.Join(pathname, filename), nil
}

// gitPush performs a git commit and push with the current date as the commit message.
func GitPush(pushDir string) error {
	today := time.Now().Format("2006-01-02")
	os.Chdir(pushDir)
	instructions := []string{
		"git add .",
		"git status",
		fmt.Sprintf("git commit -m \"%s\"", today),
		"git push",
	}

	for _, instruction := range instructions {
		cmd := exec.Command("sh", "-c", instruction)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("Git operation failed: %s\nOutput: %s", err, output)
			return err
		}
		log.Printf("~~~~~~~~~~~~~ %s finished! ~~~~~~~~", instruction)
	}
	return nil
}

// isDigit checks if a byte represents a digit (0-9).
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
