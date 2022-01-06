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

