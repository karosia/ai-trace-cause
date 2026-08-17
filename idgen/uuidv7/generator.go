package uuidv7

import (
	"github.com/google/uuid"

	"github.com/karosia/ai-trace-cause/semantic"
)

type Generator struct{}

var _ semantic.IDGenerator = (*Generator)(nil)

func New() *Generator {
	return &Generator{}
}

func (g *Generator) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
