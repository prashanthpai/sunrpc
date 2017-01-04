package main

// Example
type Args struct {
	A, B int32
}

type Arith int32

func (t *Arith) Multiply(args *Args, reply *int32) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Add(args *Args, reply *int32) error {
	*reply = args.A + args.B
	return nil
}
