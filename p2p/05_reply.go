// Send, receive and get notified about message
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

type Msg struct {
	Pong    bool
	Created time.Time
}

var (
	wg     sync.WaitGroup
	pingWg sync.WaitGroup

	proto = p2p.Protocol{
		Name:    "foo",
		Version: 42,
		Length:  1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

			pingWg.Add(1)
			ponged := false

			// Create the message structure
			msg := Msg{
				Pong:    false,
				Created: time.Now(),
			}

			// Send the message
			if err := p2p.Send(rw, 0, msg); err != nil {
				return fmt.Errorf("[send p2p] fail: %v", err)
			}
			log.Println("[sending ping] peer", p)

			for !ponged {
				// Wait for the message to come from the other side
				msg, err := rw.ReadMsg()
				if err != nil {
					return fmt.Errorf("[receive] p2p message fail: %v", err)
				}

				var decodedMsg Msg
				if err := msg.Decode(&decodedMsg); err != nil {
					return fmt.Errorf("[decoded] p2p message fail: %v", err)
				}

				if decodedMsg.Pong {
					log.Println("[received] pong", "[peer]", p)
					ponged = true
					pingWg.Done()
				} else {
					log.Println("[received] ping", "[peer]", p)
					msg := Msg{
						Pong:    true,
						Created: time.Now(),
					}
					if err := p2p.Send(rw, 0, msg); err != nil {
						return fmt.Errorf("[send] p2p message fail: %v", err)
					}
					log.Println("[sent] pong", "[peer]", p)
				}
			}

			pingWg.Wait()
			wg.Done()

			return nil
		},
	}
)

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

func main() {

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
			select {
			case evt := <-evt1:
				switch evt.Type {
				case "add":
					log.Println("[received] #node1", "[type]", evt.Type, "[peer]", evt.Peer)
				case "msgrecv":
					log.Println("[received] #node2", "[type]", evt.Type, "[peer]", evt.Peer)
				default:
				}
			case <-sub1.Err():
				return
			}
		}
	}()

	evt2 := make(chan *p2p.PeerEvent)
	sub2 := srv2.SubscribeEvents(evt2)
	wg.Add(1)

	go func() {
		for {
			select {
			case evt := <-evt2:
				switch evt.Type {
				case "add":
					log.Println("[received] #node1", "[type]", evt.Type, "[peer]", evt.Peer)
				case "msgrecv":
					log.Println("[received] #node2", "[type]", evt.Type, "[peer]", evt.Peer)
				default:
				}
			case <-sub2.Err():
				return
			}
		}
	}()

	node2 := srv2.Self()
	srv1.AddPeer(node2)
	wg.Wait()

	// Cleanup
	sub1.Unsubscribe()
	sub2.Unsubscribe()
	srv1.Stop()
	srv2.Stop()
}
