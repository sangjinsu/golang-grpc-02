# 고급 기능

## 인터셉터

인터셉터라는 확장 메커니즘을 사용해 로깅, 인증, 매트릭 등과 같은 특정 요구 사항 충족을 위해 RPC 실행을 가로채거나 클라이언트와 서버 gRPC 애플리케이션에서 인터셉터를 구현하고 설치하기 위한 간단한 API를 제공한다 

### 서버 측 인터셉터 

클라이언트가 gRPC 서비스의 원격 메서드를 호출할 때 서버에서 인터셉터를 사용해 원격 메서드 실행 전에 공통 로직을 실행 할 수 있다 

원격 메서드를 호출 하기 전에 인증과 같은 특정 기능을 적용해야 할 때 필요하다 

![Server-side interceptors ](README.assets/grpc_0501.png)

#### 단일 인터셉터

서버에서 gRPC 서비스의 단일 RPC 가로채려면 gRPC 서버에 단일 인터셉터를 구현해야 한다.

```go
func orderUnaryServerInterceptor(ctx context.Context, req interface{}, 
                                 info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // 전처리 로직
    // 인자로 넘겨진 info 를 통해 현재 RPC 호출에 대한 정보를 얻는다 
    log.Println(info.FullMethod)
    
    // 단일 RPC 의 정상 실행을 완료하고자 핸들러를 호출한다 
    m, err := handler(ctx, req)
    
    // 후처리 로직
    log.Printf("Post Proc Message : %s", m)
    
    return m, err
}

func main() {
    s := grpc.NewServer(
        grpc.UnaryInterceptor(orderUnaryServerInterceptor)
    )
}
```

1. 전처리
2. RPC 메서드 호출
3. 후처리 



#### 스트리밍 인터셉터

grpc 서버가 처리하는 모든 스트리밍 RPC 호출을 인터셉트한다 

스트리밍 인터셉터는 전처리 단계와 스트림 동작 인터셉트 단계를 포함한다 

```go
func(srv interface{}, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error
```

```go
type wrappedStream struct {
    grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m interfdace{}) error {
    return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
    return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
    return &wrappedStream{s}
}

func orderServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
                                  handler grpc.StreamHandler) error {
    err := handler(srv, newWrappedStream(ss))
    if err != nil {
        log.Printf("RPC failed with error %v", err)
    }
    return err 
}

s := grpc.NewServer(
    grpc.StreamInterceptor(orderServerStreamInterceptor)
)
```

### 

### 클라이언트 측 인터셉터

클라이언트가 gRPC 서비스의 원격 메서드를 호출하고자 RPC를 할 때 클라이언트에서 해당 RPC 호출을 가로챌 수 있다 

![Client-side Interceptors](README.assets/grpc_0502.png)



#### 단일 인터셉터

```go
func(ctx context.Context, method string, req, reply interface{},
         cc *ClientConn, invoker UnaryInvoker, opts ...CallOption) error
```

```go
func orderUnaryClientInterceptor(
	ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Preprocessor phase 전처리 단계
	log.Println("Method : " + method) 

	// Invoking the remote method 원격 메서드 호출
	err := invoker(ctx, method, req, reply, cc, opts...) 

	// Postprocessor phase 후처리 단계
	log.Println(reply) 

	return err 
}
...

func main() {
	// Setting up a connection to the server.
    // 서버로의 연결을 설정한다
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor)) 
...
```



#### 스트리밍 인터셉터

클라이언트 측 스트리밍 인터셉터는 gRPC 클라이언트가 처리하는 모든 스트리밍 RPC 호출을 인터셉트하며 구현은 서버 측 구현과 매우 유사하다

```go
func(ctx context.Context, desc *StreamDesc, cc *ClientConn,
                                      method string, streamer Streamer,
                                      opts ...CallOption) (ClientStream, error)
```

```go
func clientStreamInterceptor(
	ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string,
	streamer grpc.Streamer, opts ...grpc.CallOption)
        (grpc.ClientStream, error) {
	log.Println("======= [Client Interceptor] ", method) 
	s, err := streamer(ctx, desc, cc, method, opts...) 
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil 
}


type wrappedStream struct { 
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error { 
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error { 
	return w.ClientStream.SendMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

...

func main() {
	// Setting up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithStreamInterceptor(clientStreamInterceptor)) 
...
```



### 데드라인

데드라인과 타임아웃은 분산 컴퓨팅에서 일반적으로 사용되는 패턴이다 

타임아웃은 클라이언트 애플리케이션이 RPC가 완료될 때까지 에러와 함께 끝나기 전에 얼마나 기다리는지를 지정한다 

일반적으로는 기간으로 지정되며 각 클라이언트에 개별로 적용된다 

하나의 요청이 하나 이상의 서비스를 함께 묶는 여러 다운스트림 RPC로 구성되는 예시는 각 서비스 호출 마다 개별 RPC 를 기준으로 타임아웃을 적용할 수 있지만, 요청 전체 수명 주기에는 직접 적용할 수 없다 -> 데드라인 사용



데드라인은 요청 시작 기준으로 특정시간으로 표현되며 여러 서비스 호출에 걸쳐 적용된다 

요청을 시작하는 애플리케이션이 데드라인을 설정하면 전체 요청 체인은 데드라인까지 응답해야 한다 

gRPC 통신은 네트워크를 통해 이뤄지므로 RPC 호출과 응답 사이 지연이 발생될 수 있다 

gRPC 서비스 자체 비즈니스 로직에 따라 응답하는데 더 많은 시간이 걸릴 수 있다 

데드라인을 사용하지 않고 클라이언트 애플리케이션을 개발하면 시작된 RPC 요청에 대한 응답을 무한정 기다르며 모든 진행 중인 요청에 대해 리소스가 계속 유지된다 

