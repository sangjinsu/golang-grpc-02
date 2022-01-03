package main

import (
	"context"
	"github.com/grpc-project02/project/calculator/calculatorpb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
}

func (*server) PrimeNumberDecomposition(request *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	log.Printf("Received PrimeNumberDecomposition RPC: %v", request)
	number := request.GetNumber()
	k := int64(2)

	for number > 1 {
		if number%k == 0 {
			number /= k
			response := &calculatorpb.PrimeNumberDecompositionResponse{
				Prime: k,
			}
			err := stream.Send(response)
			if err != nil {
				log.Fatalf("While sending response, error occurred %s", err)
			}
		} else {
			k++
		}
	}
	return nil
}

func (*server) Sum(ctx context.Context, request *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	log.Printf("Recieved Sum RPC: %v", request)
	res := &calculatorpb.SumResponse{
		Result: request.GetFirstNumber() + request.GetSecondNumber(),
	}
	return res, nil
}

func main() {
	log.Println("Calculator Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	defer lis.Close()
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}