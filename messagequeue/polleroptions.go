package messagequeue

type PollerOptions struct {
}

func NewPollerOptions() *PollerOptions {
	return &PollerOptions{}
}

// Validate that the values contained in this Options are complete and within the
// bounds necessary for operation.
func (po *PollerOptions) Validate() error {
	return nil
}
