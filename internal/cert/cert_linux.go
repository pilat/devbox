//go:build linux
// +build linux

package cert

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (c *cert) isSynced() bool {
	distroCertPaths := []string{
		"/usr/local/share/ca-certificates/devbox-ca.crt", // Debian/Ubuntu
		"/etc/pki/ca-trust/source/anchors/devbox-ca.crt", // RHEL/CentOS/Fedora
	}

	for _, path := range distroCertPaths {
		certData, err := os.ReadFile(path)
		if err != nil {
			continue // Skip if file does not exist or cannot be read
		}

		// Check if the file contains the same certificate as our CA
		if bytes.Equal(certData, pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificate, Bytes: c.cert.Raw})) {
			return true
		}
	}

	return false
}

func (c *cert) installCA() error {
	distroPaths := []struct {
		certDir   string
		updateCmd string
	}{
		{"/usr/local/share/ca-certificates", "update-ca-certificates"},  // Debian/Ubuntu
		{"/etc/pki/ca-trust/source/anchors", "update-ca-trust extract"}, // RHEL/CentOS/Fedora
	}

	var lastErr error
	for _, distro := range distroPaths {
		// Ensure the directory exists
		if err := os.MkdirAll(distro.certDir, 0755); err != nil {
			lastErr = err
			continue
		}

		// Write the CA certificate to the appropriate directory
		caFilePath := filepath.Join(distro.certDir, "devbox-ca.crt")
		if err := os.WriteFile(caFilePath, pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificate, Bytes: c.cert.Raw}), 0644); err != nil {
			lastErr = err
			continue
		}

		// Run the update command to refresh the CA store
		cmd := exec.Command("sh", "-c", distro.updateCmd)
		if err := cmd.Run(); err != nil {
			lastErr = err
			continue
		}

		// Successfully installed and updated the CA
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("failed to install CA on Linux: %w", lastErr)
	}

	return fmt.Errorf("could not find a supported Linux distribution for CA installation")
}
