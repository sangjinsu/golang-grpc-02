# 보안 적용 gRPC

## TLS 를 사용한 gRPC 채널 인증

TLS Transport Level Security 는 통신하는 두 애플리케이션 간에 정보 보호와 데이터 무결성을 제공을 목표로 하며 gRPC 클라이언트와 서버 애플리케이션 간 안전한 연결을 제공한다 

클라이언트와 서버 간의 연결이 안전한 경우는 둘 중 하나를 만족 시켜야 한다 

### 연결은 private  이다

- 대칭키 암호화를 사용한다 
- 암호화와 복호화에 하나의 비밀키만 사요하는 암호화 방식으로 
- 암호화 키는 세션 시작시에 협의된 공유 암호키를 기반으로 각 연결에 대해  고유하게 생성된다 

### 연결은 reliable 신뢰적이다 

- 각 메시지 전송에 있어 감지되지 않는 데이터 손실이나 변경을 방지하고자 메시지 무결성 검사가 포함되기에 가능하다 



### 단방향 보안 연결 활성화

단방향 연결에서는 클라이언트만 서버 유효성을 검사해 원래 의도된 서버로부터 데이터가 수신됐는지 확인한다 

클라이언트와 서버 간의 연결이 시작되면 서버는 공개 인증서를 클라이언트와 공유한 다음에 클라이언트는 수신한 인증서의 유효성을 확인한다 

인증기관 CA Certificate Agency 의 전자 서명 인증서를 통해 인증서 유효성을 확인한다 

인증서 유효성이 확인되면 클라이언트는 비밀키를 사용해 암호화된 데이터를 보낸다 

TLS 활성화 하려면 인증서와 키가 만들어져야 한다 

1. server.key

   서명하고 공개키를 인증하기 위한 RSA 개인키

2. 배포를 위한 자체 서명 self-signed  ehls X.509 공개키



#### gRPC 서버에서 단방향 보안 연결 사용

- 서버는 공개키/개인키 쌍으로 초기화해야 한다 

```go
package main

import (
  "crypto/tls"
  "errors"
  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
  "log"
  "net"
)

var (
  port = ":50051"
  crtFile = "server.crt"
  keyFile = "server.key"
)

func main() {
  cert, err := tls.LoadX509KeyPair(crtFile,keyFile) 1
  if err != nil {
     log.Fatalf("failed to load key pair: %s", err)
  }
  opts := []grpc.ServerOption{
     grpc.Creds(credentials.NewServerTLSFromCert(&cert)) 2
  }

  s := grpc.NewServer(opts...) 3
  pb.RegisterProductInfoServer(s, &server{}) 4

  lis, err := net.Listen("tcp", port) 5
  if err != nil {
     log.Fatalf("failed to listen: %v", err)
  }

  if err := s.Serve(lis); err != nil { 6
     log.Fatalf("failed to serve: %v", err)
  }
}
```



####  gRPC 클라이언트에서 단방향 보안 연결 사용

```go
package main

import (
  "log"

  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc/credentials"
  "google.golang.org/grpc"
)

var (
  address = "localhost:50051"
  hostname = "localhost
  crtFile = "server.crt"
)

func main() {
  creds, err := credentials.NewClientTLSFromFile(crtFile, hostname) 1
  if err != nil {
     log.Fatalf("failed to load credentials: %v", err)
  }
  opts := []grpc.DialOption{
     grpc.WithTransportCredentials(creds), 2
  }

  conn, err := grpc.Dial(address, opts...) 3
  if err != nil {
     log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close() 5
  c := pb.NewProductInfoClient(conn) 4

  .... // Skip RPC method invocation.
}
```



### mTLS 보안 연결 활성화

클라이언트와 서버 간 mTLS 연결의 주요 목적은 서버에 연결하는 클라이언트를 제어하는 것이다 

단방향 TLS 연결과 달리 서버는 검증된 클라이언트의 제한된 그룹에서의 연결만을 수락하도록 구성한다 

1. 클라이언트는 서버의 보호된 정보에 액세스하기 위한 요청을 보낸다 
2. 서버는 X.509 인증서를 클라이언트로 보낸다 
3. 클라이언트는 CA 서명 인증서에 대해 CA를 통해 수신된 인증서의 유효성을 검사한다 
4. 확인에 성공하면 클라이언트는 인증서를 서버로 보낸다 
5. 서버도 CA를 통해 클라이언트 인증서를 검증한다 
6. 성공하면 서버는 보호된 데이터에 접근할 수 있는 권한을 부여한다 

#### gRPC 서버에서 mTLS 활성화

