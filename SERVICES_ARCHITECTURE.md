# Joker Backend - Microservices Architecture

## 개요

Joker Backend는 마이크로서비스 아키텍처로 구성되어 있으며, 각 서비스는 완전히 독립적으로 개발, 배포, 확장이 가능합니다.

## 프로젝트 구조

```
joker_backend/
├── shared/                 # 공유 모듈 (모든 서비스가 사용)
│   ├── go.mod             # shared 모듈 정의
│   ├── config/            # 환경 설정
│   ├── logger/            # 구조화된 로깅
│   ├── db/               # Database 연결 (MySQL, Redis)
│   ├── middleware/       # HTTP 미들웨어
│   ├── errors/           # 에러 핸들링
│   ├── utils/            # 유틸리티 함수
│   ├── aws/              # AWS 서비스 통합
│   └── init.go           # 통합 초기화 함수
│
├── services/
│   ├── authService/      # 인증/인가 서비스 (Port 8080)
│   │   ├── cmd/server/
│   │   ├── features/auth/
│   │   └── go.mod        # shared를 의존성으로 참조
│   │
│   └── gameService/      # 게임 관리 서비스 (Port 8081)
│       ├── cmd/server/
│       ├── features/game/
│       └── go.mod        # shared를 의존성으로 참조
│
├── docker-compose.yml     # 전체 시스템 오케스트레이션
└── scripts/              # 배포 및 유틸리티 스크립트
```

## 핵심 설계 원칙

### 1. 단일 Shared 모듈
- **하나의 shared 폴더**만 존재 (`/shared`)
- 모든 서비스가 **동일한 shared 버전**을 참조
- `go.mod`의 `replace` 지시문으로 로컬 경로 참조

```go
// services/authService/go.mod
module main

require (
    github.com/JokerTrickster/joker_backend/shared v0.0.0
)

replace github.com/JokerTrickster/joker_backend/shared => ../../shared
```

### 2. 서비스 독립성
- 각 서비스는 자체 `go.mod` 보유
- 독립적으로 빌드/배포 가능
- 서비스 간 직접 의존성 없음

### 3. 공유 초기화 패턴
모든 서비스는 `shared.Init()`을 통해 일관된 초기화:

```go
e, err := shared.Init(&shared.InitConfig{
    LogLevel:    "info",
    Environment: "production",
})
defer shared.Cleanup()
```

## 각 서비스별 특징

### 1. Auth Service (Port 8080)
- **역할**: 사용자 인증, 회원가입, 로그인
- **기술 스택**: Go + Echo + MySQL + JWT
- **엔드포인트**: `/api/v1/auth/*`

### 2. Game Service (Port 8081)
- **역할**: 게임 데이터 관리, 게임 로직
- **기술 스택**: Go + Echo + MySQL
- **엔드포인트**: `/api/v1/game/*`

## Shared 모듈 구성

### 포함 컴포넌트
- **config**: 환경 설정 로드 (.env 파일)
- **logger**: 구조화된 로깅 (zap)
- **db**: Database 연결 (MySQL, Redis)
- **middleware**: HTTP 미들웨어
  - CORS
  - Rate Limiting
  - Request ID
  - Recovery
  - Request Logger
  - Timeout
- **errors**: 커스텀 에러 핸들링
- **utils**: 공통 유틸리티 함수
- **aws**: AWS 서비스 통합 (S3, SES, SSM)
- **init.go**: 통합 초기화 함수

### 초기화 순서
1. Logger 초기화
2. Config 로드
3. Database 연결 (MySQL)
4. Echo 서버 생성
5. Middleware 설정
6. Health check 엔드포인트 등록

## 로컬 개발

### 각 서비스 개별 실행

```bash
# Auth Service
cd services/authService
go run ./cmd/server/main.go  # Port 8080

# Game Service (다른 터미널)
cd services/gameService
go run ./cmd/server/main.go  # Port 8081
```

### Shared 모듈 수정 후
```bash
# 각 서비스에서 의존성 업데이트
cd services/authService
go mod tidy

cd services/gameService
go mod tidy
```

## Docker Compose 실행

