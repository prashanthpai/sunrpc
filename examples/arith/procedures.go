package main

// Args is a struct that contains arguments to be sent to the remote procedure
type Args struct {
	A, B int32
}

// Arith is a placeholder type
type Arith int32

// Multiply multiplies two numbers contained in Args struct and places the
// result in reply argument.
func (t *Arith) Multiply(args *Args, reply *int32) error {
	*reply = args.A * args.B
	return nil
}

// Add adds two numbers contained in Args struct and places the result in
// reply argument.
func (t *Arith) Add(args *Args, reply *int32) error {
	*reply = args.A + args.B
	return nil
}
