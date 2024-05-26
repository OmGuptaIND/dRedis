package p2p

import "net"

// Peer is defined as the Remote Person that we want to communicate with
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport is the interface that wraps the basic methods for a transport.
// Used to communicate between two peers
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
