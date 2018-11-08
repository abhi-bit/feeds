package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	fb "github.com/abhi-bit/feeds/notifications"
	"github.com/couchbase/eventing/logging"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

var (
	port          string
	id            int
	numServers    int
	sleepDuration int
)

func init() {
	logging.SetLogLevel(logging.Info)
}

func (s *server) GetUUIDs(resp fb.UUIDService_GetUUIDsServer) error {
	logPrefix := "server::GetUUIDs"

	currPort, err := strconv.Atoi(port)
	if err != nil {
		logging.Errorf("%s failed to convert current port, err: %v", logPrefix, err)
		return err
	}

	serverPort := currPort - id
	var conns []fb.UUIDServiceClient

	for i := 0; i < numServers; i++ {
		conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", serverPort+i),
			grpc.WithInsecure(), grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
		if err != nil {
			logging.Fatalf("%s failed to connect to server, err: %v", logPrefix, err)
			return err
		}
		logging.Infof("%s connected to server: %d", logPrefix, serverPort+i)

		c := fb.NewUUIDServiceClient(conn)
		conns = append(conns, c)
	}

	userReq, err := resp.Recv()
	if err != nil {
		logging.Errorf("%s failed to receive user request, err: %v", logPrefix, err)
		return err
	}
	logging.Infof("%s got request from node: %s", logPrefix, string(userReq.ID()))

	var wg sync.WaitGroup
	connMutex := &sync.Mutex{}

	wg.Add(numServers)
	for i := 0; i < numServers; i++ {
		go func(conn fb.UUIDServiceClient, mu *sync.Mutex, wg *sync.WaitGroup) {
			logPrefix := "::readFromPeers"

			defer wg.Done()

			b := flatbuffers.NewBuilder(0)
			id := b.CreateString(string(userReq.ID()))

			fb.UserRequestStart(b)
			fb.UserRequestAddID(b, id)
			b.Finish(fb.UserRequestEnd(b))

			stream, err := conn.GetUUIDsInternal(context.Background(), b)
			if err != nil {
				logging.Errorf("%s read from peers, err: %v", logPrefix, err)
				return
			}

			remoteResp, err := stream.Recv()
			for err == nil {
				b.Reset()
				rUUID := b.CreateString(string(remoteResp.UUID()))

				fb.UUIDResponseStart(b)
				fb.UserRequestAddID(b, rUUID)
				b.Finish(fb.UUIDResponseEnd(b))

				mu.Lock()
				err = resp.Send(b)
				if err != nil {
					logging.Errorf("%s failed to send client, err: %v", logPrefix, err)
					mu.Unlock()
					return
				}
				mu.Unlock()
				remoteResp, err = stream.Recv()
			}
		}(conns[i], connMutex, &wg)
	}

	wg.Wait()

	return nil
}

func (s *server) GetUUIDsInternal(req *fb.UserRequest, resp fb.UUIDService_GetUUIDsInternalServer) error {
	logPrefix := "server::GetUUIDsInternal"

	b := flatbuffers.NewBuilder(0)

	for {
		b.Reset()
		uid := fmt.Sprintf("%s_%s", port, uuid.New().String())
		uidOffset := b.CreateString(uid)

		fb.UUIDResponseStart(b)
		fb.UUIDResponseAddUUID(b, uidOffset)
		b.Finish(fb.UUIDResponseEnd(b))

		err := resp.Send(b)
		if err != nil {
			logging.Errorf("%s error encountered: %v", logPrefix, err)
			return err
		}
		logging.Debugf("%s sent to caller %s", logPrefix, uid)

		if sleepDuration > 0 {
			time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
		}
	}
}

func main() {
	port = os.Args[1]
	val, err := strconv.Atoi(os.Args[2])
	if err != nil {
		logging.Fatalf("Failed to convert id, err: %v", err)
	}
	id = val

	val, err = strconv.Atoi(os.Args[3])
	if err != nil {
		logging.Fatalf("Failed to convert num servers, err: %v", err)
	}
	numServers = val

	val, err = strconv.Atoi(os.Args[4])
	if err != nil {
		logging.Fatalf("Failed to convert sleep duration, err: %v", err)
	}
	sleepDuration = val

	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		logging.Fatalf("Failed to listen: %v", err)
	} else {
		logging.Infof("Listening on port: %s", port)
	}

	s := grpc.NewServer(grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}))
	fb.RegisterUUIDServiceServer(s, &server{})

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		logging.Fatalf("Failed to serve: %v", err)
	}
}
