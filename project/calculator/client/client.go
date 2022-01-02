package main

import (
	"context"
	"github.com/grpc-project02/project/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	log.Println("Hello I'm a client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	c := calculatorpb.NewCalculatorServiceClient(conn)
	// log.Printf("Created client: %f", c)

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
