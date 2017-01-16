package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"strings"

	"github.com/prashanthpai/sunrpc"
	"github.com/soheilhy/cmux"
)

const (
	// This server listens on this port
	// You can change this and portmapper will take care of telling
	// the client about it.
	port = 49999
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

func serveHTTP(l net.Listener) {
	http.HandleFunc("/", hello)
	go http.Serve(l, nil)
}

func serveSunRPC(l net.Listener) {
	server := rpc.NewServer()
	arith := new(Arith)
	server.Register(arith)

	programNumber := uint32(12345)
	programVersion := uint32(1)

	_ = sunrpc.RegisterProcedure(sunrpc.ProcedureID{programNumber, programVersion, uint32(1)}, "Arith.Add")
	_ = sunrpc.RegisterProcedure(sunrpc.ProcedureID{programNumber, programVersion, uint32(2)}, "Arith.Multiply")

	sunrpc.DumpProcedureRegistry()

	// Tell portmapper about it
	_, err := sunrpc.PmapUnset(programNumber, programVersion)
	if err != nil {
		log.Fatal("sunrpc.PmapUnset() failed: ", err)
	}
	_, err = sunrpc.PmapSet(programNumber, programVersion, sunrpc.IPProtoTCP, uint32(port))
	if err != nil {
		log.Fatal("sunrpc.PmapSet() failed: ", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			if err != cmux.ErrListenerClosed {
				panic(err)
			}
			return
		}
		go server.ServeCodec(sunrpc.NewServerCodec(conn))
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("net.Listen() failed: ", err)
	}

	m := cmux.New(listener)
	httpl := m.Match(cmux.HTTP1Fast())
	sunrpcl := m.Match(sunrpc.CmuxMatcher())

	serveHTTP(httpl)
	go serveSunRPC(sunrpcl)

	if err := m.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		panic(err)
	}
}
