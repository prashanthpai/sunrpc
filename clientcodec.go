// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"sync"

	"github.com/rasky/go-xdr/xdr2"
)

type clientCodec struct {
	conn         io.ReadWriteCloser // network connection
	recordReader io.Reader          // reader for RPC record

	// Sun RPC responses include Seq (XID) but not ServiceMethod (procedure
	// number). Go package net/rpc expects both. So we save ServiceMethod
	// when sending the request and look it up when filling rpc.Response
	mutex   sync.Mutex        // protects pending
	pending map[uint64]string // maps Seq (XID) to ServiceMethod
}

// NewClientCodec returns a new rpc.ClientCodec using Sun RPC on conn
func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		conn:    conn,
		pending: make(map[uint64]string),
	}
}

// NewClient returns a new rpc.Client which internally uses Sun RPC codec
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

// Dial connects to a Sun-RPC server at the specified network address
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}

func (c *clientCodec) WriteRequest(req *rpc.Request, param interface{}) error {

	// rpc.Request.Seq is initialized (from 0) and incremented by net/rpc
	// package on each call. This is unit64. But XID as per RFC should
	// really be uint32. This increment should be capped till maxOf(uint32)

	procedureID, ok := GetProcedureID(req.ServiceMethod)
	if !ok {
		// Reply with standard RPC error
	}

	c.mutex.Lock()
	c.pending[req.Seq] = req.ServiceMethod
	c.mutex.Unlock()

	// Encapsulate rpc.Request.Seq and rpc.Request.ServiceMethod
	rpcBody := RPCMsgCall{
		Header: RPCMessageHeader{
			Xid:  uint32(req.Seq),
			Type: Call},
		Body: CallBody{
			RPCVersion: RPCVersionSupported,
			Program:    procedureID.ProgramNumber,
			Version:    procedureID.ProgramVersion,
			Procedure:  procedureID.ProcedureNumber,
		},
	}

	payload := new(bytes.Buffer)

	_, err := xdr.Marshal(payload, &rpcBody)
	if err != nil {
		return err
	}

	// Marshall actual params/args of the remote procedure
	_, err = xdr.Marshal(payload, &param)
	if err != nil {
		return err
	}

	// Write payload to network
	_, err = WriteFullRecord(c.conn, payload.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *clientCodec) ReadResponseHeader(resp *rpc.Response) error {

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		if err != io.EOF {
			fmt.Println("ReadFullRecord() failed: ", err)
		}
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshall record into reply payload
	var reply RPCMsgReply
	_, err = xdr.Unmarshal(c.recordReader, &reply)
	if err != nil {
		fmt.Println("xdr.Unmarshal() failed: ", err)
		return err
	}

	// Unpack rpc.Request.Seq and set rpc.Request.ServiceMethod
	resp.Seq = uint64(reply.Header.Xid)
	c.mutex.Lock()
	resp.ServiceMethod = c.pending[uint64(reply.Header.Xid)]
	delete(c.pending, uint64(reply.Header.Xid))
	c.mutex.Unlock()

	return nil
}

func (c *clientCodec) ReadResponseBody(result interface{}) error {

	if result == nil {
		// read and drain it out ?
		return nil
	}

	_, err := xdr.Unmarshal(c.recordReader, &result)
	if err != nil {
		fmt.Println("xdr.Unmarshal() failed: ", err)
		return err
	}

	return nil
}

func (c *clientCodec) Close() error {
	if tc, ok := c.conn.(*net.TCPConn); ok {
		return tc.CloseRead()
	}
	return c.conn.Close()
}
