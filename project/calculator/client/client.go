package main

import (
	"context"
	"github.com/grpc-project02/project/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
)

func main() {
	log.Println("Calculator Client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}

	c := calculatorpb.NewCalculatorServiceClient(conn)
	// log.Printf("Created client: %f", c)
	// doServerStreaming(c)
	doClientStreaming(c)

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	requests := []*calculatorpb.ComputeAverageRequest{
		{
			Number: 10,
		},
		{
			Number: 15,
		},
		{
			Number: 15,
		},
		{
			Number: 19,
		},
		{
			Number: 21,
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		return
	}

	for _, request := range requests {
		err := stream.Send(request)
		if err != nil {
			return
		}
	}

	recv, err := stream.CloseAndRecv()
	if err != nil {
		return
	}
	log.Printf("Response from ComputeAverage: %v", recv.GetAverage())
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 21012315647892,
	}
	decomposition, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		return
	}
	for {
		recv, err := decomposition.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from PrimeNumberDecomposition: %v", recv.GetPrime())
	}
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.SumRequest{
		FirstNumber:  3,
		SecondNumber: 10,
	}
	sum, err := c.Sum(context.Background(), req)
	if err != nil {
		return
	}
	log.Printf("%d", sum.GetResult())
}
