package browser

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"
)

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func normalizeURL(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}
