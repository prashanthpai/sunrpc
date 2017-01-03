// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
	"fmt"
	"io"
	"net/rpc"

	"github.com/davecgh/go-xdr/xdr2"
)

type serverCodec struct {
	conn         io.ReadWriteCloser
	recordReader io.Reader
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		if err != io.EOF {
			fmt.Println("ReadFullRecord() failed: ", err)
		}
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshall RPC message
	var payload CallPayload
	bytesRead, err := xdr.Unmarshal(c.recordReader, &payload)
	if err != nil {
		fmt.Println("xdr.Unmarshal() failed: ", err)
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
		return nil
	}

	_, err := xdr.Unmarshal(c.recordReader, &funcArgs)
	if err != nil {
		return err
	}

	return nil
}

func (c *serverCodec) WriteResponse(resp *rpc.Response, result interface{}) error {

	// The net/rpc package specifies Request.Seq and Response.Seq as uint64
	// but the XID shall always be uint32, so this should be okay.
	xid := uint32(resp.Seq)

	rpcMessage, err := CreateReplyMessage(xid, result)
	if err != nil {
		return ErrCreatingRPCReplyMessage
	}

	bytesWritten, err := WriteFullRecord(c.conn, rpcMessage)
	if err != nil || (bytesWritten != int64(len(rpcMessage))) {
		return ErrWritingRecord
	}

	return nil
}

func (c *serverCodec) Close() error {
	return c.conn.Close()
}

// NewServerCodec returns a new rpc.ServerCodec using Sun RPC on conn
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{conn: conn}
}