### 전체 시스템 시작
```bash
# MySQL + Auth Service + Game Service 모두 시작
docker-compose up -d

# 특정 서비스만 시작
docker-compose up -d auth-service
docker-compose up -d game-service

# 로그 확인
docker-compose logs -f auth-service
docker-compose logs -f game-service

# 전체 중지
docker-compose down
```

## 새 서비스 추가 방법

### 1. 서비스 폴더 생성
```bash
mkdir -p services/newService/{cmd/server,features,tests/e2e}
```

### 2. go.mod 생성
```bash
cd services/newService
cat > go.mod <<EOF
module main

go 1.24.0

require (
    github.com/JokerTrickster/joker_backend/shared v0.0.0
    github.com/labstack/echo/v4 v4.13.4
    go.uber.org/zap v1.27.0
)

replace github.com/JokerTrickster/joker_backend/shared => ../../shared
EOF

go mod tidy
```

### 3. main.go 작성
```go
package main

import (
    "github.com/JokerTrickster/joker_backend/shared"
    "main/features/newfeature/handler"
)

func main() {
    e, err := shared.Init(&shared.InitConfig{
        LogLevel: "info",
        Environment: "development",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shared.Cleanup()

    handler.RegisterRoutes(e)
    e.Start(":8082")
}
```

### 4. docker-compose.yml 업데이트
```yaml
  new-service:
    build:
      context: ./services/newService
    ports:
      - "8082:8082"
    environment:
      MYSQL_HOST: mysql
      PORT: 8082
    depends_on:
      - mysql
```

## 환경 변수

각 서비스는 다음 환경 변수를 사용합니다:

### Database
- `MYSQL_HOST`: MySQL 호스트
- `MYSQL_PORT`: MySQL 포트 (기본 3306)
- `MYSQL_USER`: MySQL 사용자
- `MYSQL_PASSWORD`: MySQL 비밀번호
- `MYSQL_DATABASE`: 데이터베이스 이름

### Application
- `ENV`: 환경 (development/production)
- `LOG_LEVEL`: 로그 레벨 (debug/info/warn/error)
- `PORT`: 서비스 포트
- `CORS_ALLOWED_ORIGINS`: CORS 허용 오리진

### Optional
- `IS_LOCAL`: 로컬 모드 (true/false)

## Shared 모듈 버전 관리

### 로컬 개발 (현재)
- `replace` 지시문으로 로컬 경로 참조
- 변경 사항 즉시 반영

### GitHub 버전 관리 (향후)
shared를 GitHub에 태그하여 버전 관리 가능:

```bash
# shared 폴더에서
git tag shared/v1.0.0
git push origin shared/v1.0.0

# 서비스 go.mod에서
require (
    github.com/JokerTrickster/joker_backend/shared v1.0.0
)
# replace 지시문 제거
```

## 모니터링 및 로깅

각 서비스는 구조화된 로그를 출력합니다 (zap):
- Request ID 추적
- 성능 메트릭
- 에러 스택 트레이스
- JSON 형식 로그

## 확장성

- **수평 확장**: Docker Compose 또는 Kubernetes로 인스턴스 추가
- **로드 밸런싱**: Nginx, HAProxy 또는 K8s Ingress
- **데이터베이스**: 읽기 복제본, 샤딩으로 확장
- **캐싱**: Redis를 통한 성능 최적화

## 배포

### 개발 환경
```bash
docker-compose up -d
```

### 프로덕션 환경
- AWS ECS/EKS
- Google Cloud Run
- Azure Container Instances
- 각 서비스 독립 배포 가능

## 장점

1. **단일 Shared 관리**
   - 중복 코드 제거
   - 일관된 동작 보장
   - 한 곳에서 업데이트

2. **서비스 독립성**
   - 각 서비스 독립 배포
   - 장애 격리
   - 팀 간 독립 개발

3. **간단한 의존성 관리**
   - `replace` 지시문으로 로컬 개발 용이
   - 필요시 GitHub 버전 관리로 전환 가능

## 주의사항

1. **모듈 이름**: 모든 서비스는 `module main`을 사용
2. **포트 충돌**: 각 서비스는 고유한 포트 사용
3. **Shared 변경**: shared 수정 후 모든 서비스에서 `go mod tidy` 실행
4. **데이터베이스**: 현재는 공유 MySQL, 향후 서비스별 DB 분리 가능
