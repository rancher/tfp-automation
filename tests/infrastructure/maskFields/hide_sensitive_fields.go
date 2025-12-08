package maskFields

import (
	"strings"
)

const (
	ami               = "ami"
	credentials       = "credentials"
	config            = "config"
	key               = "key"
	password          = "password"
	privateregistries = "privateregistries"
	secret            = "secret"
	standalone        = "standalone"
	token             = "token"
)

// HideSensitiveFields returns a copy of the map with sensitive values replaced by '••••••'.
func HideSensitiveFields(m map[string]any) map[string]any {
	masked := make(map[string]any, len(m))

	for key, value := range m {
		lower := strings.ToLower(key)

		if strings.Contains(lower, password) || strings.Contains(lower, secret) || strings.Contains(lower, token) || strings.Contains(lower, key) ||
			strings.Contains(lower, credentials) || strings.Contains(lower, ami) || strings.Contains(lower, config) ||
			strings.Contains(lower, privateregistries) || strings.Contains(lower, standalone) {
			masked[key] = "•••••"
		} else {
			masked[key] = value
		}
	}

	return masked
}
