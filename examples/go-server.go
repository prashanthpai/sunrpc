package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/prashanthpai/sunrpc"
)

// Example
type Args struct {
	A, B int32
}

type Arith int32

func (t *Arith) Multiply(args *Args, reply *int32) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Add(args *Args, reply *int32) error {
	*reply = args.A + args.B
	return nil
}

func main() {
	server := rpc.NewServer()
	arith := new(Arith)
	server.Register(arith)

	programNumber := uint32(12345)
	programVersion := uint32(1)

	// TODO: Automate this by parsing the .x file ?
	_ = sunrpc.RegisterProcedure(sunrpc.ProcedureID{programNumber, programVersion, uint32(1)}, "Arith.Add")
	_ = sunrpc.RegisterProcedure(sunrpc.ProcedureID{programNumber, programVersion, uint32(2)}, "Arith.Multiply")

	sunrpc.DumpProcedureRegistry()

	// TODO: Get port from portmapper
	listener, err := net.Listen("tcp", "127.0.0.1:34217")
	if err != nil {
		log.Fatal("net.Listen() failed: ", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("listener.Accept() failed: ", err)
		}
		go server.ServeCodec(sunrpc.NewServerCodec(conn))
	}
}
