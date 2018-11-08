package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	fb "github.com/abhi-bit/feeds/notifications"
	"github.com/couchbase/eventing/logging"
	flatbuffers "github.com/google/flatbuffers/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	addr    = "0.0.0.0:30000"
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randHostID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createAndMaintainFeed() error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
	if err != nil {
		logging.Errorf("Failed to connect: %v", err)
		return err
	}

	c := fb.NewUUIDServiceClient(conn)

	b := flatbuffers.NewBuilder(0)
	hostID := b.CreateString(randHostID(10))

	fb.UUIDResponseStart(b)
	fb.UserRequestAddID(b, hostID)
	b.Finish(fb.UUIDResponseEnd(b))

	stream, err := c.GetUUIDs(context.Background())
	if err != nil {
		logging.Errorf("GetUUIDs returned with err: %v", err)
		return err
	}

	err = stream.Send(b)
	if err != nil {
		logging.Errorf("Failed to send to server, err: %v", err)
		return err
	}

	resp, err := stream.Recv()
	for err == nil {
		fmt.Printf("%s\n", string(resp.UUID()))
		resp, err = stream.Recv()
	}

	err = stream.CloseSend()
	if err != nil {
		logging.Errorf("CloseSend returned err: %v", err)
		return err
	}

	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		switch line {
		case "exit":
			os.Exit(0)

		case "l":
			createAndMaintainFeed()
		}
	}
}
