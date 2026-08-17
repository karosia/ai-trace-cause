package semantic

type IDGenerator interface {
	NewID() (string, error)
}
