package main

import (
	"log"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {

	srv := newServer("foo", "42")
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	rpcsrv := rpc.NewServer()
	if err := rpcsrv.RegisterName("foo", &srv); err != nil {
		log.Fatal("[register api method] fail:", err)
	}
	defer rpcsrv.Stop()

	// Setup IPC endpoint
	ipcpath := "demo.ipc"
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		log.Fatal("[ipc_create] create IPC endpoint fail", err)
	}
	defer os.Remove(ipcpath)

	// Mount RPC server on IPC endpoint
	go func() {
		if err := rpcsrv.ServeListener(ipclistener); err != nil {
			log.Fatal("[mount rpc] on ipc fail", err)
		}
	}()

	// Create IPC client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		log.Fatal("[ipc dial] fail to dial", err)
	}
	defer rpcclient.Close()

	// Call the RPC method
	var nodeinfo p2p.NodeInfo
	if err := rpcclient.Call(&nodeinfo, "foo_nodeInfo"); err != nil {
		log.Fatal("[rpccall] fail to call rpc", err)
	}
	log.Println("server started")
	log.Println("enode", nodeinfo.Enode)
	log.Println("name", nodeinfo.Name)
	log.Println("ID", nodeinfo.ID)
	log.Println("IP", nodeinfo.IP)

}

func newServer(name, version string) p2p.Server {
	// make a new private key
	privkey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Generate private key failed", "err", err)
	}

	// set up p2p server
	cfg := p2p.Config{
		PrivateKey: privkey,
		Name:       common.MakeName(name, version),
	}
	srv := p2p.Server{
		Config: cfg,
	}
	return srv
}
