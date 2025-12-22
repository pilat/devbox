package hosts

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

const (
	defaultHostFile     = "/etc/hosts"
	beginMarkerTemplate = "# BEGIN: Devbox '%s' project"
	endMarkerTemplate   = "# END: Devbox: '%s' project"
)

// Save updates the hosts file with the given entries. Returns (changed, error).
// If no changes were needed, changed is false and no write occurs.
func Save(projectName string, entries []string) (bool, error) {
	return save(defaultHostFile, projectName, entries)
}

func save(hostFile, projectName string, entries []string) (bool, error) {
	markerBegin := fmt.Sprintf(beginMarkerTemplate, projectName)
	markerEnd := fmt.Sprintf(endMarkerTemplate, projectName)

	fileInfo, err := os.Stat(hostFile)
	if err != nil {
		return false, fmt.Errorf("failed to stat hosts file: %w", err)
	}
	fileMode := fileInfo.Mode()

	oldContent, err := os.ReadFile(hostFile)
	if err != nil {
		return false, fmt.Errorf("failed to read hosts file: %w", err)
	}

	var newContent strings.Builder
	replaced := false
	lookupForEnd := false

	scanner := bufio.NewScanner(bytes.NewReader(oldContent))
	for scanner.Scan() {
		line := scanner.Text()

		if lookupForEnd {
			if strings.TrimSpace(line) == markerEnd {
				lookupForEnd = false
			}
			continue
		}

		if strings.TrimSpace(line) == markerBegin {
			lookupForEnd = true
			replaced = true

			if len(entries) == 0 {
				continue
			}

			newContent.WriteString(markerBegin + "\n")
			for _, entry := range entries {
				newContent.WriteString(entry + "\n")
			}
			newContent.WriteString(markerEnd + "\n")
			continue
		}

		newContent.WriteString(line + "\n")
	}

	if lookupForEnd {
		return false, fmt.Errorf("unexpected end of file")
	}

	if !replaced && len(entries) > 0 {
		newContent.WriteString(markerBegin + "\n")
		for _, entry := range entries {
			newContent.WriteString(entry + "\n")
		}
		newContent.WriteString(markerEnd + "\n")
	}

	if newContent.String() == string(oldContent) {
		return false, nil
	}

	err = os.WriteFile(hostFile, []byte(newContent.String()), fileMode)
	if err != nil {
		return false, fmt.Errorf("failed to write hosts file: %w", err)
	}
	return true, nil
}
