package main

import (
	"io"
	"log"
	"net"
	"net/rpc"
	"strconv"

	"github.com/prashanthpai/sunrpc"
)

const (
	// This server listens on this port
	// You can change this and portmapper will take care of telling
	// the client about it.
	port = 49999
)

func main() {
	server := rpc.NewServer()
	arith := new(Arith)
	server.Register(arith)

	programNumber := uint32(12345)
	programVersion := uint32(1)

	_ = sunrpc.RegisterProcedure(sunrpc.Procedure{
		ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(1)},
		Name: "Arith.Add"})
	_ = sunrpc.RegisterProcedure(sunrpc.Procedure{
		ID:   sunrpc.ProcedureID{programNumber, programVersion, uint32(2)},
		Name: "Arith.Multiply"})

	sunrpc.DumpProcedureRegistry()

	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("net.Listen() failed: ", err)
	}

	// Tell portmapper about it
	_, err = sunrpc.PmapUnset(programNumber, programVersion)
	if err != nil {
		log.Fatal("sunrpc.PmapUnset() failed: ", err)
	}
	_, err = sunrpc.PmapSet(programNumber, programVersion, sunrpc.IPProtoTCP, uint32(port))
	if err != nil {
		log.Fatal("sunrpc.PmapSet() failed: ", err)
	}

	notifyClose := make(chan io.ReadWriteCloser, 5)
	go func() {
		for rwc := range notifyClose {
			conn := rwc.(net.Conn)
			log.Printf("Client %s disconnected", conn.RemoteAddr().String())
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("listener.Accept() failed: ", err)
		}
		log.Printf("Client %s connected", conn.RemoteAddr().String())
		// Use sunrpc's codec to handle incoming client connections
		go server.ServeCodec(sunrpc.NewServerCodec(conn, notifyClose))
	}
}
