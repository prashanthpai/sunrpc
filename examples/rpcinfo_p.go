// Example go implementation of `rpcinfo -p` command
package main

import (
	"fmt"
	"log"

	"github.com/prashanthpai/sunrpc"
)

func main() {

	maps, err := sunrpc.PmapGetMaps("")
	if err != nil {
		log.Fatal("sunrpc.PmapGetMaps() failed: " + err.Error())
	}

	protocols := make(map[uint32]string, 2)
	protocols[uint32(6)] = "tcp"
	protocols[uint32(17)] = "udp"

	fmt.Printf("\tprogram\tvers\tproto\tport\tservice\t\n")
	for _, m := range maps {
		fmt.Printf("\t%d\t%d\t%s\t%d\n", m.Program, m.Version, protocols[m.Protocol], m.Port)
	}

}
