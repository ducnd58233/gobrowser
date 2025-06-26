package browser

import (
	"crypto/rand"
	"fmt"
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
