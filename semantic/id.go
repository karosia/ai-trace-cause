package semantic

// IDGenerator generates unique IDs for entities and relationships that
// are recorded without a caller-provided ID. Implementations must be
// safe for concurrent use.
type IDGenerator interface {
	NewID() (string, error)
}
