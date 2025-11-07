# Game Service

게임 관련 API를 제공하는 마이크로서비스입니다.

## 구조

```
gameService/
├── cmd/
│   └── server/
│       └── main.go          # 서버 진입점
├── features/
│   └── game/
│       ├── handler/         # HTTP 핸들러
│       ├── usecase/         # 비즈니스 로직
│       └── repository/      # 데이터 접근
├── shared/                  # 공유 모듈 (logger, config, db, middleware 등)
├── tests/
│   └── e2e/                # E2E 테스트
├── go.mod
├── go.sum
└── Dockerfile
```

## 로컬 개발

### 1. 환경 변수 설정

`.env` 파일을 생성하세요:

```bash
# Database
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=joker_user
MYSQL_PASSWORD=jokerpassword
MYSQL_DATABASE=joker_db

# App
ENV=development
LOG_LEVEL=debug
PORT=8081

# CORS
CORS_ALLOWED_ORIGINS=*

# Local mode
IS_LOCAL=true
```

### 2. 빌드 및 실행

```bash
# 빌드
go build -o bin/server ./cmd/server

# 실행
./bin/server

# 또는 직접 실행
go run ./cmd/server/main.go
```

## Docker로 실행

### 단일 서비스 실행

```bash
# 이미지 빌드
docker build -t game-service .

# 컨테이너 실행
docker run -p 8081:8081 \
  -e MYSQL_HOST=host.docker.internal \
  -e MYSQL_PORT=3306 \
  -e MYSQL_USER=joker_user \
  -e MYSQL_PASSWORD=jokerpassword \
  -e MYSQL_DATABASE=joker_db \
  -e IS_LOCAL=true \
  game-service
```

### Docker Compose로 전체 시스템 실행

프로젝트 루트에서:

```bash
# 전체 서비스 시작 (MySQL + Auth Service + Game Service)
docker-compose up -d

# 로그 확인
docker-compose logs -f game-service

# 중지
docker-compose down
```

## API 엔드포인트

### Health Check
- `GET /health` - 서비스 상태 확인

### Game API (v1)
- `GET /api/v1/game/list` - 게임 목록 조회
- `GET /api/v1/game/:id` - 특정 게임 조회
- `POST /api/v1/game` - 게임 생성
- `PUT /api/v1/game/:id` - 게임 수정
- `DELETE /api/v1/game/:id` - 게임 삭제

## 테스트

```bash
# 유닛 테스트
go test ./...

# E2E 테스트
go test ./tests/e2e/...

# 커버리지
go test -cover ./...
```

## 포트

- **로컬 개발**: 8081
- **Docker**: 8081

## 의존성

- Go 1.24+
- MySQL 8.0+
- Echo v4 (웹 프레임워크)
- GORM (ORM)
