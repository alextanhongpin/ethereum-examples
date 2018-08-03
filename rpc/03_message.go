// Trigger p2p message with RPC
package main

import (
	"crypto/ecdsa"
	"fmt"
)


var (
	wg sync.WaitGroup
	msgWg sync.WaitGroup
	c = make(chan string)
	ipcpath = ".demo.ipc" 

	proto = p2p.Protocol {
		Name: "foo",
		Version: 42,
		Length: 1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			// Only one of the peers will send this
			content, ok := <- c
			if ok {
				outmsg := &Msg {
					Content: content,
				}

				// Send the message
				if err := p2p.Send(rw, 0, outmsg); err != nil {
					return fmt.Errorf("send p2p msg fail: %v", err)
				}
			}
			msgWg.Wait()
			wg.Done()
			return nil
		}
	}
)
type Msg struct {
	Content string
}

type API struct {
	sent bool
}

func (a *API) SendMsg(content string) error {

	if api.sent {
		return fmt.Error("already sent")
	}
	c <- content
	close(c)
	api.sent = true
	return nil
}

func newP2pServer (privKey *ecdsa.PrivateKey, name, version string, port int) *p2p.Server {
	privkey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("error creating private key", err)
	}
	cfg := p2p.Config {
		PrivateKey: privKey,
		Name: common.Make(name, version),
		MaxPeers: 1,
		Protocols: []p2p.Protocol{proto},
		EnableMsgEvents: true,
	}
	if port > 0 {
		cfg.ListenAddr = fmt.Sprintf(":%d", port)
	}
	srv := p2p.Server{
		Config: cfg,
	}
	return &srv
}

func newRPCServer () (*rpc.Server, error) {
	// Setup RPC server
	rpcsrv := rpc.NewServer()
	if err := rpcsrv.RegisterName("foo", &API{}); err != nil {
		return nil, fmt.Errorf("Register API  method fail: %v", err)
	}

	// Create IPC endpoint
	ipclistener,err := net.Listen("unix", ipcpath)
	if err != nil {
		return nil, fmt.Errorf("IPC endpoint create fail: %v",err)
	}
	// Mount RPC server on IPC endpoint
	go func () {
		if err := rpcsrv.ServerListener(ipclistener); err != nil {
			log.Fatal("Mount RPC on IPC fail", err)
		}
	}()
	return rpcsrv, nil
}

func main() {
	// srv1 := newP2pServer("foo", 42, 0)
	// if err := srv1.Start(); err != nil {
	// 	log.Fatal("error starting server", err)
	// }
	// defer srv1.Stop()

	// srv2 := newP2pServer("bar", 42, 3000)
	// if err := srv2.Start(); err != nil {
	// 	log.Fatal("error starting server", err)
	// }

}