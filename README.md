# sunrpc

This project aims to implement ONC RPC (Sun RPC) as described in
[RFC 5531](https://tools.ietf.org/html/rfc5531) in Go lang, primarily to be
consumed as a [ServerCodec](https://golang.org/pkg/net/rpc/#ServerCodec).

The initial goal here is limited to enabling existing projects written in C
and uses Sun RPC to be able to communicate with a server written in Go without
the need for C projects to change their existing code.
