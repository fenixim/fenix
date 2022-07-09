package messages

import (
	"fenix/src/pb"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// Thrown when user tries to connect with a username that is already in use
type NickExists struct {
}

func (e NickExists) Error() string {
	return "Nickname in use"
}

// gRPC messsage service.  Implements protobuf MessageService
type Messenger struct {
	pb.UnimplementedMessageServiceServer
	channels map[string]chan *pb.MessageResponse
	nicks    map[string]string
}

// gRPC bidirectional stream.  Provides
func (m *Messenger) Stream(srv pb.MessageService_StreamServer) error {
	ctx := srv.Context()
	id, e := uuid.NewRandom()
	if e != nil {
		log.Fatalf("Error creating random UUID for user: %v", e)
	}

	nick := ctx.Value("nick").(string)

	if _, ok := m.nicks[nick]; ok {
		return NickExists{}
	}
	m.nicks[nick] = id.String()

	m.channels[id.String()] = make(chan *pb.MessageResponse)
	eChan := make(chan error)

	go func() {
		msgchan := m.channels[id.String()]
		for {
			select {
			case msg := <-msgchan:
				e := srv.Send(msg)
				if e != nil {
					eChan <- e
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			msg, e := srv.Recv()
			if e != nil {
				return
			}
			res := pb.MessageResponse{
				Message:   msg.Message,
				Username:  nick,
				Timestamp: time.Now().Unix(),
			}

			for _, channel := range m.channels {
				channel <- &res
			}
		}
	}()

	<-ctx.Done()

	return nil
}

func initMessenger() *Messenger {
	m := Messenger{}
	m.channels = make(map[string]chan *pb.MessageResponse)
	m.nicks = make(map[string]string)
	return &m
}

func Listen() {
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, initMessenger())
	log.Printf("Server running on port 5000")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
