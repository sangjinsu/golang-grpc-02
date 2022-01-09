package main

import (
	"context"
	"fmt"
	"github.com/grpc-project02/project/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("Hello I'm a client")
	certFile := "ssl/ca.crt"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatalf("Error while loading CA trust certificate: %v", err)
		return
	}
	opts := grpc.WithTransportCredentials(creds)
	conn, err := grpc.Dial("localhost:50051", opts)
	defer conn.Close()
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	c := greetpb.NewGreetServiceClient(conn)
	// log.Printf("Created client: %f", c)
	doUnary(c)
	//serverStream(c)
	//clientStream(c)
	//doBiDiStream(c)
	//doUnaryWithDeadline(c, 1*time.Second)
	//doUnaryWithDeadline(c, 5*time.Second)
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	log.Println("Starting to do a UnaryWithDeadline RPC..")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jinsu",
			LastName:  "Sang",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	response, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Println("Timeout was hit! Deadline was exceeded")
			} else {
				log.Printf("unexpected error %v", statusErr)
			}
			return
		} else {
			log.Fatalf("error while calling GreetWithDeadline RPC: %v\n", err)
		}
	}
	log.Printf("Response from GreetWithDeadline: %v\n", response.GetResult())
}

func doBiDiStream(c greetpb.GreetServiceClient) {
	log.Println("Starting to do a BiDi Streaming RPC...")

	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jinsu",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Wanhee",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Seongho",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Hana",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Hyeonbin",
			},
		},
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream")
		return
	}

	var wg sync.WaitGroup
	for _, request := range requests {
		wg.Add(1)
		go func(request *greetpb.GreetEveryoneRequest) {
			defer wg.Done()
			log.Printf("Sending message %v\n", request)
			stream.Send(request)
		}(request)
	}

	wg.Wait()
	stream.CloseSend()

	results := make(chan string)
	go func() {
		defer close(results)
		for {
			recv, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving %v", err)
			}
			results <- recv.GetResult()

		}
	}()

	for result := range results {
		log.Printf("Received %v", result)
	}
}

func clientStream(c greetpb.GreetServiceClient) {
	log.Println("Starting to do a Client Streaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jinsu",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Wanhee",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Seongho",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Hana",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Hyeonbin",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}

	// we iterate over our slice and send each message individually
	for _, request := range requests {
		err := stream.Send(request)
		if err != nil {
			log.Fatalln("error while sending request")
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from LongGreet: %v", err)
	}

	log.Printf("LongGreet Response: %v\n", response)
}

func serverStream(c greetpb.GreetServiceClient) {
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
