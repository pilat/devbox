package git

import (
	"net/url"
	"regexp"
	"strings"
)

// URL normalization patterns
var (
	sshPattern           = regexp.MustCompile(`^git@([^:]+):(.+)/([^/]+?)(?:\.git)?$`)
	sshExplicitPattern   = regexp.MustCompile(`^(?:ssh|git\+ssh)://(?:[^@]+@)?([^/:]+)(?::\d+)?/(.+)/([^/]+?)(?:\.git)?$`)
	gitProtocolPattern   = regexp.MustCompile(`^git://([^/:]+)(?::\d+)?/(.+)/([^/]+?)(?:\.git)?$`)
	azureSSHPattern      = regexp.MustCompile(`^git@ssh\.dev\.azure\.com:v3/([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`)
	azureHTTPSPattern    = regexp.MustCompile(`^https?://(?:[^@]+@)?dev\.azure\.com/([^/]+)/([^/]+)/_git/([^/]+?)(?:\.git)?$`)
	azureOldHTTPSPattern = regexp.MustCompile(`^https?://(?:[^@]+@)?([^.]+)\.visualstudio\.com/([^/]+)/_git/([^/]+?)(?:\.git)?$`)
)

// NormalizeURL parses and normalizes a git URL for comparison.
// Returns the normalized form (host/owner/repo) regardless of protocol (HTTPS, SSH, etc).
func NormalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)

	// Azure DevOps patterns (different hosts for SSH vs HTTPS, need special handling)
	if matches := azureSSHPattern.FindStringSubmatch(rawURL); matches != nil {
		return strings.ToLower("dev.azure.com/" + matches[1] + "/" + matches[2] + "/" + matches[3])
	}
	if matches := azureHTTPSPattern.FindStringSubmatch(rawURL); matches != nil {
		return strings.ToLower("dev.azure.com/" + matches[1] + "/" + matches[2] + "/" + matches[3])
	}
	if matches := azureOldHTTPSPattern.FindStringSubmatch(rawURL); matches != nil {
		return strings.ToLower("dev.azure.com/" + matches[1] + "/" + matches[2] + "/" + matches[3])
	}

	// SSH pattern: git@host:owner/repo.git
	if matches := sshPattern.FindStringSubmatch(rawURL); matches != nil {
		return strings.ToLower(matches[1] + "/" + matches[2] + "/" + matches[3])
	}

	// Explicit SSH: ssh://git@host/owner/repo.git or git+ssh://
	if matches := sshExplicitPattern.FindStringSubmatch(rawURL); matches != nil {
		return strings.ToLower(matches[1] + "/" + matches[2] + "/" + matches[3])
	}

	// git:// protocol
	if matches := gitProtocolPattern.FindStringSubmatch(rawURL); matches != nil {
		return strings.ToLower(matches[1] + "/" + matches[2] + "/" + matches[3])
	}

	// Standard URL (http, https)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fallbackNormalize(rawURL)
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	return strings.ToLower(parsed.Host + "/" + path)
}

// fallbackNormalize provides basic normalization when parsing fails.
func fallbackNormalize(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "ssh://")
	s = strings.TrimPrefix(s, "git+ssh://")
	s = strings.TrimPrefix(s, "git://")
	s = strings.TrimPrefix(s, "git@")
	s = strings.ReplaceAll(s, ":", "/")
	s = strings.TrimSuffix(s, ".git")
	return s
}
