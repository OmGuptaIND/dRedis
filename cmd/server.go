package main

import (
	"fmt"
	"log"
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

	quitCh chan struct{}
}

// NewFileServer creates a new file server
func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		quitCh:         make(chan struct{}),
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
