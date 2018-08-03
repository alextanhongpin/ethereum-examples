// initializing and starting the p2p Server

package main

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

func main() {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[privateKey]: %#v\n", privKey)

	name := common.MakeName("foo", "42")
	log.Println("[name]:", name)

	cfg := p2p.Config{
		PrivateKey: privKey,
		Name:       name,
	}
	srv := p2p.Server{
		Config: cfg,
	}
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	nodeinfo := srv.NodeInfo()
	log.Println(
		"server started",
		"enode:", nodeinfo.Enode,
		"name:", nodeinfo.Name,
		"ID:", nodeinfo.ID,
		"IP:", nodeinfo.IP,
	)

	srv.Stop()
}
