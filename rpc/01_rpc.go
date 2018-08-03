// Basic RPC hello world
package main

import (
	"log"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

type FooAPI struct {
}

// A valid API method is exported, has a pointer receiver and returns error as last argument. The method is called with <registeredname>_helloWorld, where the first letter in method is lowercase, module name and method name separated by underscore
func (a *FooAPI) HelloWorld() (string, error) {
	return "foobar", nil
}

func main() {
	// Setup RPC server
	rpcsrv := rpc.NewServer()
	if err := rpcsrv.RegisterName("foo", &FooAPI{}); err != nil {
		log.Fatalf("[register_rpc] fail: %v", err)
	}
	// Stop the rpcserver
	defer rpcsrv.Stop()

	// Setup IPC endpoint
	ipcpath := ".hello.ipc"
	ipclistener, err := net.Listen("unix", ipcpath)
	if err != nil {
		log.Fatalf("[register_ipc] create fail: %v", err)
	}
	defer os.Remove(ipcpath)

	// Mount RPC server on IPC endpoint
	go func() {
		if err := rpcsrv.ServeListener(ipclistener); err != nil {
			log.Fatalf("[ipc dial] error dialing: %v", err)
		}
	}()

	// Setup IPC Client
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		log.Fatalf("[ipc_dial] dialing IPC fail: %v", err)
	}
	defer rpcclient.Close()

	// Call the RPC method
	var result string
	if err := rpcclient.Call(&result, "foo_helloWorld"); err != nil {
		log.Fatalf("[rpc_call] fail calling RPC method helloworld: %v", err)
	}

	log.Printf("[rpc_result] success calling method helloworld: %v\n", result)
}
