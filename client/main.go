package main

import (
	"log"

	fb "github.com/abhi-bit/feeds/protocol"
	flatbuffers "github.com/google/flatbuffers/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var name = "abhishek"
var addr = "0.0.0.0:50551"

func main() {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	c := fb.NewGreeterClient(conn)

	b := flatbuffers.NewBuilder(0)
	name := b.CreateString(name)
	fb.HelloRequestStart(b)
	fb.HelloRequestAddName(b, name)
	b.Finish(fb.HelloRequestEnd(b))

	r, err := c.SayHello(context.Background(), b)
	if err != nil {
		log.Fatalf("Failed to say hello: %v", err)
	}
	log.Printf("Hello Greeting: %s", string(r.Message()))

	r, err = c.SayHelloAgain(context.Background(), b)
	if err != nil {
		log.Fatalf("Failed to say hello again: %v", err)
	}
	log.Printf("Hello Greeting: %s", string(r.Message()))
}
