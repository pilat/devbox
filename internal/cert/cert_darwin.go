//go:build darwin
// +build darwin

package cert

import (
	"bytes"
	"crypto/x509"
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

	certFromKeychain, err := decodePEM[x509.Certificate](outBuf.Bytes())
	if err != nil {
		return false
	}

	return c.cert.Equal(certFromKeychain)
}

func (c *cert) installCA() error {
	cmd := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", c.certFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install certificate: %w", err)
	}

	return nil
}
