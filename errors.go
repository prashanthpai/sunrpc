// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"errors"
)

var (
	ErrInvalidFragmentSize     = errors.New("The RPC fragment size is invalid")
	ErrReadingRecordFragment   = errors.New("Error reading RPC record fragment from network")
	ErrWritingRecordFragment   = errors.New("Error writing RPC record fragment to network")
	ErrWritingRecord           = errors.New("Error writing RPC record to network")
	ErrCreatingRPCReplyMessage = errors.New("Could not create RPC message reply")
)
