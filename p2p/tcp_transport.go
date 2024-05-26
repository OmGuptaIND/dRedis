package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// TCPPeer Represents the Remote Node over the TCP Connection
type TCPPeer struct {
	conn net.Conn

	// outbound is true if the connection was started by this node
	outbound bool
}

// Close closes the connection
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// RemoteAddr returns the remote address of the connection
func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

// NewTCPPeer creates a new TCPPeer
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr string
	ShakeHands HandShakeFunc
	Decoder    Decoder
	OnPeer     func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	rpcChan  chan RPC
	listener net.Listener
}

// NewTCPTransport creates a new TCPTransport
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcChan:          make(chan RPC),
	}
}

// Close closes the TCPTransport
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Consume returns a channel to consume the RPC messages
// return the read-only channel
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

// Listen of a listenAddr provided in the TCPTransport
// Accepts the connection and starts a goroutine to handle the connection
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP Transport Listening on %s\n", t.ListenAddr)

	return nil
}

// Dial dials the address provided in the TCPTransport
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

// Start starts the TCPTransport
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
		}

		go t.handleConn(conn, false)
	}

}

// handle the connection when a new connection is received
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	log.Println("Handling connection", conn.RemoteAddr(), "LocalAddrr", t.ListenAddr, "Outbound", outbound)

	var err error

	defer func() error {
		fmt.Println("Dropping the connection", conn.RemoteAddr())
		return conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.ShakeHands(); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			fmt.Println("Error in OnPeer", err)
			return
		}
	}

	rpc := RPC{}

	for {

		err = t.Decoder.Decode(conn, &rpc)

		if err != nil {
			fmt.Printf("TCP Read Error: %v\n", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcChan <- rpc
	}

}
