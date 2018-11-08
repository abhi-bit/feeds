package main

import (
	"fmt"
	"log"

	fb "github.com/abhi-bit/feeds/protocol"
	flatbuffers "github.com/google/flatbuffers/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var name = "abhishek"
var addr = "0.0.0.0:50551"

// Single request and response based communication
func sayHello(c fb.GreeterClient) {
	logPrefix := "client::sayHello"

	b := flatbuffers.NewBuilder(0)
	n := b.CreateString(name)
	fb.HelloRequestStart(b)
	fb.HelloRequestAddName(b, n)
	b.Finish(fb.HelloRequestEnd(b))

	r, err := c.SayHello(context.Background(), b)
	if err != nil {
		log.Fatalf("%s failed to say hello: %v\n", logPrefix, err)
	}
	log.Printf("%s %s\n", logPrefix, string(r.Message()))
}

// Client makes a request and server keeps streaming responses
func getManyHellos(c fb.GreeterClient) {
	logPrefix := "client::getManyHellos"

	b := flatbuffers.NewBuilder(0)
	n := b.CreateString(name)
	fb.ManyHelloRequestStart(b)
	fb.ManyHelloRequestAddName(b, n)
	fb.ManyHelloRequestAddNumGreetings(b, 10)
	b.Finish(fb.ManyHelloRequestEnd(b))

	respC, err := c.GetManyHellos(context.Background(), b)
	if err != nil {
		log.Fatalf("%s failed to get many hellos: %v\n", logPrefix, err)
	}

	hResp, err := respC.Recv()
	if err != nil {
		log.Fatalf("%s failed receive message: %v\n", logPrefix, err)
	}

	for err == nil {
		log.Printf("%s data: %s\n", logPrefix, string(hResp.Message()))
		hResp, err = respC.Recv()
	}
}

func sayManyHellos(c fb.GreeterClient) {
	logPrefix := "client::sayManyHellos"

	respCS, err := c.SayManyHellos(context.Background())
	if err != nil {
		log.Fatalf("%s failed to say many hellos: %v", logPrefix, err)
	}

	b := flatbuffers.NewBuilder(0)

	for i := 0; i < 10; i++ {
		b.Reset()
		msg := fmt.Sprintf("%s:%d", name, i)
		n := b.CreateString(msg)
		fb.HelloRequestStart(b)
		fb.HelloRequestAddName(b, n)
		b.Finish(fb.HelloRequestEnd(b))

		err = respCS.Send(b)
		if err != nil {
			log.Printf("%s failed to send SayManyHellos: %v\n", logPrefix, err)
		} else {
			log.Printf("%s sent msg: %s\n", logPrefix, msg)
		}
	}

	respCSS, err := respCS.CloseAndRecv()
	if err != nil {
		log.Printf("%s err on close and recv: %v\n", logPrefix, err)
	} else {
		log.Printf("%s %s\n", logPrefix, string(respCSS.Message()))
	}
}

func main() {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	c := fb.NewGreeterClient(conn)

	sayHello(c)
	getManyHellos(c)
	sayManyHellos(c)
}
