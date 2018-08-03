// Send, receive, get notified about a message

package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

// Creates a protocol that can take care of message sending
// the p2p.Peer represents the remote peer, and p2p.MsgReadWriter represents the io between the node and its peer
var (
	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			// Simplest payload possible - a byte slice
			outmsg := "foobar"
			if err := p2p.Send(rw, 0, outmsg); err != nil {
				return fmt.Errorf("send p2p msg fail: %v", err)
			}
			log.Println("[send] message", "[peer]", p, "[outmsg]", outmsg)

			// Wait for the message to come in from the other side
			if inmsg, err := rw.ReadMsg(); err != nil {
				return fmt.Errorf("receive p2p msg fail: %v", err)
			} else {
				log.Println("[receive] message", "[peer]", p, "[msg]", inmsg)
			}
			// Terminate the protocol
			return nil
		},
	}
)

func main() {
	var wg sync.WaitGroup

	srv1 := newServer("foo", "v1", 0)
	srv2 := newServer("bar", "v2", 30000)

	if err := srv1.Start(); err != nil {
		log.Fatal("failed to start srv1", err)
	}
	if err := srv2.Start(); err != nil {
		log.Fatal("failed to start srv2", err)
	}

	// Set up the event subscription on both servers
	evt1 := make(chan *p2p.PeerEvent)
	sub1 := srv1.SubscribeEvents(evt1)
	wg.Add(1)

	go func() {
		for {
			evt := <-evt1
			switch evt.Type {
			case "add":
				log.Println("[node1] received event", "[type]", evt.Type, "[peer]", evt.Peer)
			case "msgrecv":
				log.Println("[node1] receive message", "[type]", evt.Type, "[peer]", evt.Peer)
				wg.Done()
				return
			default:
			}
		}
	}()

	evt2 := make(chan *p2p.PeerEvent)
	sub2 := srv2.SubscribeEvents(evt2)
	wg.Add(1)

	go func() {
		for {
			evt := <-evt2
			switch evt.Type {
			case "add":
				log.Println("[node2] received event", "[type]", evt.Type, "[peer]", evt.Peer)
			case "msgrecv":
				log.Println("[node2] receive message", "[type]", evt.Type, "[peer]", evt.Peer)
				wg.Done()
				return
			default:
			}
		}
	}()

	// Get the node instance of the second server
	node2 := srv2.Self()

	// Add it as a peer to the first node
	// The connection and the crypto handshake will be performed automatically
	srv1.AddPeer(node2)

	wg.Wait()

	sub1.Unsubscribe()
	sub2.Unsubscribe()

	srv1.Stop()
	srv2.Stop()
}

func newServer(name string, version string, port int) *p2p.Server {

	// We need to explicitly allow at least one peer otherwise the connection attempt will be refused
	privKey, _ := crypto.GenerateKey()

	// Explicitly tell the server to generate events for messages
	cfg := p2p.Config{
		PrivateKey:      privKey,
		Name:            common.MakeName(name, version),
		MaxPeers:        1,
		Protocols:       []p2p.Protocol{proto},
		EnableMsgEvents: true,
	}
	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}
	srv := &p2p.Server{
		Config: cfg,
	}
	return srv
}
