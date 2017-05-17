package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"strconv"

	"github.com/prashanthpai/sunrpc"
)

func main() {

	programNumber := uint32(12345)
	programVersion := uint32(1)

	_ = sunrpc.RegisterProcedure(sunrpc.Procedure{
		ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(1)},
		Name: "Arith.Add"}, true)
	_ = sunrpc.RegisterProcedure(sunrpc.Procedure{
		ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(2)},
		Name: "Arith.Multiply"}, true)

	sunrpc.DumpProcedureRegistry()

	// Get port from portmapper
	port, err := sunrpc.PmapGetPort("", programNumber, programVersion, sunrpc.IPProtoTCP)
	if err != nil {
		log.Fatal("sunrpc.PmapGetPort() failed: ", err)
	}

	// Connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(int(port)))
	if err != nil {
		log.Fatal("net.Dial() failed: ", err)
	}

	// Get notified on server closes the connection
	notifyClose := make(chan io.ReadWriteCloser, 5)
	go func() {
		for rwc := range notifyClose {
			conn := rwc.(net.Conn)
			log.Printf("Server %s disconnected", conn.RemoteAddr().String())
		}
	}()

	// Create client using sunrpc codec
	client := rpc.NewClientWithCodec(sunrpc.NewClientCodec(conn, notifyClose))

	// Remote function's arguments and results placeholder
	args := Args{7, 8}
	var reply int

	err = client.Call("Arith.Add", args, &reply)
	if err != nil {
		log.Print("client.Call() failed: ", err)
	}
	fmt.Printf("Arith Add: %d + %d = %d\n", args.A, args.B, reply)

	err = client.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Print("client.Call() failed: ", err)
	}
	fmt.Printf("Arith Multiply: %d * %d = %d\n", args.A, args.B, reply)
}
