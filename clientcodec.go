// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
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
		return ErrProcUnavail
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

	if _, err := xdr.Marshal(payload, &rpcBody); err != nil {
		return err
	}

	// Marshall actual params/args of the remote procedure
	if _, err := xdr.Marshal(payload, &param); err != nil {
		return err
	}

	// Write payload to network
	_, err := WriteFullRecord(c.conn, payload.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *clientCodec) ReadResponseHeader(resp *rpc.Response) error {

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshal record as RPC reply
	var reply RPCMsgReply
	if _, err = xdr.Unmarshal(c.recordReader, &reply); err != nil {
		return err
	}

	// Unpack rpc.Request.Seq and set rpc.Request.ServiceMethod
	resp.Seq = uint64(reply.Header.Xid)
	c.mutex.Lock()
	resp.ServiceMethod = c.pending[resp.Seq]
	delete(c.pending, resp.Seq)
	c.mutex.Unlock()

	if reply.Header.Type != Reply {
		return ErrInvalidRPCMessageType
	}

	// Filter out all valid RPC error cases
	switch reply.Stat {
	case MsgAccepted:
		switch reply.Areply.Stat {
		case Success:
		case ProgMismatch:
			return ErrProgMismatch{
				reply.Areply.MismatchInfo.Low,
				reply.Areply.MismatchInfo.High}
		case ProgUnavail:
			return ErrProgUnavail
		case ProcUnavail:
			return ErrProcUnavail
		case GarbageArgs:
			return ErrGarbageArgs
		case SystemErr:
			return ErrSystemErr
		default:
			return ErrInvalidMsgAccepted
		}
	case MsgDenied:
		switch reply.Rreply.Stat {
		case RPCMismatch:
			return ErrRPCMismatch{
				reply.Rreply.MismatchInfo.Low,
				reply.Rreply.MismatchInfo.High}
		case AuthError:
			return ErrAuthError
		default:
			return ErrInvalidMsgDeniedType
		}
	default:
		return ErrInvalidRPCRepyType
	}

	return nil
}

func (c *clientCodec) ReadResponseBody(result interface{}) error {

	if result == nil {
		// read and drain it out ?
		return nil
	}

	if _, err := xdr.Unmarshal(c.recordReader, &result); err != nil {
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
