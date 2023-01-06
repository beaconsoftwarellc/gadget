package transaction

// Begin has methods for starting transactions
type Begin interface {
	// Begin transaction on the underlying transactable datastructure and
	// return it
	Begin() (Implementation, error)
}
