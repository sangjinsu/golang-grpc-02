package main

import (
	"context"
	"fmt"
	"github.com/grpc-project02/project/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
)

func main() {
	fmt.Println("Hello I'm a client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	c := greetpb.NewGreetServiceClient(conn)
	// log.Printf("Created client: %f", c)
	//doUnary(c)
	log.Println("Starting to do a Server Streaming RPC...")
	req := &greetpb.GreetManyTimesRequest{Greeting: &greetpb.Greeting{
		FirstName: "Jinsu",
		LastName:  "Sang",
	}}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}

	for {
		response, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", response.GetResult())
	}

}

func doUnary(c greetpb.GreetServiceClient) {
	log.Println("Starting to do a Unary RPC..")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jinsu",
			LastName:  "Sang",
		},
	}
	response, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v\n", err)
	}
	log.Printf("Response from Greet: %v\n", response.GetResult())
}