서비스는 물론 클라이언트도 리소스가 부족해질 수 있으므로 서비스 대기 시간이 길어지게 되며 결국 전체 gRPC 서비스가 중단될 수도 있다 

![Using Deadlines when calling services](README.assets/grpc_0503.png)

```go
conn, err := grpc.Dial(address, grpc.WithInsecure())
if err != nil {
    log.Fatalf("did not connect: %v", err)
}
defer conn.Close()
client := pb.NewOrderManagementClient(conn)

clientDeadline := time.Now().Add(
    time.Duration(2 * time.Second))
ctx, cancel := context.WithDeadline(
    context.Background(), clientDeadline) 1

defer cancel()

// Add Order
order1 := pb.Order{Id: "101",
    Items:[]string{"iPhone XS", "Mac Book Pro"},
    Destination:"San Jose, CA",
    Price:2300.00}
res, addErr := client.AddOrder(ctx, &order1) 2

if addErr != nil {
    got := status.Code(addErr) 3
    log.Printf("Error Occured -> addOrder : , %v:", got) 4
} else {
    log.Print("AddOrder Response -> ", res.Value)
}
```



### 취소 처리 

클라이언트와 서버 애플리케이션 사이의 gRPC 연결에서 클라이언트와 서버는 모두 통신 성공 여부를 독립적이고 개별적으로 결정한다

#### gRPC 취소 처리

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) 1


streamProcOrder, _ := client.ProcessOrders(ctx) 2
_ = streamProcOrder.Send(&wrapper.StringValue{Value:"102"}) 3
_ = streamProcOrder.Send(&wrapper.StringValue{Value:"103"})
_ = streamProcOrder.Send(&wrapper.StringValue{Value:"104"})

channel := make(chan bool, 1)

go asncClientBidirectionalRPC(streamProcOrder, channel)
time.Sleep(time.Millisecond * 1000)

// Canceling the RPC
cancel() 4
log.Printf("RPC Status : %s", ctx.Err()) 5

_ = streamProcOrder.Send(&wrapper.StringValue{Value:"101"})
_ = streamProcOrder.CloseSend()

<- channel

func asncClientBidirectionalRPC (
    streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan bool) {
...
		combinedShipment, errProcOrder := streamProcOrder.Recv()
		if errProcOrder != nil {
			log.Printf("Error Receiving messages %v", errProcOrder) 6
...
}
```



### 에러 처리

gRPC 를 호출하면 클라이언트는 성공 산태의 응답을 받거나 에러 상태를 갖는 에러를 받는다 

클라이언트 애플리케이션은 발생 가능한 모든 에러와 에러 상태를 처리하는 방식으로 작성해야 한다 

에러가 발생하면 gRPC 는 에러 상태의 자세한 정보를 제공하는 선택적 에러 메시지와 함께 에러 상태 코드를 반환한다 

상태 객체는 다른 언어에 대한 모든  gRPC 구현에 공통적인 정수 코드와 문자열 메시지로 구성된다 

[상태 코드](https://learning.oreilly.com/library/view/grpc-up-and/9781492058328/ch05.html#idm46536639832888)

#### 서버에서의 에러 생성과 전파

```go
if orderReq.Id == "-1" { 1
    log.Printf("Order ID is invalid! -> Received Order ID %s",
        orderReq.Id)

    errorStatus := status.New(codes.InvalidArgument,
        "Invalid information received") 2
    ds, err := errorStatus.WithDetails( 3
        &epb.BadRequest_FieldViolation{
            Field:"ID",
            Description: fmt.Sprintf(
                "Order ID received is not valid %s : %s",
                orderReq.Id, orderReq.Description),
        },
    )
    if err != nil {
        return nil, errorStatus.Err()
    }

    return nil, ds.Err() 4
    }
    ...
```

#### 클라이언트에서의 에러 처리

```go
order1 := pb.Order{Id: "-1",
	Items:[]string{"iPhone XS", "Mac Book Pro"},
	Destination:"San Jose, CA", Price:2300.00} 1
res, addOrderError := client.AddOrder(ctx, &order1) 2


if addOrderError != nil {
	errorCode := status.Code(addOrderError) 3
	if errorCode == codes.InvalidArgument { 4
		log.Printf("Invalid Argument Error : %s", errorCode)
		errorStatus := status.Convert(addOrderError) 5
		for _, d := range errorStatus.Details() {
			switch info := d.(type) {
			case *epb.BadRequest_FieldViolation: 6
				log.Printf("Request Field Invalid: %s", info)
			default:
				log.Printf("Unexpected error type: %s", info)
			}
		}
	} else {
		log.Printf("Unhandled error : %s ", errorCode)
	}
} else {
	log.Print("AddOrder Response -> ", res.Value)
}
```

### 멀티플렉싱

gRPC를 사용하면 동일한 gRPC 서버에서 여러 gRPC 서비스를 실행할 수 있고 여러 gRPC 클라이언트 스텁에 동일한  gRPC 클라이언트 연결을 사용할 수 있다 

![Multiplexing multiple gRPC services in the same server application](README.assets/grpc_0504.png)

#### 동일한 grpc.Server 를 공유하는 두 개의  gRPC 서비스

```go
func main() {
	initSampleData()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer() 1

	// Register Order Management service on gRPC orderMgtServer
	ordermgt_pb.RegisterOrderManagementServer(grpcServer, &orderMgtServer{}) 2

	// Register Greeter Service on gRPC orderMgtServer
	hello_pb.RegisterGreeterServer(grpcServer, &helloServer{}) 3

      ...
}
```

