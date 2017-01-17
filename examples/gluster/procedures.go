package main

import (
	"fmt"
	"io/ioutil"
)

const (
	GLUSTER_HNDSK_PROGRAM uint32 = 14398633
	GLUSTER_HNDSK_VERSION uint32 = 2
)

const (
	// rpc/rpc-lib/src/protocol-common.h
	GF_HNDSK_NULL uint32 = iota
	GF_HNDSK_SETVOLUME
	GF_HNDSK_GETSPEC // 2
	GF_HNDSK_PING
	GF_HNDSK_SET_LK_VER
	GF_HNDSK_EVENT_NOTIFY
	GF_HNDSK_GET_VOLUME_INFO
	GF_HNDSK_GET_SNAPSHOT_INFO
	GF_HNDSK_MAXVALUE
)

type GfGetspecReq struct {
	Flags uint
	Key   string
	Xdata []byte // serialized dict
}

type GfGetspecRsp struct {
	OpRet   int
	OpErrno int
	Spec    string
	Xdata   []byte // serialized dict
}

type GfHandshake int32

func (t *GfHandshake) ServerGetspec(args *GfGetspecReq, reply *GfGetspecRsp) error {
	var err error
	var fileContents []byte

	_, err = DictUnserialize(args.Xdata)
	if err != nil {
		fmt.Println(err)
		goto Out
	}

	fileContents, err = ioutil.ReadFile("/var/lib/glusterd/vols/test/trusted-test.tcp-fuse.vol")
	if err != nil {
		fmt.Println(err)
		goto Out
	}
	reply.Spec = string(fileContents)
	reply.OpRet = len(reply.Spec)
	reply.OpErrno = 0

Out:
	if err != nil {
		reply.OpRet = -1
		reply.OpErrno = 0
	}

	return nil
}
