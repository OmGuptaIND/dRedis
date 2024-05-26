package main

import (
	"bytes"
	"encoding/gob"
	"io"
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
	go f.handleQuitSignal()

	return nil
}

type Payload struct {
	Key  string
	Data []byte
}

// broadCast sends the payload to all the peers
func (f *FileServer) broadCast(payload *Payload) error {
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(payload)
}

func (f *FileServer) StoreFile(key string, r io.Reader) error {

	buf := new(bytes.Buffer)

	tee := io.TeeReader(r, buf)

	if err := f.Store.Write(key, tee); err != nil {
		return err
	}

	if _, err := io.Copy(buf, r); err != nil {
		return err
	}

	log.Println("Storing file", buf.Bytes())

	payload := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}

	return f.broadCast(payload)
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
			var payload Payload

			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&payload); err != nil {
				log.Println("Error decoding payload", err)
			}

			log.Println("Received payload", payload)

		case <-f.quitCh:
			return
		}
	}
}
