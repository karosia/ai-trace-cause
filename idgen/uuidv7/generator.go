// Package uuidv7 implements semantic.IDGenerator using UUIDv7, the
// default ID generator used by aitracecause.Trace when no other
// generator is configured.
package uuidv7

import (
	"github.com/google/uuid"

	"github.com/karosia/ai-trace-cause/semantic"
)

// Generator is a semantic.IDGenerator that produces UUIDv7 values.
// Its zero value is ready to use.
type Generator struct{}

var _ semantic.IDGenerator = (*Generator)(nil)

// New creates a Generator.
func New() *Generator {
	return &Generator{}
}

// NewID returns a newly generated UUIDv7 string.
func (g *Generator) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
