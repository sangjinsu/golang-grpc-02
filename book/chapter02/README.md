# gRPC 시작

```protobuf
syntax = "proto3";
package ecommerce;

option go_package = "ecommerce/ecommercepb";
service ProductInfo{
  rpc addProduct(Product) returns (ProductID);
  rpc getProduct(ProductID) returns (Product);
}

message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  float price = 4;
}

message ProductID {
  string value = 1;
}
```

### Server

```go
package main

import (
	"context"
	"github.com/grpc-project02/book/chapter02/ecommerce/ecommercepb"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

type server struct {
	productMap map[string]*ecommercepb.Product
}

func (s *server) AddProduct(ctx context.Context, product *ecommercepb.Product) (*ecommercepb.ProductID, error) {
	out, err := uuid.NewV4()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating Product ID", err)
	}

	product.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*ecommercepb.Product)
	}

	s.productMap[product.GetId()] = product
	return &ecommercepb.ProductID{Value: product.Id}, status.New(codes.OK, "").Err()
}

func (s *server) GetProduct(ctx context.Context, id *ecommercepb.ProductID) (*ecommercepb.Product, error) {
	product, exists := s.productMap[id.Value]
	if exists {
		return product, status.New(codes.OK, "").Err()
	}
	return nil, status.Errorf(codes.NotFound, "Product does not exist", id.Value)
}

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	ecommercepb.RegisterProductInfoServer(s, &server{})

	log.Println("Starting gRPC listener on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("faild to serve: %v\n", err)
	}
}

```

### Client

```go
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

```

