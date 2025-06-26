package browser

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type IDGenerator interface {
	Generate() string
}

type idGenerator struct{}

func NewIDGenerator() IDGenerator {
	return &idGenerator{}
}

func (g *idGenerator) Generate() string {
	bytes := make([]byte, IDByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("id_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", bytes)
}

type URLNormalizer interface {
	Normalize(rawURL string) (string, error)
	IsValidScheme(scheme string) bool
}

type urlNormalizer struct {
	supportedSchemes map[string]bool
}

func NewURLNormalizer() URLNormalizer {
	return &urlNormalizer{
		supportedSchemes: map[string]bool{
			"http":  true,
			"https": true,
		},
	}
}

func (n *urlNormalizer) Normalize(rawURL string) (string, error) {
	if rawURL == "" {
		return "", ErrInvalidURL
	}

	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrInvalidURL, err)
	}

	return parsed.String(), nil
}

func (n *urlNormalizer) IsValidScheme(scheme string) bool {
	return n.supportedSchemes[scheme]
}
