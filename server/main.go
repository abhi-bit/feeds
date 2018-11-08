package main

import (
	"fmt"
	"log"
	"net"

	fb "github.com/abhi-bit/feeds/protocol"
	flatbuffers "github.com/google/flatbuffers/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50551"
)

type server struct{}

// SayHello says hello
func (s *server) SayHello(ctx context.Context, req *fb.HelloRequest) (*flatbuffers.Builder, error) {
	logPrefix := "server::SayHello"

	b := flatbuffers.NewBuilder(0)
	msg := b.CreateString(fmt.Sprintf("%s hello %s ", logPrefix, string(req.Name())))

	fb.HelloResponseStart(b)
	fb.HelloResponseAddMessage(b, msg)
	b.Finish(fb.HelloResponseEnd(b))

	return b, nil
}

// GetManyHellos receives many hellos
func (s *server) GetManyHellos(req *fb.ManyHelloRequest, resp fb.Greeter_GetManyHellosServer) error {
	logPrefix := "server::GetManyHellos"

	b := flatbuffers.NewBuilder(0)
	msg := b.CreateString(fmt.Sprintf("%s hello %s", logPrefix, string(req.Name())))

	fb.HelloResponseStart(b)
	fb.HelloResponseAddMessage(b, msg)
	b.Finish(fb.HelloResponseEnd(b))

	for i := 0; i < int(req.NumGreetings()); i++ {
		err := resp.Send(b)
		if err != nil {
			log.Printf("%s Failed to send msg: %v\n", logPrefix, err)
			return err
		}
		log.Printf("%s Sent msg to client, id: %d\n", logPrefix, i)
	}

	return nil
}

// SayManyHellos says many hellos
func (s *server) SayManyHellos(resp fb.Greeter_SayManyHellosServer) error {
	logPrefix := "server::SayManyHellos"

	var msgCount int

	req, err := resp.Recv()
	for err == nil {
		msgCount++
		log.Printf("%s received request payload: %s\n", logPrefix, string(req.Name()))
		req, err = resp.Recv()
	}

	b := flatbuffers.NewBuilder(0)
	msg := b.CreateString(fmt.Sprintf("%s received %d messages", logPrefix, msgCount))

	fb.HelloResponseStart(b)
	fb.HelloResponseAddMessage(b, msg)
	b.Finish(fb.HelloResponseEnd(b))

	err = resp.SendAndClose(b)
	if err != nil {
		log.Printf("%s SendAndClose err: %v", logPrefix, err)
	}

	return err
}

// ChatterManyHellos is handler for bidirectional communication
func (s *server) ChatterManyHellos(resp fb.Greeter_ChatterManyHellosServer) error {
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}))
	fb.RegisterGreeterServer(s, &server{})

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
