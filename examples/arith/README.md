**C code**

```sh
# gcc -Wall -o c-client c-client.c arith_clnt.c arith_xdr.c arith.h
# gcc -Wall -o c-server c-server.c arith_svc.c arith_xdr.c arith.h
```

**Go code**

```sh
# go build go-server.go procedures.go
# go build go-client.go procedures.go
```
You should be able to use C client with Go server and
also Go client with C server.
