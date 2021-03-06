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
	"net"
	"time"
)

type server struct {
}

func (s *server) GreetWithDeadline(ctx context.Context, request *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	log.Printf("GreetWithDeadline function was invoked with %v\n", request)

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			log.Println("The client canceled the request")
			return nil, status.Errorf(codes.Canceled, "client canceled the request")
		}
		time.Sleep(1 * time.Second)
	}

	greeting := request.GetGreeting()
	firstName := greeting.GetFirstName()
	lastName := greeting.GetLastName()
	result := fmt.Sprintf("Hello %s %s", firstName, lastName)
	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}
	return res, nil
}

func (s *server) GreetEveryone(everyoneServer greetpb.GreetService_GreetEveryoneServer) error {
	log.Println("GreetEveryone function was invoked with a streaming request")

	for {
		request, err := everyoneServer.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v\n", err)
			return err
		}

		firstName := request.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "! "
		err = everyoneServer.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if err != nil {
			log.Fatalf("Error while sending client stream: %v\n", err)
			return err
		}
	}
}

func (*server) LongGreet(greetServer greetpb.GreetService_LongGreetServer) error {

	log.Println("LongGreet function was invoked with a streaming request")
	result := "Hello"
	for {
		req, err := greetServer.Recv()
		if err == io.EOF {
			return greetServer.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v\n", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result += firstName + "! "
	}
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
			return err
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

	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed loading certificates: %v", err)
		return
	}

	opts := grpc.Creds(creds)

	s := grpc.NewServer(opts)
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
