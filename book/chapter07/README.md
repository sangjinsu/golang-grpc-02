# 서비스 수준 gRPC 실행

- 단위 테스트
- 통합 테스트
- 지속적인 통합 CI

## gRPC 애플리케이션 테스트

### Go 를 사용한 gRPC 서버 측 테스트

```go
func TestServer_AddProduct(t *testing.T) { 1
	grpcServer := initGRPCServerHTTP2() 2
	conn, err := grpc.Dial(address, grpc.WithInsecure()) 3
	if err != nil {

           grpcServer.Stop()
           t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewProductInfoClient(conn)

	name := "Sumsung S10"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in
	February 2019"
	price := float32(700.0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.AddProduct(ctx, &pb.Product{Name: name,
	                                Description: description, Price: price}) 4
	if err != nil { 5
		t.Fatalf("Could not add product: %v", err)
	}

	if r.Value == "" {
		t.Errorf("Invalid Product ID %s", r.Value)
	}
	log.Printf("Res %s", r.Value)
      grpcServer.Stop()
}
```

### gRPC 클라이언트 테스트

- gRPC 서버를 시작시키고 mock 서비스 구현한다

[gomock github](https://github.com/golang/mock#running-mockgen)

#### gomock을 활용한 gRPC 클라이언트 측 테스트

```go
func TestAddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocklProdInfoClient := NewMockProductInfoClient(ctrl) 1
     ...
	req := &pb.Product{Name: name, Description: description, Price: price}

	mocklProdInfoClient. 2
	 EXPECT().AddProduct(gomock.Any(), &rpcMsg{msg: req},). 3
	 Return(&wrapper.StringValue{Value: "ABC123" + name}, nil) 4

	testAddProduct(t, mocklProdInfoClient) 5
}

func testAddProduct(t *testing.T, client pb.ProductInfoClient) {
	ctx, cancel :  = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	...

	r, err := client.AddProduct(ctx, &pb.Product{Name: name,
    Description: description, Price: price})

	// test and verify response.
}
```

gRPC 서버를 모킹해도 실제 gRPC 서버와 동일한 동작을 제공하지 않는다 

gRPC 서버에 존재하는 모든 에러 로직을 다시 구현하지 않으면 특정 기능은 테스트를 통해 검증되지 못할 수 있다 

모킹을 통해 일부 선택적인 지능 집합을 검증하고 나머지는 실제 gRPC 서버 구현을 통해 검증해야 한다 



### 부하 테스트

기존 도구를 사용해 gRPC 애플리케이션의 부하 테스트와 벤치 마킹을 수행하기는 쉽지 않다 

gRPC의 경우 서버에 대한 RPC 가상 부하를 생성해 gRPC 서버를 부하 테스트할 수 있는 맞춤형 부하 테스트 도구가 필요하다 

https://ghz.sh/

https://github.com/bojand/ghz

ghz 는 맞춤형 부하를 지원하는 부하 테스트 도구이다 

Go 언어를 사용해 커맨드라인 유틸리티로 구현된다 

로컬 서비스를 테스트하거나 디버깅할 수도 있지만 성능에 대한 회귀 테스트를 위한 자동화된 CI 환경에서 사용할 수도 있다 

```bash
ghz --insecure \
	--proto ./greeter.proto \
	--call helloworld.Greeter.SayHello \
	-d '{"name":"Joe"}' \
	-n 2000
	-c 20 \
	0.0.0.0:50051
```

총 요청수 2000 과 동시성 스레드 20개를 지정한다 



### 지속적인 통합

#### 도커

1. 애플리케이션 빌드
2. 훨씬 가벼운 런타임으로 실행하는 멀티단계 multistage 도커 빌드를 사용한다
3. 생성된 서버 측 코든느 애플리케이션을 빌드하기 전에 컨테이너에 먼저 추가된다 

```dockerfile
# Multistage build

# Build stage I: 1
FROM golang AS build
ENV location /go/src/github.com/grpc-up-and-running/samples/ch07/grpc-docker/go
WORKDIR ${location}/server

ADD ./server ${location}/server
ADD ./proto-gen ${location}/proto-gen

RUN go get -d ./... 2
RUN go install ./... 3

RUN CGO_ENABLED=0 go build -o /bin/grpc-productinfo-server 4

# Build stage II: 5
FROM scratch
COPY --from=build /bin/grpc-productinfo-server /bin/grpc-productinfo-server 6

ENTRYPOINT ["/bin/grpc-productinfo-server"]
EXPOSE 50051
```

```bash
# 이미지 빌드
docker image build -t grpc-productinfo-server -f server/Dockerfile

# 실행
docker run -it --network=my-net --name=productinfo \
    --hostname=productinfo
    -p 50051:50051  grpc-productinfo-server 

docker run -it --network=my-net \
    --hostname=client grpc-productinfo-client 
```



#### 쿠버네티스

쿠버네티스 플랫폼은 컨테이너를 직접 관리하지 않고  pod 라는 추상화를 사용한다 

이 pod 는 하나 이상의 컨테이너를 포함할 수 있는 논리적 단위이자 쿠버네티스의 복제 단위이다 

```yaml
apiVersion: apps/v1
kind: Deployment 
metadata:
  name: grpc-productinfo-server 
spec:
  replicas: 1 
  selector:
    matchLabels:
      app: grpc-productinfo-server
  template:
    metadata:
      labels:
        app: grpc-productinfo-server
    spec:
      containers:
      - name: grpc-productinfo-server 
        image: kasunindrasiri/grpc-productinfo-server 
        resources:
          limits:
            memory: "128Mi"
            cpu: "500m"
        ports:
        - containerPort: 50051
          name: grpc
```



#### gRPC 서버를 위한 쿠버네티스 서비스 리소스

쿠버네티스 서비스를 생성하고 이를 매칭되는 파드와 연결하면 트래픽을 해당 파드로 자동 라우팅하는 DNS 이름을 얻는다 

##### Go gRPC 서버 애플리케이션의 쿠버네티스 서비스 서술자

```yaml
apiVersion: v1
kind: Service 
metadata:
  name: productinfo 
spec:
  selector:
    app: grpc-productinfo-server 
  ports:
  - port: 50051 
    targetPort: 50051
    name: grpc
  type: NodePort
```



## 관찰 가능성

시스템 내부 상태가 외부 출력의 정보로부터 얼마나 잘 추론될 수 있는지를 나타내는 척도 

### 메트릭

- 일정 간격 동안 측정된 데이터 숫자 표현
- 시스템 수준 메트릭
  - CPU 사용량
  - 메모리 사용량 

- 애플리케이션 수준 매트릭
  - 인바운드 요청률 
  - 요청 에러율

#### 오픈 센서스 gRPC 활용

클라이언트와 서버 애플리케이션 모두에 핸들러를 추가해 휩게 활성화 한다 

#### 프로메테우스 gRPC 활용

프로메테우스는 시스템 모니터링과 알림을 위한 오픈소스 툴킷이다 



## 로그

시간이 지남에 따라 발생하는 불변의 타임스탬프된 개별 이벤트 기록이다 

gRPC 애플리케이션에서는 인터셉터를 사용해 로깅을 활성화 할 수 있다 



## 디버깅과 문제 해결

이슈를 디버깅하고 해결하려면 테스트 환경에서 동일한 문제를 재현해야 한다 

포스트맨 같은 일반 도구로는 gRPC 서비스를 테스트할 수 없다

gRPC 애플리케이션을 디버깅하는 가장 일반적인 방법 중 하나는 추가 로깅을 사용하는 것이다 

### 추가 로깅 활성화 

```
GRPC_GO_LOG_VERBOSITY_LEVEL=99 
GRPC_GO_LOG_SEVERITY_LEVEL=info 
```

