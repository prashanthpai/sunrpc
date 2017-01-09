// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
	"fmt"
	"io"
	"net/rpc"

	"github.com/rasky/go-xdr/xdr2"
)

type serverCodec struct {
	conn         io.ReadWriteCloser
	recordReader io.Reader // Make this a map[uint64]io.Reader ?
}

// NewServerCodec returns a new rpc.ServerCodec using Sun RPC on conn
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{conn: conn}
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshall RPC message
	var payload RPCMsgCall
	bytesRead, err := xdr.Unmarshal(c.recordReader, &payload)
	if err != nil {
		return err
	}

	// TODO: Remove
	fmt.Printf("CallPayload: %+v\nPayloadSize: %d\nParamSize: %d\n\n", payload, bytesRead, len(record)-bytesRead)

	// Set req.Seq and req.ServiceMethod
	req.Seq = uint64(payload.Header.Xid)
	procedureID := ProcedureID{payload.Body.Program, payload.Body.Version, payload.Body.Procedure}
	procedureName, ok := GetProcedureName(procedureID)
	if ok {
		req.ServiceMethod = procedureName
	} else {
		// Reply with standard RPC error
	}

	return nil
}

func (c *serverCodec) ReadRequestBody(funcArgs interface{}) error {

	if funcArgs == nil {
		// read and drain it out ?
		return nil
	}

	if _, err := xdr.Unmarshal(c.recordReader, &funcArgs); err != nil {
		return err
	}

	return nil
}

func (c *serverCodec) WriteResponse(resp *rpc.Response, result interface{}) error {

	var buf bytes.Buffer

	// TODO: Error handling and error reply

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
		return err
	}

	// Marshal and fill procedure-specific reply into the buffer
	if _, err := xdr.Marshal(&buf, result); err != nil {
		return err
	}

	// Write buffer contents to network
	if _, err := WriteFullRecord(c.conn, buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func (c *serverCodec) Close() error {
	return c.conn.Close()
}
