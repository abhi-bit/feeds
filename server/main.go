package main

import (
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
func (s *server) SayHello(ctx context.Context, in *fb.HelloRequest) (*flatbuffers.Builder, error) {
	b := flatbuffers.NewBuilder(0)
	msg := b.CreateString("hello from flatbuf: " + string(in.Name()))

	fb.HelloResponseStart(b)
	fb.HelloResponseAddMessage(b, msg)
	b.Finish(fb.HelloResponseEnd(b))

	return b, nil
}

// SayHelloAgain says hello again
func (s *server) SayHelloAgain(ctx context.Context, in *fb.HelloRequest) (*flatbuffers.Builder, error) {
	b := flatbuffers.NewBuilder(0)
	msg := b.CreateString("hello again from flatbuf: " + string(in.Name()))

	fb.HelloResponseStart(b)
	fb.HelloResponseAddMessage(b, msg)
	b.Finish(fb.HelloResponseEnd(b))

	return b, nil
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
