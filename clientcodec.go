// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/rpc"

	"github.com/rasky/go-xdr/xdr2"
)

type clientCodec struct {
	conn         io.ReadWriteCloser
	recordReader io.Reader // Make this a map[uint64]io.Reader ?
}

// NewClientCodec returns a new rpc.ClientCodec using Sun RPC on conn
func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{conn: conn}
}

func (c *clientCodec) WriteRequest(req *rpc.Request, param interface{}) error {

	// req.Seq which is initialized (from 0) and incremented by net/rpc
	// for each call is unit64. But XID as per RFC should really be uint32.
	// TODO: This increment should be capped till maxOf(uint32) or we
	// should generate our own XID here.

	procedureID, ok := GetProcedureID(req.ServiceMethod)
	if !ok {
		// Reply with standard RPC error
	}

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

	// Marshall param/args of RPC
	_, err = xdr.Marshal(payload, &param)
	if err != nil {
		return err
	}

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

	return nil
}

func (c *clientCodec) ReadResponseBody(result interface{}) error {

	if result == nil {
		// drain it out ?
		_, _ = ioutil.ReadAll(c.recordReader)
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
