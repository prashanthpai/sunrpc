// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"

	"github.com/rasky/go-xdr/xdr2"
)

// CreateReplyMessage will create an RPC reply message
func CreateReplyMessage(xid uint32, result interface{}) ([]byte, error) {
	var buf bytes.Buffer

	replyMessage := RPCMsgReply{
		Header: RPCMessageHeader{
			Xid:  xid,
			Type: Reply,
		},
		Stat: MsgAccepted,
		Areply: AcceptedReply{
			Stat: Success,
		},
	}

	if _, err := xdr.Marshal(&buf, replyMessage); err != nil {
		return nil, err
	}

	// Marshall and fill procedure-specific reply into the buffer
	if _, err := xdr.Marshal(&buf, result); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
