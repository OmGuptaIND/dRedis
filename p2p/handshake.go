package p2p

// HandShakeFunc is a function that is called when a new connection is established
type HandShakeFunc func() error

// NOPHandShake is a No Operation Handshake function
func NOPHandShake() error {
	return nil
}