```go
package main

import (
  "crypto/tls"
  "crypto/x509"
  "errors"
  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
  "io/ioutil"
  "log"
  "net"
)

var (
  port = ":50051"
  crtFile = "server.crt"
  keyFile = "server.key"
  caFile = "ca.crt"
)

func main() {
  certificate, err := tls.LoadX509KeyPair(crtFile, keyFile) 1
  if err != nil {
     log.Fatalf("failed to load key pair: %s", err)
  }

  certPool := x509.NewCertPool() 2
  ca, err := ioutil.ReadFile(caFile)
  if err != nil {
     log.Fatalf("could not read ca certificate: %s", err)
  }

  if ok := certPool.AppendCertsFromPEM(ca); !ok { 3
     log.Fatalf("failed to append ca certificate")
  }

  opts := []grpc.ServerOption{
     // Enable TLS for all incoming connections.
     grpc.Creds( 4
        credentials.NewTLS(&tls.Config {
           ClientAuth:   tls.RequireAndVerifyClientCert,
           Certificates: []tls.Certificate{certificate},
           ClientCAs:    certPool,
           },
        )),
  }

  s := grpc.NewServer(opts...) 5
  pb.RegisterProductInfoServer(s, &server{}) 6

  lis, err := net.Listen("tcp", port) 7
  if err != nil {
     log.Fatalf("failed to listen: %v", err)
  }

  if err := s.Serve(lis); err != nil { 8
     log.Fatalf("failed to serve: %v", err)
  }
}
```



#### gRPC 클라이언트에서 mTLS 활성화

```go
package main

import (
  "crypto/tls"
  "crypto/x509"
  "io/ioutil"
  "log"

  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
)



var (
  address = "localhost:50051"
  hostname = "localhost"
  crtFile = "client.crt"
  keyFile = "client.key"
  caFile = "ca.crt"
)

func main() {
  certificate, err := tls.LoadX509KeyPair(crtFile, keyFile) 1
  if err != nil {
     log.Fatalf("could not load client key pair: %s", err)
  }

  certPool := x509.NewCertPool() 2
  ca, err := ioutil.ReadFile(caFile)
  if err != nil {
     log.Fatalf("could not read ca certificate: %s", err)
  }

  if ok := certPool.AppendCertsFromPEM(ca); !ok { 3
     log.Fatalf("failed to append ca certs")
  }

  opts := []grpc.DialOption{
     grpc.WithTransportCredentials( credentials.NewTLS(&tls.Config{ 4
        ServerName:   hostname, // NOTE: this is required!
        Certificates: []tls.Certificate{certificate},
        RootCAs:      certPool,
     })),
  }

  conn, err := grpc.Dial(address, opts...) 5
  if err != nil {
     log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()7
  c := pb.NewProductInfoClient(conn) 6

  .... // Skip RPC method invocation.
}
```



### gRPC 호출 인증

gRPC 는 견고한 인증 메커니즘을 사용하도록 설계되었음

gRPC 서버는 클라이언트 요청을 가로채서 모든 수신 통신에서 자격증명을 확인할 수 있다 

#### 베이직 인증 사용

Basic Authentication 은 가장 간단한 인증 메커니즘이다 

- 커스텀 자격증명 전달을 위한 PerRPCCredentials  인터페이스 구현

```go
type basicAuth struct { 1
  username string
  password string
}

func (b basicAuth) GetRequestMetadata(ctx context.Context,
  in ...string)  (map[string]string, error) { 2
  auth := b.username + ":" + b.password
  enc := base64.StdEncoding.EncodeToString([]byte(auth))
  return map[string]string{
     "authorization": "Basic " + enc,
  }, nil
}

func (b basicAuth) RequireTransportSecurity() bool { 3
  return true
}
```

- 클라이언트

```go
package main

import (
  "log"
  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc/credentials"
  "google.golang.org/grpc"
)

var (
  address = "localhost:50051"
  hostname = "localhost"
  crtFile = "server.crt"
)

func main() {
  creds, err := credentials.NewClientTLSFromFile(crtFile, hostname)
  if err != nil {
     log.Fatalf("failed to load credentials: %v", err)
  }

  auth := basicAuth{ 1
    username: "admin",
    password: "admin",
  }

  opts := []grpc.DialOption{
     grpc.WithPerRPCCredentials(auth), 2
     grpc.WithTransportCredentials(creds),
  }

  conn, err := grpc.Dial(address, opts...)
  if err != nil {
     log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()
  c := pb.NewProductInfoClient(conn)

  .... // Skip RPC method invocation.
}
```

- 서버

