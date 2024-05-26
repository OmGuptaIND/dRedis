package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/OmGuptaIND/p2p"
)

type FileServerOpts struct {
	Store          *Store
	Transport      p2p.Transport
	bootStrapNodes []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.RWMutex
	peers    map[string]p2p.Peer

	quitCh chan struct{}
}

// NewFileServer creates a new file server
func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		quitCh:         make(chan struct{}),
		peerLock:       sync.RWMutex{},
		peers:          make(map[string]p2p.Peer),
	}
}

// Start starts the file server
func (f *FileServer) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}

	go f.bootStrapNodesNetwork()

	f.handleQuitSignal()

	return nil
}

// OnPeer is called when a new peer is connected
func (f *FileServer) OnPeer(peer p2p.Peer) error {
	f.peerLock.Lock()
	defer f.peerLock.Unlock()

	f.peers[peer.RemoteAddr().String()] = peer

	return nil
}

// bootStrapNodesNetwork bootstraps the network with the provided nodes
func (f *FileServer) bootStrapNodesNetwork() {
	for _, addr := range f.bootStrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			if err := f.Transport.Dial(addr); err != nil {
				log.Println("Error dialing bootstrap node", addr, err)
			}
		}(addr)
	}
}

// Stop stops the file server
func (f *FileServer) Stop() {
	close(f.quitCh)
}

// Stop stops the file server
func (f *FileServer) handleQuitSignal() {
	defer func() {
		time.Sleep(2 * time.Second)
		log.Println("File Server stopped")

		f.Transport.Close()
	}()

	for {
		select {
		case msg := <-f.Transport.Consume():
			fmt.Println("Received message", msg)
		case <-f.quitCh:
			return
		}

	}
}
