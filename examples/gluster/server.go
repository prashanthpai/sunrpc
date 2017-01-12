package main

import (
	"log"
	"net"
	"net/rpc"
	"strconv"

	"github.com/prashanthpai/sunrpc"
)

const (
	port uint32 = 24007
)

func main() {

	server := rpc.NewServer()
	server.Register(new(GfHandshake))
	_ = sunrpc.RegisterProcedure(
		sunrpc.ProcedureID{
			GLUSTER_HNDSK_PROGRAM,
			GLUSTER_HNDSK_VERSION,
			GF_HNDSK_GETSPEC,
		},
		"GfHandshake.ServerGetspec")
	sunrpc.DumpProcedureRegistry()

	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(port)))
	if err != nil {
		log.Fatal("net.Listen() failed: ", err)
	}

	// Tell portmapper about the port this program is listening on
	_, err = sunrpc.PmapUnset(GLUSTER_HNDSK_PROGRAM, GLUSTER_HNDSK_VERSION)
	if err != nil {
		log.Fatal("sunrpc.PmapUnset() failed: ", err)
	}
	_, err = sunrpc.PmapSet(GLUSTER_HNDSK_PROGRAM, GLUSTER_HNDSK_VERSION, sunrpc.IPProtoTCP, port)
	if err != nil {
		log.Fatal("sunrpc.PmapSet() failed: ", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("listener.Accept() failed: ", err)
		}
		go server.ServeRequest(sunrpc.NewServerCodec(conn))
	}
}
