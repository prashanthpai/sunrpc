// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"

	"github.com/davecgh/go-xdr/xdr2"
)

// CreateReplyMessage will create an RPC reply message
func CreateReplyMessage(xid uint32, result interface{}) ([]byte, error) {
	var buf bytes.Buffer

	rpcHeader := RPCMessage{
		Xid:  xid,
		Type: Reply,
	}

	if _, err := xdr.Marshal(&buf, rpcHeader); err != nil {
		return nil, err
	}

	if _, err := xdr.Marshal(&buf, ReplyBody{Stat: MsgAccepted}); err != nil {
		return nil, err
	}

	if _, err := xdr.Marshal(&buf, ReplyPayload{Stat: Success}); err != nil {
		return nil, err
	}

	// Marshall and fill actual reply into the buffer
	if _, err := xdr.Marshal(&buf, result); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
