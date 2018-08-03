// Receiving notification of completed connection. Get notified when the peer connection has been completed.
package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

func main() {
	quit := make(chan bool)

	srv1 := newServer("foo", "v1", 0)
	// When setting srv2 port to 0, it will hang
	srv2 := newServer("bar", "v2", 30000)

	if err := srv1.Start(); err != nil {
		log.Fatal("failed to start srv1", err)
	}
	if err := srv2.Start(); err != nil {
		log.Fatal("failed to start srv2", err)
	}

	evt := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(evt)

	// Listen for events
	go func() {
		preEvt := <-evt
		log.Println("[preEvent] received preEvent", "[type]", preEvt.Type, "[peer]", preEvt.Peer)
		quit <- true
	}()

	// Get the node instance of the second server
	node2 := srv2.Self()

	srv1.AddPeer(node2)

	<-quit

	// Inspect the results
	log.Println("[peers]")
	log.Println("[node1]", srv1.Peers())
	log.Println("[node2]", srv2.Peers())

	// Terminate subscription
	sub1.Unsubscribe()

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
