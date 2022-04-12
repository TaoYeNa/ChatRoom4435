package main

import (
	"context"
	"fmt"
	chat "github.com/bartmika/simple-grpc/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	port = ":50051"
)

type server struct {
	chat.UnimplementedChatServer
}

func (s *server) SendMessage(ctx context.Context, in *chat.Message) (*chat.MessageReply, error) {
	log.Printf("Received: %v", in.Name, in.Event, in.Content)
	return &chat.MessageReply{Name: in.Name, Content: in.Content, Event: in.Event}, nil
}

func main() {
	fmt.Println("Server start")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	chat.RegisterChatServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
