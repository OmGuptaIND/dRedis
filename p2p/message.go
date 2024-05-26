package p2p

import "net"

// Message is the basic structure of a message that is sent between two peers
type RPC struct {
	From    net.Addr
	Payload []byte
}
