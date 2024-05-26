package main

import (
	"fmt"
	"log"

	"github.com/OmGuptaIND/p2p"
)

// Make File Servers
func makeServer(addr string, nodes ...string) *FileServer {
	store := NewStore(StoreOpts{
		Root:              fmt.Sprintf("%s_%s", addr, "store"),
		PathTransformFunc: CASPathTransformFunc,
	})

	tcpTransport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddr: addr,
		ShakeHands: p2p.NOPHandShake,
		Decoder:    &p2p.DefaultDecoder{},
	})

	fileServer := NewFileServer(FileServerOpts{
		Store:          store,
		Transport:      tcpTransport,
		bootStrapNodes: nodes,
	})

	return fileServer
}

func main() {
	server1 := makeServer(":3000", "")
	server2 := makeServer(":4000", ":3000")

	go func() {
		if err := server1.Start(); err != nil {
			log.Println("Error starting server1", err)
		}
	}()

	if err := server2.Start(); err != nil {
		log.Println("Error starting server2", err)
	}

	select {}
}
