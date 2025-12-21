package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	certName           = "devbox development CA"
	pemTypeCertificate = "CERTIFICATE"
	pemTypePrivateKey  = "RSA PRIVATE KEY"
)

type cert struct {
	certFile string
	keyFile  string

	cert *x509.Certificate
	key  *rsa.PrivateKey
}

func SetupCA(appDir string) error {
	c := &cert{
		certFile: filepath.Join(appDir, "ca.crt"),
		keyFile:  filepath.Join(appDir, "ca.key"),
	}
	return c.setupCA()
}

func GeneratePair(appDir string, certFile, keyFile string, hosts []string) error {
	c := &cert{
		certFile: filepath.Join(appDir, "ca.crt"),
		keyFile:  filepath.Join(appDir, "ca.key"),
	}

	return c.generatePair(certFile, keyFile, hosts)
}

// setup sets up the certificate authority. Might require sudo on some systems.
func (c *cert) setupCA() error {
	err := c.loadCA()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	// if CA is expired or not found we are generating new one
	if c.isCAExpired() {
		if err = c.generateCA(); err != nil {
			return fmt.Errorf("failed to generate CA: %w", err)
		}
	}

	if !c.isSynced() {
		if err = c.installCA(); err != nil {
			return fmt.Errorf("failed to install CA: %w", err)
		}
	}

	return nil
}

func (c *cert) generatePair(certFile, keyFile string, hosts []string) error {
	if len(hosts) == 0 {
		return fmt.Errorf("no hosts provided")
	}

	err := c.loadCA()
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	needNew := func() bool {
		existingCertData, certErr := os.ReadFile(certFile)
		_, keyErr := os.ReadFile(keyFile)

		if certErr != nil || keyErr != nil {
			return true // At least one of the files does not exist
		}

		existingCert, err := decodePEM[x509.Certificate](existingCertData)
		if err != nil {
			return true // Failed to decode the certificate
		}

		// Check if the certificate is expired
		if time.Now().After(existingCert.NotAfter) {
			return true
		}

		// Check if the certificate has the same set of hosts
		certHosts := map[string]bool{}
		for _, dns := range existingCert.DNSNames {
			certHosts[dns] = true
		}

		if len(certHosts) != len(hosts) {
			return true
		}

		for _, host := range hosts {
			if !certHosts[host] {
				return true
			}
		}

		// Check if the certificate is signed by our CA
		if err := existingCert.CheckSignatureFrom(c.cert); err != nil {
			return true
		}

		return false
	}()

	if !needNew {
		return nil
	}

	clientCertPEM, clientKeyPEM, err := generateCertificate(false, c.cert, c.key, hosts[0], hosts[1:]...)
	if err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	if err := os.WriteFile(certFile, clientCertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write client certificate: %w", err)
	}

	if err := os.WriteFile(keyFile, clientKeyPEM, 0644); err != nil {
		return fmt.Errorf("failed to write client key: %w", err)
	}

	return nil
}

func (c *cert) loadCA() error {
	certData, err := os.ReadFile(c.certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse the certificate
	c.cert, err = decodePEM[x509.Certificate](certData)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %w", err)
	}

	// Read the key file
	keyData, err := os.ReadFile(c.keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse the key
	c.key, err = decodePEM[rsa.PrivateKey](keyData)
	if err != nil {
		return fmt.Errorf("failed to parse key: %w", err)
	}

	return nil
}

func (c *cert) isCAExpired() bool {
	if c.cert == nil {
		return true
	}

	return time.Now().After(c.cert.NotAfter)
}

func (c *cert) generateCA() error {
	caCertPEM, caKeyPEM, err := generateCertificate(true, nil, nil, certName)
	if err != nil {
		return fmt.Errorf("failed to generate CA certificate: %w", err)
	}

	// Parse the certificate
	c.cert, err = decodePEM[x509.Certificate](caCertPEM)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Write the certificate
	if err := os.WriteFile(c.certFile, caCertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	// Parse the key
	c.key, err = decodePEM[rsa.PrivateKey](caKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse CA key: %w", err)
	}

	// Write the key
	if err := os.WriteFile(c.keyFile, caKeyPEM, 0644); err != nil {
		return fmt.Errorf("failed to write CA key: %w", err)
	}

	return nil
}

// decodePEM decodes PEM data into a given type
func decodePEM[T any](data []byte) (*T, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM data: no block found")
	}

	var result any
	var err error

	switch any((*T)(nil)).(type) {
	case *x509.Certificate:
		if block.Type != pemTypeCertificate {
			return nil, fmt.Errorf("invalid PEM block type: expected %s, got %s", pemTypeCertificate, block.Type)
		}

		result, err = x509.ParseCertificate(block.Bytes)
	case *rsa.PrivateKey:
		if block.Type != pemTypePrivateKey {
			return nil, fmt.Errorf("invalid PEM block type: expected %s, got %s", pemTypePrivateKey, block.Type)
		}

		result, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported type for decoding")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM block: %w", err)
	}

	typedResult, ok := result.(*T)
	if !ok {
		return nil, fmt.Errorf("unexpected type in PEM decoding")
	}

	return typedResult, nil
}

func generateCertificate(isCA bool, parentCert *x509.Certificate, parentKey *rsa.PrivateKey, commonName string, extra ...string) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	if isCA {
		certTemplate.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		certTemplate.NotAfter = time.Now().Add(2 * 365 * 24 * time.Hour) // 2 year for CA
	} else {
		certTemplate.DNSNames = append(certTemplate.DNSNames, commonName)
		certTemplate.DNSNames = append(certTemplate.DNSNames, extra...)
	}

	if parentCert == nil || parentKey == nil {
		parentCert = certTemplate
		parentKey = key
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, parentCert, &key.PublicKey, parentKey)
	if err != nil {
		return nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificate, Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: pemTypePrivateKey, Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}
