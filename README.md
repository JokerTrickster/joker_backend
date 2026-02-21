# 🎭 Joker Backend - Cloud Storage Microservices Platform

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/architecture-microservices-orange.svg)](https://microservices.io/)

## 📋 Overview

Joker Backend는 Go 기반의 마이크로서비스 아키텍처로 구축된 클라우드 스토리지 플랫폼입니다. 사용자 인증, 파일 관리, 활동 추적 등의 기능을 제공하며, 높은 확장성과 유지보수성을 목표로 설계되었습니다.

## ✨ Key Features

- **🔐 JWT 기반 인증 시스템** - Access/Refresh 토큰을 활용한 보안 인증
- **📁 파일 관리** - AWS S3를 활용한 안정적인 파일 저장소
- **🖼️ 썸네일 지원** - 이미지 파일의 효율적인 렌더링을 위한 썸네일 처리
- **📊 사용자 통계** - 실시간 스토리지 사용량 및 활동 추적
- **📅 활동 기록** - 일별 업로드/다운로드 활동 내역 관리
- **🏷️ 태그 시스템** - 파일 분류 및 검색 최적화
- **⚡ Rate Limiting** - API 남용 방지 및 서버 보호
- **🔄 CORS 설정** - 크로스 오리진 요청 처리

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.24
- **Framework**: Echo v4 (High performance web framework)
- **ORM**: GORM (Object-Relational Mapping)
- **Database**: MySQL 8.0
- **Authentication**: JWT (JSON Web Tokens)

### Cloud & Infrastructure
- **Storage**: AWS S3
- **AWS SDK**: aws-sdk-go-v2
- **Environment**: Docker support for containerization

### Architecture Patterns
- **Clean Architecture** - 계층별 관심사 분리
- **Repository Pattern** - 데이터 액세스 추상화
- **Use Case Pattern** - 비즈니스 로직 캡슐화
- **Dependency Injection** - 느슨한 결합과 테스트 용이성

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     API Gateway                          │
└──────────┬──────────────────────┬───────────────────────┘
           │                      │
    ┌──────▼──────┐        ┌─────▼──────┐
    │ Auth Service │        │ Cloud Repo │
    │   (18081)    │        │  Service   │
    │              │        │  (18080)   │
    └──────┬──────┘        └─────┬──────┘
           │                      │
    ┌──────▼──────────────────────▼──────┐
    │          MySQL Database            │
    │            (3307)                  │
    └────────────────────────────────────┘
           │                      │
    ┌──────▼──────┐        ┌─────▼──────┐
    │  JWT Tokens │        │   AWS S3    │
    └─────────────┘        └─────────────┘
```

## 📦 Project Structure

```
joker_backend/
├── services/
│   ├── authService/                # 인증 서비스
│   │   ├── cmd/                   # 애플리케이션 진입점
│   │   ├── features/              # 기능별 모듈
│   │   │   └── auth/
│   │   │       ├── handler/       # HTTP 핸들러
│   │   │       ├── usecase/       # 비즈니스 로직
│   │   │       ├── repository/    # 데이터 액세스
│   │   │       └── model/         # 데이터 모델
│   │   └── .env.example           # 환경변수 예제
│   │
│   ├── cloudRepositoryService/     # 클라우드 저장소 서비스
│   │   ├── cmd/
│   │   └── features/
│   │       └── cloudRepository/
│   │
│   └── lottoDefenseService/        # 로또 디펜스 게임 서비스
│       ├── cmd/
│       ├── features/
│       │   └── lottoDefense/
│       │       ├── handler/
│       │       ├── usecase/
│       │       ├── repository/
│       │       └── model/
│       └── pkg/lotto/
│
├── shared/                         # 공통 모듈
│   ├── database/                  # DB 연결 관리
│   ├── errors/                    # 에러 처리
│   ├── jwt/                       # JWT 유틸리티
│   ├── middleware/                # 공통 미들웨어
│   └── utils/                     # 유틸리티 함수
│
└── README.md                       # 프로젝트 문서
```

## 🚀 Getting Started

### Prerequisites

- Go 1.24 이상
- MySQL 8.0
- AWS 계정 (S3 사용)
- Docker (선택사항)

### Installation

1. **Repository Clone**
```bash
git clone https://github.com/JokerTrickster/joker_backend.git
cd joker_backend
```

2. **Dependencies 설치**
```bash
go mod download
```

3. **환경 변수 설정**

각 서비스 디렉토리에 `.env` 파일 생성:

**services/authService/.env**
```env
IS_LOCAL=true
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3307
MYSQL_USER=root
MYSQL_PASSWORD=rootpassword
MYSQL_DATABASE=joker_db
JWT_SECRET=your-secret-key
SERVER_PORT=18081
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

**services/cloudRepositoryService/.env**
```env
IS_LOCAL=true
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3307
MYSQL_USER=root
MYSQL_PASSWORD=rootpassword
MYSQL_DATABASE=joker_db
JWT_SECRET=your-secret-key
SERVER_PORT=18080
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
AWS_REGION=ap-south-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
S3_BUCKET_NAME=joker-cloud-repository-dev
```

4. **Database Setup**
```bash
# MySQL 시작 (Docker 사용 시)
docker run --name joker_mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=rootpassword -e MYSQL_DATABASE=joker_db -d mysql:8.0

# 테이블은 서비스 시작 시 자동 마이그레이션됨
```

5. **서비스 실행**

각 서비스를 별도 터미널에서 실행:

```bash
# Auth Service 실행
cd services/authService
go run cmd/main.go

# Cloud Repository Service 실행
cd services/cloudRepositoryService
go run cmd/main.go

# Lotto Defense Service 실행
cd services/lottoDefenseService
go run cmd/main.go
```

## 📚 API Documentation

### Authentication Service (Port: 18081)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v0.1/auth/signin` | 사용자 로그인 |
| POST | `/v0.1/auth/signup` | 사용자 회원가입 |
| POST | `/v0.1/auth/refresh` | 토큰 갱신 |
| POST | `/v0.1/auth/signout` | 로그아웃 |

### Cloud Repository Service (Port: 18080)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/files/upload` | 파일 업로드 URL 요청 | ✅ |
| POST | `/api/v1/files/download` | 파일 다운로드 URL 요청 | ✅ |
| GET | `/api/v1/files` | 파일 목록 조회 | ✅ |
| DELETE | `/api/v1/files/{id}` | 파일 삭제 | ✅ |
| GET | `/api/v1/user/stats` | 사용자 통계 조회 | ✅ |
| GET | `/api/v1/user/activity` | 활동 내역 조회 | ✅ |
| POST | `/api/v1/tags` | 태그 생성 | ✅ |
| POST | `/api/v1/files/{id}/tags` | 파일에 태그 추가 | ✅ |

### Lotto Defense Service (Port: 18082)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/game/rounds` | 라운드 시작 | ✅ |
| PATCH | `/api/v1/game/rounds/:id/end` | 라운드 종료 (점수 제출, 로또 번호 생성) | ✅ |
| GET | `/api/v1/game/rounds` | 내 라운드 목록 | ✅ |
| GET | `/api/v1/game/rounds/:id` | 라운드 상세 (완료 시 로또 번호 포함) | ✅ |
| GET | `/api/v1/game/leaderboard` | 랭킹 (best score per user) | ✅ |

### API 사용 예시

**1. 로그인**
```bash
curl -X POST http://localhost:18081/v0.1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

Response:
```json
{
  "accessToken": "eyJhbGciOiJI...",
  "refreshToken": "eyJhbGciOiJ...",
  "expiresIn": 86400
}
```

**2. 파일 업로드 URL 요청**
```bash
curl -X POST http://localhost:18080/api/v1/files/upload \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "image.png",
    "file_type": "image",
    "content_type": "image/png",
    "file_size": 1024
  }'
