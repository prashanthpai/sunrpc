// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"errors"
	"fmt"
)

// Internal errors
var (
	ErrInvalidFragmentSize = errors.New("The RPC fragment size is invalid")
)

// RPC errors

type ErrRPCMismatch struct {
	Low  uint32
	High uint32
}

func (e ErrRPCMismatch) Error() string {
	return fmt.Sprintf("RPC version not supported by server. Lowest and highest supported versions are %d and %d respectively", e.Low, e.High)
}

type ErrProgMismatch struct {
	Low  uint32
	High uint32
}

func (e ErrProgMismatch) Error() string {
	return fmt.Sprintf("Program version not supported. Lowest and highest supported versions are %d and %d respectively", e.Low, e.High)
}

var (
	ErrProgUnavail = errors.New("Remote server has not exported program")
	ErrProcUnavail = errors.New("Remote server has no such procedure")
	ErrGarbageArgs = errors.New("Remote procedure cannot decode params")
	ErrSystemErr   = errors.New("System error on remote server")
)

var (
	ErrInvalidRPCMessageType = errors.New("Invalid RPC message type received.")
	ErrInvalidRPCRepyType    = errors.New("Invalid RPC reply received. Reply type should be MsgAccepted or MsgDenied")
	ErrInvalidMsgDeniedType  = errors.New("Invalid MsgDenied reply. Possible values are RPCMismatch and AuthError")
	ErrInvalidMsgAccepted    = errors.New("Invalid MsgAccepted reply received")
	ErrAuthError             = errors.New("Remote server rejected identity of the caller")
)
