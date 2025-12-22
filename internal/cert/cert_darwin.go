//go:build darwin
// +build darwin

package cert

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os/exec"
)

func (c *cert) isSynced() bool {
	outBuf := new(bytes.Buffer)
	cmd := exec.Command("security", "find-certificate", "-c", certName, "-a", "-p")
	cmd.Stdout = outBuf
	if err := cmd.Run(); err != nil {
		return false
	}

	// The command may return multiple PEM blocks if there are multiple
	// certificates with matching names. We need to check all of them.
	data := outBuf.Bytes()
	for len(data) > 0 {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		data = rest

		if block.Type != pemTypeCertificate {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		if c.cert.Equal(cert) {
			return true
		}
	}

	return false
}

func (c *cert) installCA() error {
	cmd := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", c.certFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install certificate: %w", err)
	}

	return nil
}
