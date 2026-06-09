package inspect

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var freezeFormats = map[string]bool{
	"svg": true,
	"png": true,
}

func isFreezeFormat(format string) bool {
	return freezeFormats[format]
}

func renderFreeze(vtContent string, format string, outputPath string) (string, error) {
	freezePath, err := exec.LookPath("freeze")
	if err != nil {
		return "", fmt.Errorf("freeze not found in PATH: %w (install with: go install github.com/charmbracelet/freeze@latest)", err)
	}

	if outputPath == "" {
		tmpFile, err := os.CreateTemp("", "tabdance-test-*."+format)
		if err != nil {
			return "", fmt.Errorf("create temp file: %w", err)
		}
		outputPath = tmpFile.Name()
		tmpFile.Close()
	}

	args := []string{
		"--language", "ansi",
		"--output", outputPath,
	}

	cmd := exec.Command(freezePath, args...)
	cmd.Stdin = strings.NewReader(vtContent)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("freeze failed: %w: %s", err, stderr.String())
	}

	if format == "svg" {
		data, err := os.ReadFile(outputPath)
		if err != nil {
			return "", fmt.Errorf("read output: %w", err)
		}
		return string(data), nil
	}

	return outputPath, nil
}
