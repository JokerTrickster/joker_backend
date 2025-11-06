# Joker Backend

통합 백엔드 서비스 플랫폼 - Go, Echo, MySQL 기반의 클린 아키텍처

## 기술 스택

- **언어**: Go 1.23+
- **프레임워크**: Echo v4
- **데이터베이스**: MySQL 8.0
- **아키텍처**: Clean Architecture
- **컨테이너**: Docker & Docker Compose

## 프로젝트 구조

```
joker_backend/
├── cmd/
│   └── server/          # 애플리케이션 엔트리 포인트
│       └── main.go
├── config/              # 설정 관리
│   └── config.go
├── internal/            # 비즈니스 로직 (Clean Architecture)
│   ├── handler/         # HTTP 핸들러 (Presentation Layer)
│   ├── service/         # 비즈니스 로직 (Use Case Layer)
│   ├── repository/      # 데이터 접근 (Data Layer)
│   ├── model/           # 도메인 모델
│   └── middleware/      # 커스텀 미들웨어
├── pkg/                 # 공용 패키지
│   ├── database/        # 데이터베이스 연결
│   ├── logger/          # 로깅 유틸리티
│   └── response/        # 표준 응답 포맷
├── scripts/             # 데이터베이스 초기화 스크립트
├── docker-compose.yml   # 다중 서비스 오케스트레이션
├── Dockerfile           # 컨테이너 이미지 정의
└── Makefile             # 빌드 자동화

```

## 클린 아키텍처 레이어

```
Handler (Presentation) → Service (Use Case) → Repository (Data) → Database
```

- **Handler**: HTTP 요청/응답 처리, 입력 검증
- **Service**: 비즈니스 로직 구현, 트랜잭션 관리
- **Repository**: 데이터 영속성, SQL 쿼리 실행
- **Model**: 도메인 엔티티 정의

## 빠른 시작

### 사전 요구사항

- Go 1.23 이상
- Docker & Docker Compose
- Make (선택사항)

### 로컬 개발 (Docker Compose)

```bash
# 1. 환경 변수 설정
cp .env.example .env

# 2. Docker Compose로 모든 서비스 시작
make docker-up
# 또는
docker-compose up -d

# 3. 로그 확인
make docker-logs

# 4. 서비스 중지
make docker-down
```

### 로컬 개발 (Go 직접 실행)

```bash
# 1. 의존성 설치
go mod tidy

# 2. MySQL 시작 (Docker)
docker-compose up -d mysql

# 3. 환경 변수 설정
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=joker_user
export DB_PASSWORD=joker_password
export DB_NAME=joker_backend

# 4. 애플리케이션 실행
make run
# 또는
go run ./cmd/server/main.go
```

### 빌드

```bash
# 바이너리 빌드
make build

# 실행
./bin/server
```

## API 엔드포인트

### Health Check

```bash
GET /health
```

### Users API (v1)

```bash
# 사용자 조회
GET /api/v1/users/:id

# 사용자 생성
POST /api/v1/users
Content-Type: application/json

{
  "name": "홍길동",
  "email": "hong@example.com"
}
```

### 응답 형식

**성공 응답**:
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

**에러 응답**:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

## Make 명령어

```bash
make help           # 사용 가능한 명령어 보기
make build          # Go 애플리케이션 빌드
make run            # 로컬에서 애플리케이션 실행
make test           # 테스트 실행
make clean          # 빌드 아티팩트 삭제
make docker-up      # Docker Compose 서비스 시작
make docker-down    # Docker Compose 서비스 중지
make docker-logs    # Docker 로그 확인
make docker-rebuild # Docker 서비스 재빌드
make tidy           # Go 모듈 정리
make fmt            # Go 코드 포맷팅
```

## 개발 가이드

### 새로운 API 추가하기

1. **Model 정의** (`internal/model/`)
2. **Repository 구현** (`internal/repository/`)
3. **Service 로직 작성** (`internal/service/`)
4. **Handler 생성** (`internal/handler/`)
5. **Routes 등록** (`internal/handler/routes.go`)

### 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `DB_HOST` | MySQL 호스트 | localhost |
| `DB_PORT` | MySQL 포트 | 3306 |
| `DB_USER` | MySQL 사용자 | root |
| `DB_PASSWORD` | MySQL 비밀번호 | - |
| `DB_NAME` | 데이터베이스 이름 | joker_backend |
| `PORT` | API 서버 포트 | 8080 |
| `LOG_LEVEL` | 로그 레벨 | info |

## CI/CD

이 프로젝트는 GitHub Actions와 Self-hosted Runner를 사용한 자동 배포를 지원합니다.

### 배포 방법

```bash
# 자동 배포: main 브랜치에 push
git push origin main

# 수동 배포: 배포 스크립트 사용
./scripts/deploy.sh [service-name] [port]

# 예시
./scripts/deploy.sh joker-backend 6000
```

자세한 내용은 [CI/CD 가이드](docs/CICD.md)를 참고하세요.

## 다중 서비스 지원

각 서비스는 독립된 포트에서 실행됩니다:

- 인증 서비스: 포트 6000 (DB: 3309)
- 게임 서버: 포트 6001 (DB: 3310) - 추가 예정
- 결제 서비스: 포트 6002 (DB: 3311) - 추가 예정

새 서비스 추가 방법은 [CI/CD 가이드](docs/CICD.md)를 참고하세요.

## 라이센스

MIT
