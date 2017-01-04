package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/prashanthpai/sunrpc"
)

func main() {

	programNumber := uint32(12345)
	programVersion := uint32(1)

	// TODO: Automate this by parsing the .x file ?
	_ = sunrpc.RegisterProcedure(sunrpc.ProcedureID{programNumber, programVersion, uint32(1)}, "Arith.Add")
	_ = sunrpc.RegisterProcedure(sunrpc.ProcedureID{programNumber, programVersion, uint32(2)}, "Arith.Multiply")

	sunrpc.DumpProcedureRegistry()

	// TODO: Get port from portmapper
	conn, err := net.Dial("tcp", "127.0.0.1:41707")
	if err != nil {
		log.Fatal("net.Dial() failed: ", err)
	}

	client := rpc.NewClientWithCodec(sunrpc.NewClientCodec(conn))
	args := Args{7, 8}
	var reply int
	err = client.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	fmt.Printf("Arith: %d*%d=%d", args.A, args.B, reply)
}
