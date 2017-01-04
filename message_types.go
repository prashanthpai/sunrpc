// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

const RPCVersionSupported = 2

// As per XDR (RFC 4506):
// Enumerations have the same representation as 32 bit signed integers.

type MsgType int32

const (
	Call  MsgType = 0
	Reply MsgType = 1
)

type ReplyStat int32

const (
	MsgAccepted ReplyStat = 0
	MsgDenied   ReplyStat = 1
)

type AcceptStat int32

const (
	Success AcceptStat = iota
	ProgUnavail
	ProgMismatch
	ProcUnavail
	GarbageArgs
	SystemErr
)

type RejectStat int32

const (
	RPCMismatch RejectStat = 0
	AuthError   RejectStat = 1
)

type AuthStat int32

const (
	AuthOk AuthStat = iota
	AuthBadcred
	AuthRejectedcred
	AuthBadverf
	AuthRejectedVerf
	AuthTooweak
	AuthInvalidresp
	AuthFailed
	AuthKerbGeneric
	AuthTimeexpire
	AuthTktFile
	AuthDecode
	AuthNetAddr
	RPCsecGssCredproblem
	RPCsecGssCtxproblem
)

type AuthFlavour int32

const (
	AuthNone AuthFlavour = iota
	AuthSys
	AuthShort
	AuthDh
	RPCsecGss = 6
)

type OpaqueAuth struct {
	Flavour AuthFlavour
	Body    []byte
}

type RPCMessageHeader struct {
	Xid  uint32
	Type MsgType
}

type CallBody struct {
	RPCVersion uint32
	Program    uint32
	Version    uint32
	Procedure  uint32
	Cred       OpaqueAuth
	Vers       OpaqueAuth
}

type RPCMsgCall struct {
	Header RPCMessageHeader
	Body   CallBody
	// procedure-specific parameters start here
}

// Section 9 of RFC 5531  uses discriminated union to specify procedure reply.
// Section 4.15 of RFC 4506 defines XDR representation of discriminated union.
// However, this is not provided by xdr package in go yet. As a simple
// workaround, created multiple reply templates here.

// MsgAccepted cases

type MsgAcceptedSuccess struct {
	Header RPCMessageHeader
	Type   ReplyStat // MsgAccepted
	Verf   OpaqueAuth
	Stat   AcceptStat // Success
	// procedure-specific results start here
}

type MsgAcceptedProgMismatch struct {
	Header       RPCMessageHeader
	Type         ReplyStat // MsgAccepted
	Verf         OpaqueAuth
	Stat         AcceptStat // ProgMismatch
	MismatchInfo struct {
		Low  uint32
		High uint32
	}
}

type MsgAcceptedOtherError struct {
	Header RPCMessageHeader
	Type   ReplyStat // MsgAccepted
	Verf   OpaqueAuth
	Stat   AcceptStat // ProgUnavail, ProcUnavail, GarbageArgs, SystemErr
}

// MsgDenied Cases

type MsgDeniedRPCMismatch struct {
	Header       RPCMessageHeader
	Type         ReplyStat  // MsgDenied
	Stat         RejectStat // RPCMismatch
	MismatchInfo struct {
		Low  uint32
		High uint32
	}
}

type MsgDeniedAuthError struct {
	Header   RPCMessageHeader
	Type     ReplyStat  // MsgDenied
	Stat     RejectStat // AuthError
	AuthStat AuthStat
}
