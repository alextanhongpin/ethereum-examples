// Connecting two p2p servers
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

func main() {
	// If the port is zero, the operating system will pick a port
	srv1 := newServer("foo", "v1", 0)
	if err := srv1.Start(); err != nil {
		log.Fatal("failed to start srv1:", err)
	}
	srv2 := newServer("foo", "v2", 31234)
	if err := srv2.Start(); err != nil {
		log.Fatal("failed to start srv2:", err)
	}
	// Get the node instance of the second server
	node2 := srv2.Self()
	log.Println("[addPeer] srv1 adding peers")
	srv1.AddPeer(node2)

	// Wait for connection to establish
	time.Sleep(time.Millisecond * 100)
	log.Println("[srv1 peers]", srv1.Peers())
	log.Println("[srv2 peers]", srv2.Peers())

	// Stop the server
	srv1.Stop()
	srv2.Stop()
}

func newServer(name string, version string, port int) *p2p.Server {

	// We need to explicitly allow at least one peer otherwise the connection attempt will be refused
	privKey, _ := crypto.GenerateKey()

	cfg := p2p.Config{
		PrivateKey: privKey,
		Name:       common.MakeName(name, version),
		MaxPeers:   1,
	}
	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}
	srv := &p2p.Server{
		Config: cfg,
	}
	return srv
}