```

Response:
```json
{
  "file_id": 1,
  "upload_url": "https://s3.amazonaws.com/...",
  "s3_key": "users/1/files/uuid-image.png",
  "thumbnail_upload_url": "https://s3.amazonaws.com/...",
  "thumbnail_key": "users/1/thumbnails/uuid-image_thumb.png",
  "expires_in": 900
}
```

**3. 사용자 통계 조회**
```bash
curl -X GET http://localhost:18080/api/v1/user/stats \
  -H "Authorization: Bearer {access_token}"
```

Response:
```json
{
  "storage": {
    "used": 9190,
    "total": 16106127360,
    "percentage": 0.000057
  },
  "monthlyStats": {
    "uploads": 5,
    "downloads": 12,
    "tagsCreated": 3
  }
}
```

**4. 활동 내역 조회**
```bash
curl -X GET "http://localhost:18080/api/v1/user/activity?month=2025-11" \
  -H "Authorization: Bearer {access_token}"
```

Response:
```json
{
  "2025-11-26": {
    "uploads": 3,
    "downloads": 5,
    "tags": ["vacation", "family"]
  }
}
```

## 🧪 Testing

```bash
# 단위 테스트 실행
go test ./...

# 커버리지 확인
go test -cover ./...

# E2E 테스트
cd services/cloudRepositoryService
go test -tags=e2e ./...
```

## 📈 Performance Features

- **Connection Pooling**: 데이터베이스 연결 최적화
- **Rate Limiting**: 10 RPS 제한으로 서버 보호
- **Presigned URLs**: S3 직접 업로드로 서버 부하 감소
- **Graceful Shutdown**: 안전한 서버 종료 처리
- **Context Timeout**: 30초 요청 타임아웃 설정

## 🔒 Security

- JWT 기반 인증 (Access Token: 24시간, Refresh Token: 7일)
- bcrypt를 활용한 비밀번호 해싱
- CORS 설정으로 허용된 오리진만 접근
- Rate Limiting으로 DDoS 공격 방지
- SQL Injection 방지 (GORM 파라미터 바인딩)
- 환경 변수를 통한 민감 정보 관리

## 🗂️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

### Cloud Files Table
```sql
CREATE TABLE cloud_files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    s3_key VARCHAR(512) UNIQUE NOT NULL,
    thumbnail_key VARCHAR(512),
    file_type VARCHAR(20) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_file_type (file_type)
);
```

### Activity Logs Table
```sql
CREATE TABLE activity_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    file_id BIGINT,
    activity_type VARCHAR(20) NOT NULL,
    tag_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_activity (user_id, activity_type, created_at)
);
```

## 📊 Monitoring & Logging

### Health Check
```bash
# Auth Service
curl http://localhost:18081/health

