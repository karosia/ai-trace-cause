package uuidv7_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/karosia/ai-trace-cause/idgen/uuidv7"
)

func TestGeneratorNewID(t *testing.T) {
	generator := uuidv7.New()

	id, err := generator.NewID()
	if err != nil {
		t.Fatalf(
			"NewID() error = %v",
			err,
		)
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf(
			"uuid.Parse() error = %v",
			err,
		)
	}

	if parsed.Version() != 7 {
		t.Errorf(
			"Version() = %d, want 7",
			parsed.Version(),
		)
	}
}

func TestGeneratorProducesDifferentIDs(
	t *testing.T,
) {
	generator := uuidv7.New()

	first, err := generator.NewID()
	if err != nil {
		t.Fatal(err)
	}

	second, err := generator.NewID()
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatalf(
			"generated duplicate IDs: %q",
			first,
		)
	}
}
