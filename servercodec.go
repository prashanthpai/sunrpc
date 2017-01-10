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
	recordReader io.Reader
}

// NewServerCodec returns a new rpc.ServerCodec using Sun RPC on conn
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{conn: conn}
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	// FIXME:
	// Errors returned here aren't relayed back to client via WriteResponse
	// Will need some minor changes in net/rpc package to support this.

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		log.Println(err)
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshall RPC message
	var call RPCMsgCall
	bytesRead, err := xdr.Unmarshal(c.recordReader, &call)
	if err != nil {
		log.Println(err)
		return err
	}

	if call.Header.Type != Call {
		log.Println(ErrInvalidRPCMessageType)
		return ErrInvalidRPCMessageType
	}

	// TODO: Remove
	log.Printf("CallPayload: %+v PayloadSize: %d ParamSize: %d", call, bytesRead, len(record)-bytesRead)

	// Set req.Seq and req.ServiceMethod
	req.Seq = uint64(call.Header.Xid)
	procedureID := ProcedureID{call.Body.Program, call.Body.Version, call.Body.Procedure}
	procedureName, ok := GetProcedureName(procedureID)
	if ok {
		req.ServiceMethod = procedureName
	} else {
		// Due to our simpler map implementation, we cannot distinguish
		// between ErrProgUnavail and ErrProcUnavail
		log.Println(ErrProcUnavail)
		return ErrProcUnavail
	}

	return nil
}

func (c *serverCodec) ReadRequestBody(funcArgs interface{}) error {

	if funcArgs == nil {
		return nil
	}

	if _, err := xdr.Unmarshal(c.recordReader, &funcArgs); err != nil {
		return err
	}

	return nil
}

func (c *serverCodec) WriteResponse(resp *rpc.Response, result interface{}) error {

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