# Cloud Repository Service
curl http://localhost:18080/health
```

### Log Levels
- **DEBUG**: Detailed debugging information
- **INFO**: General information
- **WARN**: Warning messages
- **ERROR**: Error messages
- **FATAL**: Fatal errors causing service shutdown

### Metrics
- Request count and latency
- Error rates
- Storage usage per user
- API endpoint performance

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Coding Standards

- Follow Go best practices and idioms
- Use gofmt for code formatting
- Write unit tests for new features
- Maintain >80% code coverage
- Document exported functions and types
- Use meaningful variable and function names

## 🚢 Deployment

### Docker Deployment
```bash
# Build images
docker build -t joker-auth:latest ./services/authService
docker build -t joker-cloud:latest ./services/cloudRepositoryService

# Run with docker-compose
docker-compose up -d
```

### Production Considerations

- Use environment-specific configuration
- Enable TLS/SSL for HTTPS
- Set up proper logging and monitoring
- Configure auto-scaling policies
- Implement backup strategies for database
- Use secrets management for sensitive data

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**JokerTrickster**
- GitHub: [@JokerTrickster](https://github.com/JokerTrickster)

## 🙏 Acknowledgments

- Echo Framework for the excellent web framework
- GORM team for the powerful ORM
- AWS SDK Go team for S3 integration
- All contributors who helped improve this project

---

⭐ Star this repository if you find it helpful!