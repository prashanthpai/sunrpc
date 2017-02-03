// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
	"io"
	"log"
	"net/rpc"

	"github.com/rasky/go-xdr/xdr2"
)

type serverCodec struct {
	conn         io.ReadWriteCloser
	closed       bool
	recordReader io.Reader
}

// NewServerCodec returns a new rpc.ServerCodec using Sun RPC on conn
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{conn: conn}
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	// NOTE:
	// Errors returned by this function aren't relayed back to the client
	// as WriteResponse() isn't called. The net/rpc package will call
	// c.Close() when this function returns an error.

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		}
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshall RPC message
	var call RPCMsgCall
	_, err = xdr.Unmarshal(c.recordReader, &call)
	if err != nil {
		log.Println(err)
		return err
	}

	if call.Header.Type != Call {
		log.Println(ErrInvalidRPCMessageType)
		return ErrInvalidRPCMessageType
	}

	// Set req.Seq and req.ServiceMethod
	req.Seq = uint64(call.Header.Xid)
	procedureID := ProcedureID{call.Body.Program, call.Body.Version, call.Body.Procedure}
	procedureName, ok := GetProcedureName(procedureID)
	if ok {
		req.ServiceMethod = procedureName
	} else {
		// Due to our simpler map implementation, we cannot distinguish
		// between ErrProgUnavail and ErrProcUnavail
		log.Printf("%s: %+v\n", ErrProcUnavail, procedureID)
		return ErrProcUnavail
	}

	return nil
}

func (c *serverCodec) ReadRequestBody(funcArgs interface{}) error {

	if funcArgs == nil {
		return nil
	}

	if _, err := xdr.Unmarshal(c.recordReader, &funcArgs); err != nil {
		c.Close()
		return err
	}

	return nil
}

func (c *serverCodec) WriteResponse(resp *rpc.Response, result interface{}) error {

	if resp.Error != "" {
		// The remote function returned error (shouldn't really happen)
		log.Println(resp.Error)
	}

	var buf bytes.Buffer

	replyMessage := RPCMsgReply{
		Header: RPCMessageHeader{
			Xid:  uint32(resp.Seq),
			Type: Reply,
		},
		Stat: MsgAccepted,
		Areply: AcceptedReply{
			Stat: Success,
		},
	}

	if _, err := xdr.Marshal(&buf, replyMessage); err != nil {
		c.Close()
		return err
	}

	// Marshal and fill procedure-specific reply into the buffer
	if _, err := xdr.Marshal(&buf, result); err != nil {
		c.Close()
		return err
	}

	// Write buffer contents to network
	if _, err := WriteFullRecord(c.conn, buf.Bytes()); err != nil {
		c.Close()
		return err
	}

	return nil
}

func (c *serverCodec) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}
