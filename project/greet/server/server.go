package main

import (
	"context"
	"fmt"
	"github.com/grpc-project02/project/greet/greetpb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
}

func (*server) GreetManyTimes(request *greetpb.GreetManyTimesRequest, timesServer greetpb.GreetService_GreetManyTimesServer) error {
	log.Printf("Greet Many times function was invoked with %v", request)
	greeting := request.GetGreeting()
	for i := 0; i < 10; i++ {
		result := fmt.Sprintf("Hello %s number %d", greeting.GetFirstName(), i)
		response := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		err := timesServer.Send(response)
		if err != nil {
			log.Fatalf("While sending response, error occurred %s", err)
		}
	}
	return nil
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("Greet function was invoked with %v\n", req)
	greeting := req.GetGreeting()
	firstName := greeting.GetFirstName()
	lastName := greeting.GetLastName()
	result := fmt.Sprintf("Hello %s %s", firstName, lastName)
	res := &greetpb.GreetResponse{
		Result: result,
	}
	return res, nil
}

func main() {
	log.Println("Hello World")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	defer lis.Close()
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
