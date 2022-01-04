# gRPC 통신 패턴

## 단순 RPC (단일 RPC, unary)

- 클라이언트가 서버의 원격 기능을 호출하고자 단일 요청을 서버로 보내고 상태에 대한 세부 정보 및 후행 메타데이터와 함께 단일 응답을 받는다.

## 서버 스트리밍 RPC

- 서버가 클라이언트 요청 메시지를 받은 후 일련의 응답을 다시 보낸다
- 모든 서버 응답을 보낸후에 서버는 서버의 상태 정보를 후행 메타데이터로 클라이언트에 전송해 스트림의 종료를 알린다 

## 클라이언트 스트리밍 RPC

- 클라이언트가 여러 메시지를 서버로 보내고 서버는 클라이언트에게 단일 응답을 보낸다.
- 필요한 로직에 따라 스트림에서 하나 또는 여러 메시지를 읽거나 모든 메시지를 읽은 후 응답을 보낼 수 있다 

## 양방향 스트리밍 RPC

- 클라이언트는 메시지 스트림으로 서버에 요청을 보내고, 서버는 메시지 스트림으로도 응답한다
- 호출은 클라이언트가 시작하지만 통신은 gRPC 클라이언트와 서버의 애플리케이션 로직에 따라 완전히 다르다


