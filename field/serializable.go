package field

// Serializable is an interface which is able to encode/decode itself
// from binary representation.
type Serializable interface {
	// AppendTo appends its binary representation to specified array.
	AppendTo([]byte) []byte

	// PruneFrom decodes itself from the top of specified array. It
	// returns remaining data and true if decoding was successful. If
	// data was insufficient implementations must return false and a
	// slice of zero length (or nil).
	PruneFrom([]byte) ([]byte, bool)
}
