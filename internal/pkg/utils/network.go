package utils

import (
	"fmt"
	"strings"

	"golang.org/x/net/idna"
)

func ConvertToRFCHostname(input string) (string, error) {
	input = strings.ReplaceAll(input, ".", "-")
	input = strings.ReplaceAll(input, "_", "-")

	// Trim any unnecessary whitespace
	trimmed := strings.TrimSpace(input)

	// Ensure the string is not empty after trimming
	if len(trimmed) == 0 {
		return "", fmt.Errorf("input string is empty")
	}

	// Use idna package to convert UTF-8 string to ASCII (Punycode) per RFC 3492/5891
	ascii, err := idna.New().ToASCII(trimmed)
	if err != nil {
		return "", fmt.Errorf("failed to convert to RFC hostname: %v", err)
	}

	// Split hostname into labels and validate each
	labels := strings.Split(ascii, ".")
	for _, label := range labels {
		if len(label) == 0 {
			return "", fmt.Errorf("invalid hostname: empty label")
		}
		if len(label) > 63 {
			return "", fmt.Errorf("invalid hostname: label exceeds 63 characters")
		}
	}

	// Ensure the total hostname length is valid
	if len(ascii) > 253 {
		return "", fmt.Errorf("invalid hostname: exceeds 253 characters")
	}

	return ascii, nil
}
