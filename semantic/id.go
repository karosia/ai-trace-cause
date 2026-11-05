package semantic

// IDGenerator generates unique IDs for entities that are recorded
// without a caller-provided ID. Relationship IDs are not generated;
// they are derived deterministically by RelationID so that recording
// the same relationship twice is idempotent. Implementations must be
// safe for concurrent use.
type IDGenerator interface {
	NewID() (string, error)
}