```go
package main

import (
  "context"
  "crypto/tls"
  "encoding/base64"
  "errors"
  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/credentials"
  "google.golang.org/grpc/metadata"
  "google.golang.org/grpc/status"
  "log"
  "net"
  "path/filepath"
  "strings"
)


var (
  port = ":50051"
  crtFile = "server.crt"
  keyFile = "server.key"
  errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
  errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid credentials")
)

type server struct {
  productMap map[string]*pb.Product
}


func main() {
  cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
  if err != nil {
     log.Fatalf("failed to load key pair: %s", err)
  }
  opts := []grpc.ServerOption{
     // Enable TLS for all incoming connections.
     grpc.Creds(credentials.NewServerTLSFromCert(&cert)),

     grpc.UnaryInterceptor(ensureValidBasicCredentials), 1
  }

  s := grpc.NewServer(opts...)
  pb.RegisterProductInfoServer(s, &server{})

  lis, err := net.Listen("tcp", port)
  if err != nil {
     log.Fatalf("failed to listen: %v", err)
  }

  if err := s.Serve(lis); err != nil {
     log.Fatalf("failed to serve: %v", err)
  }
}

func valid(authorization []string) bool {
  if len(authorization) < 1 {
     return false
  }
  token := strings.TrimPrefix(authorization[0], "Basic ")
  return token == base64.StdEncoding.EncodeToString([]byte("admin:admin"))
}

func ensureValidBasicCredentials(ctx context.Context, req interface{}, info
*grpc.UnaryServerInfo,
     handler grpc.UnaryHandler) (interface{}, error) { 2
  md, ok := metadata.FromIncomingContext(ctx) 3
  if !ok {
     return nil, errMissingMetadata
  }
  if !valid(md["authorization"]) {
     return nil, errInvalidToken
  }
  // Continue execution of handler after ensuring a valid token.
  return handler(ctx, req)
}
```



### OAuth 2.0 사용

- 클라이언트

```go
package main

import (
  "google.golang.org/grpc/credentials"
  "google.golang.org/grpc/credentials/oauth"
  "log"

  pb "productinfo/server/ecommerce"
  "golang.org/x/oauth2"
  "google.golang.org/grpc"
)

var (
  address = "localhost:50051"
  hostname = "localhost"
  crtFile = "server.crt"
)

func main() {
  auth := oauth.NewOauthAccess(fetchToken()) 1

  creds, err := credentials.NewClientTLSFromFile(crtFile, hostname)
  if err != nil {
     log.Fatalf("failed to load credentials: %v", err)
  }

  opts := []grpc.DialOption{
     grpc.WithPerRPCCredentials(auth), 2
     grpc.WithTransportCredentials(creds),
  }

  conn, err := grpc.Dial(address, opts...)
  if err != nil {
     log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()
  c := pb.NewProductInfoClient(conn)

  .... // Skip RPC method invocation.
}

func fetchToken() *oauth2.Token {
  return &oauth2.Token{
     AccessToken: "some-secret-token",
  }
}
```

- Server

```go
package main

import (
  "context"
  "crypto/tls"
  "errors"
  "log"
  "net"
  "strings"

  pb "productinfo/server/ecommerce"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/credentials"
  "google.golang.org/grpc/metadata"
  "google.golang.org/grpc/status"
)

// server is used to implement ecommerce/product_info.
type server struct {
  productMap map[string]*pb.Product
}

var (
  port = ":50051"
  crtFile = "server.crt"
  keyFile = "server.key"
  errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
  errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

func main() {
  cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
  if err != nil {
     log.Fatalf("failed to load key pair: %s", err)
  }
  opts := []grpc.ServerOption{
     grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
     grpc.UnaryInterceptor(ensureValidToken), 1
  }

  s := grpc.NewServer(opts...)
  pb.RegisterProductInfoServer(s, &server{})

  lis, err := net.Listen("tcp", port)
  if err != nil {
     log.Fatalf("failed to listen: %v", err)
  }

  if err := s.Serve(lis); err != nil {
     log.Fatalf("failed to serve: %v", err)
  }
}

func valid(authorization []string) bool {
  if len(authorization) < 1 {
     return false
  }
  token := strings.TrimPrefix(authorization[0], "Bearer ")
  return token == "some-secret-token"
}

func ensureValidToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
     handler grpc.UnaryHandler) (interface{}, error) { 2
  md, ok := metadata.FromIncomingContext(ctx)
  if !ok {
     return nil, errMissingMetadata
  }
  if !valid(md["authorization"]) {
     return nil, errInvalidToken
  }
  return handler(ctx, req)
}
```



### JWT

- client

```go
jwtCreds, err := oauth.NewJWTAccessFromFile(“token.json”) 1
if err != nil {
  log.Fatalf("Failed to create JWT credentials: %v", err)
}

creds, err := credentials.NewClientTLSFromFile("server.crt",
     "localhost")
if err != nil {
    log.Fatalf("failed to load credentials: %v", err)
}
opts := []grpc.DialOption{
  grpc.WithPerRPCCredentials(jwtCreds),
  // transport credentials.
  grpc.WithTransportCredentials(creds), 2
}

// Set up a connection to the server.
conn, err := grpc.Dial(address, opts...)
if err != nil {
  log.Fatalf("did not connect: %v", err)
}
  .... // Skip Stub generation and RPC method invocation.
```

