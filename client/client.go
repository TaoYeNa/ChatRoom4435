package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"os"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := chat.NewChatClient(conn)
	var content string
	// Contact the server and print out its response.
	if len(os.Args) > 1 {
		content = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SendMessage(ctx, &chat.Message{Event: "talk", Name: "Jacky", Content: content})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Message from : %s", r.Name)
}
