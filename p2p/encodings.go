package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is the interface that wraps the basic Decode method.
type Decoder interface {
	Decode(io.Reader, *RPC) error
}

// GoDecoder is a Decoder that uses the gob package to decode the RPC
type GOBDecoder struct{}

func (dec *GOBDecoder) Decode(r io.Reader, v *RPC) error {
	return gob.NewDecoder(r).Decode(v)
}

// DefaultDecoder is a Decoder that reads the RPC from the reader
type DefaultDecoder struct{}

func (dec *DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	buf := make([]byte, 1028)

	n, err := r.Read(buf)

	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
