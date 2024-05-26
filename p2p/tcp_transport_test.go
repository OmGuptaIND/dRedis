package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTranspor(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr: ":8080",
		ShakeHands: NOPHandShake,
		Decoder:    &DefaultDecoder{},
	}

	listenAddr := ":8080"
	tr := NewTCPTransport(opts)
	assert.Equal(t, listenAddr, tr.ListenAddr)
	assert.Nil(t, tr.ListenAndAccept())
}
