package main

import (
	"context"
	"github.com/grpc-project02/book/chapter02/ecommerce/ecommercepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error while connecting: %v", err)
	}

	defer conn.Close()
	c := ecommercepb.NewProductInfoClient(conn)

	name := "Samsung A70"
	description := "Meet Samsung A70."
	price := float32(1000)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	product, err := c.AddProduct(ctx, &ecommercepb.Product{Name: name, Description: description, Price: price})
	if err != nil {
		log.Fatalf("Error while adding product: %v", err)
	}
	log.Printf("Product ID: %s added successfully", product.Value)

	getProduct, err := c.GetProduct(ctx, &ecommercepb.ProductID{Value: product.Value})
	if err != nil {
		log.Fatalf("Error while getting product: %v", err)
	}
	log.Printf("Product: %v", getProduct.String())
}
