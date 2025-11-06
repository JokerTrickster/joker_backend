# Joker Backend - Multi-Service Platform

통합 백엔드 서비스 플랫폼 - Go, Echo, MySQL 기반의 마이크로서비스 아키텍처

## 기술 스택

- **언어**: Go 1.23+
- **프레임워크**: Echo v4
- **데이터베이스**: MySQL 8.0 (공유)
- **아키텍처**: Clean Architecture + Microservices
- **컨테이너**: Docker & Docker Compose
- **CI/CD**: GitHub Actions (경로 기반 자동 배포)

## 프로젝트 구조

```
joker_backend/
├── services/                # 마이크로서비스들
│   ├── auth-service/        # 인증 서비스 (포트 6000)
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── pkg/
│   │   ├── config/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── game-service/        # 게임 서비스 (포트 6001) [예정]
│   └── payment-service/     # 결제 서비스 (포트 6002) [예정]
│
├── shared/                  # 공통 코드
│   ├── models/             # 공통 모델
│   ├── utils/              # 유틸리티
│   └── middleware/         # 공통 미들웨어
│
├── scripts/                # 배포 스크립트
│   ├── deploy-service.sh   # 통합 배포 스크립트
│   ├── cleanup.sh
│   └── init.sql
│
├── .github/
│   └── workflows/
│       └── deploy.yml      # 경로 기반 자동 배포
│
├── docker-compose.yml      # 로컬 개발용
├── docker-compose.prod.yml # 프로덕션 템플릿
└── README.md
```

## 서비스 포트 구조

| 서비스 | 포트 | 상태 | 설명 |
|--------|------|------|------|
| Auth Service | 6000 | ✅ 운영중 | 사용자 인증 및 권한 관리 |
| Game Service | 6001 | 📋 예정 | 게임 로직 및 매칭 |
| Payment Service | 6002 | 📋 예정 | 결제 처리 |

**공통 리소스:**
- MySQL: 포트 3306 (모든 서비스 공유)
- Database: `backend_dev` (모든 서비스 공유)

## 빠른 시작

### 사전 요구사항

- Go 1.23 이상
- Docker & Docker Compose
- Make (선택사항)

### 로컬 개발

```bash
# 1. 저장소 클론
git clone https://github.com/JokerTrickster/joker_backend.git
cd joker_backend

# 2. 환경 변수 설정
cp .env.example .env

# 3. 모든 서비스 시작 (Docker Compose)
docker-compose up -d

# 4. 로그 확인
docker-compose logs -f auth-service

# 5. 서비스 중지
docker-compose down
```

### 개별 서비스 개발

```bash
# Auth Service 개발
cd services/auth-service
go mod tidy
go run ./cmd/server/main.go

# 환경 변수 설정 필요
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=joker_user
export DB_PASSWORD=joker_password
export DB_NAME=backend_dev
export PORT=6000
```

## CI/CD - 자동 배포

### 경로 기반 배포

변경된 서비스만 자동으로 배포됩니다:

```bash
# Auth Service 수정 후 push
git add services/auth-service/
git commit -m "Update auth service"
git push origin main
# → Auth Service만 자동 배포 (포트 6000)

# Shared 코드 수정 후 push
git add shared/
git commit -m "Update shared utilities"
git push origin main
# → 모든 서비스 자동 재배포
```

### 수동 배포

GitHub Actions에서 수동으로 특정 서비스 배포:

1. GitHub Repository → Actions 탭
2. "Deploy Services" 워크플로우 선택
3. "Run workflow" 클릭
4. 배포할 서비스 선택 (auth-service, game-service, payment-service, all)

### 배포 스크립트 직접 사용

```bash
# 서버에서 직접 배포
./scripts/deploy-service.sh auth-service 6000
./scripts/deploy-service.sh game-service 6001
./scripts/deploy-service.sh payment-service 6002
```

## API 엔드포인트

### Auth Service (포트 6000)

```bash
# Health Check
GET http://localhost:6000/health

# 사용자 조회
GET http://localhost:6000/api/v1/users/:id

# 사용자 생성
POST http://localhost:6000/api/v1/users
Content-Type: application/json

{
  "name": "홍길동",
  "email": "hong@example.com"
}
```

### 응답 형식

**성공 응답:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

**에러 응답:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

## 새 서비스 추가하기

### 1. 서비스 디렉토리 생성

```bash
mkdir -p services/your-service
cd services/your-service
```

### 2. Go 모듈 초기화

```bash
go mod init joker_backend/services/your-service
```

### 3. 서비스 코드 작성

Auth Service 구조를 참고하여 작성:
- `cmd/server/main.go` - 엔트리 포인트
- `internal/` - 비즈니스 로직
- `config/` - 설정 관리
- `Dockerfile` - 컨테이너 이미지

### 4. docker-compose.yml에 추가

```yaml
your-service:
  build:
    context: ./services/your-service
    dockerfile: Dockerfile
  container_name: joker_your_api
  environment:
    DB_HOST: mysql
    DB_NAME: backend_dev
    PORT: 6003  # 새 포트 할당
  ports:
    - "6003:6003"
  depends_on:
    - mysql
  networks:
    - joker_network
```

### 5. GitHub Actions 업데이트

`.github/workflows/deploy.yml`에 새 서비스 job 추가

### 6. 배포 테스트

```bash
# 로컬 테스트
docker-compose up -d your-service

# 프로덕션 배포
git add services/your-service/
git commit -m "Add your-service"
git push origin main
```

## 개발 가이드

### Clean Architecture 레이어

```
Handler (Presentation) → Service (Use Case) → Repository (Data) → Database
```

- **Handler**: HTTP 요청/응답 처리, 입력 검증
- **Service**: 비즈니스 로직 구현, 트랜잭션 관리
- **Repository**: 데이터 영속성, SQL 쿼리 실행
- **Model**: 도메인 엔티티 정의

### 공통 코드 사용

```go
// shared 패키지 import
import (
    "joker_backend/shared/models"
    "joker_backend/shared/utils"
)

// 사용 예시
type User struct {
    models.BaseModel
    Name  string `json:"name"`
    Email string `json:"email"`
}

dbHost := utils.GetEnv("DB_HOST", "localhost")
```

## 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `DB_HOST` | MySQL 호스트 | localhost |
| `DB_PORT` | MySQL 포트 | 3306 |
| `DB_USER` | MySQL 사용자 | joker_user |
| `DB_PASSWORD` | MySQL 비밀번호 | - |
| `DB_NAME` | 데이터베이스 이름 | backend_dev |
| `PORT` | API 서버 포트 | 6000 (서비스별 다름) |
| `LOG_LEVEL` | 로그 레벨 | info |

## 모니터링

```bash
# 전체 서비스 상태
docker ps --filter "name=joker"

# 특정 서비스 로그
docker logs -f auth-service_api

# 헬스체크
curl http://localhost:6000/health  # Auth Service
curl http://localhost:6001/health  # Game Service
curl http://localhost:6002/health  # Payment Service
```

## 트러블슈팅

상세한 트러블슈팅 가이드는 [CI/CD 문서](docs/CICD.md)를 참고하세요.

## 라이센스

MIT
