# Joker Backend 서버 명세서

> 개발 및 개선 시 참고하기 위한 종합 명세 문서

**최종 업데이트:** 2025-12-11
**작성자:** Claude Code
**버전:** 1.0

---

## 목차

1. [프로젝트 개요](#1-프로젝트-개요)
2. [아키텍처](#2-아키텍처)
3. [폴더 구조](#3-폴더-구조)
4. [구현된 기능](#4-구현된-기능)
5. [데이터베이스 스키마](#5-데이터베이스-스키마)
6. [코드 컨벤션](#6-코드-컨벤션)
7. [개발 가이드](#7-개발-가이드)

---

## 1. 프로젝트 개요

### 1.1 프로젝트 설명

Joker Backend는 마이크로서비스 아키텍처로 구성된 Go 기반의 백엔드 플랫폼입니다. 클라우드 파일 관리, 사용자 인증, 실시간 처리 상태 업데이트 등의 기능을 제공합니다.

### 1.2 기술 스택

**언어 & 프레임워크:**
- Go 1.21+
- Echo v4 (Web Framework)
- GORM (ORM)

**데이터베이스:**
- MySQL 8.0
- Redis

**인프라:**
- AWS S3 (파일 저장소)
- AWS SES (이메일)
- AWS SSM (시크릿 관리)
- Docker & Docker Compose

**외부 서비스:**
- Google OAuth
- FFmpeg (비디오 처리)

### 1.3 서비스 구성

| 서비스명 | 포트 | 설명 |
|---------|------|------|
| authService | 18081 | 사용자 인증 및 관리 |
| cloudRepositoryService | 18080 | 클라우드 파일 관리 및 처리 |

---

## 2. 아키텍처

### 2.1 전체 아키텍처

```
┌─────────────────────────────────────────────────────────┐
│                     Client Layer                        │
│                   (Web, Mobile, API)                    │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
┌───────▼──────────┐    ┌────────▼──────────┐
│  authService     │    │ cloudRepository   │
│  (Port 18081)    │    │ Service           │
│                  │    │ (Port 18080)      │
│  - 회원가입/로그인 │    │ - 파일 관리       │
│  - JWT 인증      │    │ - 폴더 구조       │
│  - Google OAuth  │    │ - 공유/권한       │
└─────────┬────────┘    │ - 비디오 처리     │
          │             │ - WebSocket       │
          │             └─────────┬─────────┘
          │                       │
          └───────┬───────────────┘
                  │
    ┌─────────────┴─────────────┐
    │      Shared Module        │
    │  - Config, Logger         │
    │  - DB, Redis, AWS         │
    │  - Middleware, Utils      │
    └─────────────┬─────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
┌───▼────┐  ┌──────┐  ┌────────▼───┐
│ MySQL  │  │Redis │  │  AWS S3    │
└────────┘  └──────┘  └────────────┘
```

### 2.2 클린 아키텍처 (Clean Architecture)

각 서비스는 계층형 아키텍처를 따릅니다:

```
┌─────────────────────────────────────────┐
│         Handler Layer                   │
│    (HTTP 요청/응답, 입력 검증)             │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│         UseCase Layer                   │
│   (비즈니스 로직, 트랜잭션 관리)            │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│       Repository Layer                  │
│  (데이터 접근, 외부 서비스 호출)            │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│    Database / External Services         │
│      (MySQL, S3, Redis, etc)            │
└─────────────────────────────────────────┘
```

**계층별 책임:**

1. **Handler Layer (프레젠테이션 계층)**
   - HTTP 요청/응답 처리
   - 입력 데이터 검증 (Validation)
   - JWT 토큰 추출 및 검증
   - DTO 변환 (Request ↔ Domain Model)

2. **UseCase Layer (애플리케이션 계층)**
   - 비즈니스 로직 구현
   - 트랜잭션 관리
   - Repository 호출 조율
   - 컨텍스트 타임아웃 처리 (기본 30초)

3. **Repository Layer (인프라 계층)**
   - 데이터베이스 CRUD 작업
   - 외부 서비스 호출 (S3, Redis, etc)
   - 데이터 영속성 관리

4. **Model Layer**
   - **Entity**: 도메인 모델 (DB 테이블)
   - **Interface**: 계약 정의
   - **Request**: 입력 DTO
   - **Response**: 출력 DTO

### 2.3 의존성 흐름

```
Handler → UseCase → Repository → Database/External Services
   ↓         ↓          ↓
Request   Entity    Entity
   ↓         ↓          ↓
Response  Response  Database
```

### 2.4 Middleware 스택

요청 처리 시 다음 순서로 Middleware가 실행됩니다:

```
1. Request ID 생성
   ↓
2. Recovery (Panic 복구)
   ↓
3. Request Logging
   ↓
4. CORS 처리
   ↓
5. Rate Limiting (10 req/s, burst 20)
   ↓
6. Timeout (30초)
   ↓
7. Handler 실행
```

---

## 3. 폴더 구조

### 3.1 전체 프로젝트 구조

```
joker_backend/
├── services/                    # 마이크로서비스들
│   ├── authService/            # 인증 서비스
│   └── cloudRepositoryService/ # 클라우드 저장소 서비스
├── shared/                      # 공유 모듈
│   ├── aws/                    # AWS SDK 통합
│   ├── config/                 # 설정 관리
│   ├── db/                     # DB 연결
│   ├── errors/                 # 커스텀 에러
│   ├── jwt/                    # JWT 유틸리티
│   ├── logger/                 # 로깅 (zap)
│   ├── middleware/             # HTTP 미들웨어
│   ├── migrate/                # 마이그레이션 유틸
│   ├── models/                 # 공유 모델
│   └── utils/                  # 공통 유틸리티
├── migrations/                  # DB 마이그레이션 파일
├── scripts/                     # 유틸리티 스크립트
│   ├── deploy-service.sh       # 서비스 배포
│   ├── migrate.sh              # DB 마이그레이션 실행
│   └── test-deployment.sh      # 배포 테스트
├── docs/                        # 문서
├── claudedocs/                  # Claude 생성 문서
├── docker-compose.yml           # Docker Compose 설정
├── go.work                      # Go Workspace
└── README.md
```

### 3.2 서비스 내부 구조

각 서비스는 동일한 구조를 따릅니다:

```
service/
├── cmd/
│   └── main.go                  # 애플리케이션 진입점
├── features/                    # 기능별 모듈
│   └── [feature-name]/
│       ├── handler/             # HTTP 핸들러
│       │   └── [feature]Handler.go
│       ├── usecase/             # 비즈니스 로직
│       │   └── [feature]UseCase.go
│       ├── repository/          # 데이터 접근
│       │   └── [feature]Repository.go
│       └── model/
│           ├── entity/          # 도메인 엔티티
│           ├── interface/       # 인터페이스 정의
│           ├── request/         # 요청 DTO
│           └── response/        # 응답 DTO
├── pkg/                         # 서비스 전용 패키지
│   ├── queue/                   # 메시지 큐
│   ├── websocket/               # WebSocket
│   └── processor/               # 비디오 프로세서
├── tests/                       # 테스트 파일
├── Dockerfile
├── go.mod
└── go.sum
```

**예시: cloudRepositoryService**

```
cloudRepositoryService/
├── cmd/
│   └── main.go
├── features/
│   └── cloudRepository/
│       ├── handler/
│       │   ├── uploadCloudHandler.go
│       │   ├── downloadCloudHandler.go
│       │   ├── folderHandler.go
│       │   ├── folderShareHandler.go
│       │   ├── fileShareHandler.go
│       │   ├── tagsHandler.go
│       │   ├── favoritesHandler.go
│       │   └── multipartUploadHandler.go
│       ├── usecase/
│       │   ├── uploadCloudUseCase.go
│       │   ├── downloadCloudUseCase.go
│       │   ├── folderUseCase.go
│       │   ├── folderShareUseCase.go
│       │   ├── fileShareUseCase.go
│       │   ├── tagsUseCase.go
│       │   ├── favoritesUseCase.go
│       │   └── multipartUploadUseCase.go
│       ├── repository/
│       │   ├── uploadCloudRepository.go
│       │   ├── downloadCloudRepository.go
│       │   ├── folderRepository.go
│       │   ├── folderShareRepository.go
│       │   ├── fileShareRepository.go
│       │   ├── tagsRepository.go
│       │   ├── favoritesRepository.go
│       │   └── multipartUploadRepository.go
│       └── model/
│           ├── entity/
│           │   ├── cloudFile.go
│           │   ├── folder.go
│           │   ├── folderShare.go
│           │   ├── fileShare.go
│           │   ├── tag.go
│           │   ├── favorite.go
│           │   └── multipartUpload.go
│           ├── interface/
│           │   └── ICloudRepository.go
│           ├── request/
│           │   └── cloudRepositoryReq.go
│           └── response/
│               └── cloudRepositoryRes.go
├── pkg/
│   ├── queue/
│   │   └── videoQueue.go
│   ├── websocket/
│   │   └── handler.go
│   └── processor/
│       └── videoProcessor.go
└── tests/
```

### 3.3 Shared 모듈 상세

```
shared/
├── aws/
│   ├── s3.go                    # S3 클라이언트
│   ├── ses.go                   # 이메일 서비스
│   └── ssm.go                   # 시크릿 관리
├── config/
│   └── config.go                # 환경 설정
├── db/
│   ├── mysql/
│   │   └── mysql.go             # MySQL 연결
│   └── redis/
│       └── redis.go             # Redis 연결
├── errors/
│   └── errors.go                # 커스텀 에러 타입
├── jwt/
│   └── jwt.go                   # JWT 생성/검증
├── logger/
│   └── logger.go                # 구조화된 로깅 (zap)
├── middleware/
│   ├── cors.go                  # CORS 설정
│   ├── ratelimit.go             # Rate Limiting
│   └── middleware.go            # 공통 미들웨어
├── migrate/
│   └── migrate.go               # DB 마이그레이션
├── models/
│   └── user.go                  # 공유 모델
├── utils/
│   ├── response/
│   │   └── response.go          # 응답 유틸
│   ├── validation.go            # 검증 유틸
│   └── echo_validator.go        # Echo 검증기
└── init.go                      # 초기화 로직
```

---

## 4. 구현된 기능

### 4.1 authService (인증 서비스)

**Base URL:** `http://localhost:18081`

#### 4.1.1 회원가입/로그인

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/v0.1/auth/signup` | 회원가입 | ❌ |
| POST | `/v0.1/auth/signin` | 로그인 (이메일/비밀번호) | ❌ |
| POST | `/v0.1/auth/google/signin` | Google OAuth 로그인 | ❌ |
| POST | `/v0.1/auth/refresh` | Access Token 갱신 | ❌ (Refresh Token 필요) |
| POST | `/v0.1/auth/logout` | 로그아웃 | ✅ |
| POST | `/v0.1/auth/check-email` | 이메일 중복 확인 | ❌ |

**주요 기능:**
- JWT 기반 인증 (Access Token + Refresh Token)
- bcrypt 비밀번호 해싱
- Google OAuth 통합
  - 신규 사용자 자동 생성
  - 삭제된 계정 자동 복원 (Soft-Delete 복구)
- 이메일 검증
- Soft Delete 지원

#### 4.1.2 Google OAuth 동작 상세

**인증 플로우:**
```
1. Client → Google OAuth 로그인
2. Google → ID Token 발급
3. Client → POST /v0.1/auth/google/signin (ID Token 전송)
4. Server → ID Token 검증
5. Server → 사용자 조회/생성/복원
6. Server → JWT 토큰 발급 (Access + Refresh)
7. Server → Client에 토큰 반환
```

**사용자 상태별 처리:**

| 상태 | 조건 | 동작 | 결과 |
|------|------|------|------|
| 신규 사용자 | 이메일이 DB에 없음 | 새 사용자 생성 (provider='google') | 201 Created + JWT 토큰 |
| 기존 사용자 | 이메일 존재, deleted_at=NULL | 기존 사용자 조회 | 200 OK + JWT 토큰 |
| 삭제된 사용자 | 이메일 존재, deleted_at!=NULL | 계정 복원 (deleted_at=NULL) | 200 OK + JWT 토큰 |

**계정 복원 동작:**
- 삭제된 계정(Soft-Delete)으로 Google 로그인 시 자동 복원
- 기존 사용자 데이터 완전 보존:
  - User ID 유지 (외래 키 관계 유지)
  - created_at 타임스탬프 유지 (감사 추적)
  - 모든 사용자 메타데이터 유지
- 복원 프로세스:
  1. `.Unscoped()` 쿼리로 삭제된 사용자 조회
  2. `deleted_at` 필드를 NULL로 설정
  3. 데이터베이스에 업데이트
  4. 정상 로그인 플로우 진행
- 사용자 경험: 투명하게 처리 (에러 없이 정상 로그인)
- 보안: Google ID Token 검증으로 본인 인증 보장

**관련 문서:**
- [Google OAuth 설정 가이드](../GOOGLE_OAUTH_SETUP.md)
- [Google 로그인 테스트](../GOOGLE_LOGIN_TEST.md)
- [Soft-Delete 복원 기술 문서](./google_oauth_soft_delete_fix.md)

---

### 4.2 cloudRepositoryService (클라우드 저장소 서비스)

**Base URL:** `http://localhost:18080`

#### 4.2.1 파일 관리

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/api/v1/files/upload` | 업로드용 Presigned URL 요청 | ✅ |
| POST | `/api/v1/files/upload/batch` | 배치 업로드 URL 요청 | ✅ |
| POST | `/api/v1/files/:id/complete-upload` | 업로드 완료 및 처리 시작 | ✅ |
| GET | `/api/v1/files/:id/download` | 다운로드용 Presigned URL 요청 | ✅ |
| GET | `/api/v1/files` | 파일 목록 조회 (필터링 지원) | ✅ |
| DELETE | `/api/v1/files/:id` | 파일 삭제 | ✅ |

**파일 업로드 플로우:**
```
1. Client → POST /files/upload (파일 정보 전송)
2. Server → Presigned URL 생성 및 DB 레코드 생성
3. Client → S3로 직접 업로드 (Presigned URL 사용)
4. Client → POST /files/:id/complete-upload
5. Server → 비디오 처리 큐에 추가 (비동기)
6. Background → FFmpeg으로 썸네일 생성, 메타데이터 추출
7. WebSocket → 처리 상태 실시간 업데이트
```

**지원 파일 타입:**
- `image`: 이미지 파일
- `video`: 비디오 파일 (자동 처리)

#### 4.2.2 폴더 관리

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/api/v1/folders` | 폴더 생성 | ✅ |
| GET | `/api/v1/folders` | 폴더 목록 조회 | ✅ |
| GET | `/api/v1/folders/:id` | 폴더 상세 조회 | ✅ |
| PATCH | `/api/v1/folders/:id` | 폴더 수정 | ✅ |
| DELETE | `/api/v1/folders/:id` | 폴더 삭제 | ✅ |
| GET | `/api/v1/folders/:id/files` | 폴더 내 파일 목록 | ✅ |
| POST | `/api/v1/files/batch/move` | 파일 일괄 이동 | ✅ |

**폴더 구조:**
- 계층적 폴더 구조 지원 (parent-child)
- 폴더 삭제 시 하위 폴더 및 파일도 연쇄 삭제 (CASCADE)

#### 4.2.3 폴더 공유

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/api/v1/folders/:id/share` | 폴더 공유 생성 | ✅ |
| GET | `/api/v1/folders/:id/shares` | 폴더 공유 목록 조회 | ✅ |
| DELETE | `/api/v1/folders/:id/shares/:userId` | 폴더 공유 취소 | ✅ |
| GET | `/api/v1/folders/shared-with-me` | 나에게 공유된 폴더 목록 | ✅ |
| GET | `/api/v1/folders/shared-by-me` | 내가 공유한 폴더 목록 | ✅ |

**권한 타입:**
- `read`: 읽기 전용
- `write`: 읽기 + 쓰기

#### 4.2.4 파일 공유

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/api/v1/files/:id/share` | 파일 공유 생성 | ✅ |
| GET | `/api/v1/files/:id/shares` | 파일 공유 목록 조회 | ✅ |
| DELETE | `/api/v1/files/:id/shares/:userId` | 파일 공유 취소 | ✅ |
| GET | `/api/v1/files/shared-with-me` | 나에게 공유된 파일 목록 | ✅ |
| GET | `/api/v1/files/shared-by-me` | 내가 공유한 파일 목록 | ✅ |

#### 4.2.5 태그 관리

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| PUT | `/api/v1/files/:id/tags` | 파일 태그 전체 업데이트 | ✅ |
| POST | `/api/v1/files/:id/tags` | 파일에 태그 추가 | ✅ |
| DELETE | `/api/v1/files/:id/tags/:tag_name` | 파일에서 태그 제거 | ✅ |

**태그 특징:**
- 파일과 Many-to-Many 관계
- 사용자별 태그 관리
- 파일 필터링에 사용 가능

#### 4.2.6 즐겨찾기

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/api/v1/favorites` | 즐겨찾기 추가 | ✅ |
| DELETE | `/api/v1/favorites/:fileId` | 즐겨찾기 제거 | ✅ |
| GET | `/api/v1/favorites` | 즐겨찾기 목록 조회 | ✅ |

#### 4.2.7 멀티파트 업로드 (대용량 파일)

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| POST | `/api/v1/files/multipart/initiate` | 멀티파트 업로드 시작 | ✅ |
| POST | `/api/v1/files/multipart/presigned-urls` | 파트별 Presigned URL 생성 | ✅ |
| POST | `/api/v1/files/multipart/complete` | 멀티파트 업로드 완료 | ✅ |
| POST | `/api/v1/files/multipart/abort` | 멀티파트 업로드 중단 | ✅ |

**멀티파트 업로드 플로우:**
```
1. Client → POST /multipart/initiate
   - 파일 정보 전송
   - Upload ID 생성

2. Client → POST /multipart/presigned-urls
   - 파트 번호 리스트 전송
   - 각 파트별 Presigned URL 수신

3. Client → S3에 각 파트 병렬 업로드
   - Part Number와 ETag 저장

4. Client → POST /multipart/complete
   - 모든 Part Number와 ETag 전송
   - S3 멀티파트 완료 처리
```

#### 4.2.8 사용자 통계

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| GET | `/api/v1/user/stats` | 스토리지 사용량 통계 | ✅ |
| GET | `/api/v1/user/activity` | 활동 내역 조회 | ✅ |

**통계 정보:**
- 총 파일 개수
- 총 스토리지 사용량
- 파일 타입별 분포
- 최근 활동 내역

#### 4.2.9 처리 상태

| Method | Endpoint | 설명 | 인증 필요 |
|--------|----------|------|----------|
| GET | `/api/v1/files/:id/processing-status` | 파일 처리 상태 조회 | ✅ |
| POST | `/api/v1/files/processing-status/batch` | 여러 파일 처리 상태 조회 | ✅ |

**처리 상태:**
- `pending`: 대기 중
- `processing`: 처리 중
- `completed`: 완료
- `failed`: 실패

**처리 단계:**
- `validating_file`: 파일 검증
- `extracting_metadata`: 메타데이터 추출
- `generating_thumbnail`: 썸네일 생성
- `uploading_thumbnail`: 썸네일 업로드
- `finalizing`: 마무리
- `done`: 완료

#### 4.2.10 WebSocket

| Protocol | Endpoint | 설명 |
|----------|----------|------|
| WS | `/ws/?token=<jwt_token>` | 실시간 처리 상태 업데이트 |

**WebSocket 메시지 형식:**
```json
{
  "type": "processing_update",
  "file_id": 123,
  "status": "processing",
  "progress": 50,
  "stage": "generating_thumbnail",
  "timestamp": "2025-12-11T10:00:00Z"
}
```

---

## 5. 데이터베이스 스키마

### 5.1 ERD (Entity Relationship Diagram)

```
┌─────────────┐
│    users    │
│─────────────│
│ id (PK)     │────┐
│ name        │    │
│ email       │    │
│ password    │    │
│ provider    │    │
│ created_at  │    │
│ updated_at  │    │
│ deleted_at  │    │
└─────────────┘    │
                   │
                   │ 1:N
                   │
    ┌──────────────┴──────────────┐
    │                             │
┌───▼──────────┐          ┌───────▼────────┐
│ cloud_files  │          │    folders     │
│──────────────│          │────────────────│
│ id (PK)      │◄──┐      │ id (PK)        │
│ user_id (FK) │   │      │ user_id (FK)   │
│ folder_id(FK)│───┘      │ parent_id (FK) │◄─┐
│ file_name    │          │ folder_name    │  │
│ s3_key       │          │ created_at     │  │
│ file_type    │          │ updated_at     │  │
│ file_size    │          │ deleted_at     │  │
│ duration     │          └────────────────┘  │
│ proc_status  │                   │          │
│ proc_progress│                   └──────────┘
│ ...          │                   (Self-Reference)
└──────────────┘
      │ M:N
      │
┌─────▼──────┐         ┌──────────────┐
│   tags     │◄───────►│  file_tags   │
│────────────│         │──────────────│
│ id (PK)    │         │ tag_id (FK)  │
│ user_id    │         │ file_id (FK) │
│ name       │         └──────────────┘
└────────────┘

┌──────────────┐        ┌─────────────────┐
│ folder_shares│        │   file_shares   │
│──────────────│        │─────────────────│
│ id (PK)      │        │ id (PK)         │
│ folder_id(FK)│        │ file_id (FK)    │
│ owner_id     │        │ owner_id        │
│ shared_with  │        │ shared_with     │
│ permission   │        │ permission      │
└──────────────┘        └─────────────────┘

┌──────────────┐        ┌─────────────────┐
│  favorites   │        │ activity_logs   │
│──────────────│        │─────────────────│
│ id (PK)      │        │ id (PK)         │
│ user_id (FK) │        │ user_id (FK)    │
│ file_id (FK) │        │ file_id (FK)    │
│ favorited_at │        │ activity_type   │
└──────────────┘        │ tag_name        │
                        │ created_at      │
                        └─────────────────┘
```

### 5.2 테이블 상세 스키마

#### 5.2.1 users (사용자)

```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL DEFAULT '',
    provider VARCHAR(50) NOT NULL DEFAULT 'local',
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,

    INDEX idx_email (email),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**필드 설명:**
- `provider`: 인증 제공자 (`local`, `google`)
- `deleted_at`: Soft Delete용 필드

#### 5.2.2 cloud_files (파일)

```sql
CREATE TABLE cloud_files (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    folder_id BIGINT UNSIGNED NULL,
    file_name VARCHAR(255) NOT NULL,
    s3_key VARCHAR(512) NOT NULL UNIQUE,
    thumbnail_key VARCHAR(512) NULL,
    file_type VARCHAR(20) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    duration DECIMAL(10,2) NULL,
    processing_status VARCHAR(20) DEFAULT 'pending',
    processing_progress INT DEFAULT 0,
    processing_stage VARCHAR(50) NULL,
    processing_error VARCHAR(512) NULL,
    processing_started_at TIMESTAMP NULL,
    processing_completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,

    INDEX idx_user_id (user_id),
    INDEX idx_folder_id (folder_id),
    INDEX idx_file_type (file_type),
    INDEX idx_processing_status (processing_status),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**필드 설명:**
- `s3_key`: S3 객체 키 (고유값)
- `thumbnail_key`: 썸네일 S3 키 (비디오만 해당)
- `file_type`: `image`, `video`
- `duration`: 비디오 재생 시간 (초)
- `processing_status`: `pending`, `processing`, `completed`, `failed`
- `processing_stage`: 현재 처리 단계
- `processing_progress`: 처리 진행률 (0-100)

#### 5.2.3 folders (폴더)

```sql
CREATE TABLE folders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    folder_name VARCHAR(255) NOT NULL,
    parent_folder_id BIGINT UNSIGNED NULL,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,

    INDEX idx_user_id (user_id),
    INDEX idx_parent_folder_id (parent_folder_id),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**필드 설명:**
- `parent_folder_id`: 부모 폴더 ID (계층 구조)
- NULL이면 루트 폴더

#### 5.2.4 folder_shares (폴더 공유)

```sql
CREATE TABLE folder_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    folder_id BIGINT UNSIGNED NOT NULL,
    owner_id INT NOT NULL,
    shared_with_id INT NOT NULL,
    permission VARCHAR(10) NOT NULL DEFAULT 'read',
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,

    UNIQUE KEY uniq_folder_user (folder_id, shared_with_id, deleted_at),
    INDEX idx_folder_id (folder_id),
    INDEX idx_shared_with_id (shared_with_id),
    INDEX idx_permission (permission),

    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**필드 설명:**
- `permission`: `read`, `write`
- `owner_id`: 공유자 (소유자)
- `shared_with_id`: 피공유자

#### 5.2.5 file_shares (파일 공유)

```sql
CREATE TABLE file_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_id BIGINT UNSIGNED NOT NULL,
    owner_id INT NOT NULL,
    shared_with_id INT NOT NULL,
    permission VARCHAR(10) NOT NULL DEFAULT 'read',
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,

    UNIQUE KEY uniq_file_user (file_id, shared_with_id, deleted_at),
    INDEX idx_file_id (file_id),
    INDEX idx_shared_with_id (shared_with_id),
    INDEX idx_permission (permission),

    FOREIGN KEY (file_id) REFERENCES cloud_files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 5.2.6 tags (태그)

```sql
CREATE TABLE tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NULL,

    INDEX idx_user_id (user_id),
    UNIQUE KEY uniq_user_tag (user_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 5.2.7 file_tags (파일-태그 매핑)

```sql
CREATE TABLE file_tags (
    tag_id BIGINT UNSIGNED NOT NULL,
    cloud_file_id BIGINT UNSIGNED NOT NULL,

    PRIMARY KEY (tag_id, cloud_file_id),

    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    FOREIGN KEY (cloud_file_id) REFERENCES cloud_files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 5.2.8 favorites (즐겨찾기)

```sql
CREATE TABLE favorites (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    file_id BIGINT UNSIGNED NOT NULL,
    favorited_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uniq_user_file (user_id, file_id),
    INDEX idx_user_favorited_at (user_id, favorited_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES cloud_files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 5.2.9 activity_logs (활동 로그)

```sql
CREATE TABLE activity_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    file_id BIGINT UNSIGNED NULL,
    activity_type VARCHAR(20) NOT NULL,
    tag_name VARCHAR(100) NULL,
    created_at TIMESTAMP NULL,

    INDEX idx_user_activity (user_id, activity_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**필드 설명:**
- `activity_type`: `upload`, `download`, `tag_add`, `tag_del`

#### 5.2.10 multipart_uploads (멀티파트 업로드)

```sql
CREATE TABLE multipart_uploads (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    upload_id VARCHAR(255) NOT NULL,
    file_key VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NULL,
    part_size INT NOT NULL,
    total_parts INT NOT NULL,
    status ENUM('initiated', 'uploading', 'completed', 'aborted') DEFAULT 'initiated',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_upload_id (upload_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 5.3 주요 관계

```
users (1) ───── (N) cloud_files
users (1) ───── (N) folders
folders (1) ─── (N) cloud_files
folders (1) ─── (N) folders (Self-Reference)

cloud_files (N) ─── (M) tags (via file_tags)
users (N) ──────── (M) cloud_files (via favorites)

folder_shares: owner/shared_with → users
file_shares: owner/shared_with → users
```

---

## 6. 코드 컨벤션

### 6.1 파일 네이밍

**Handler 파일:**
```
[feature]Handler.go
예: uploadCloudHandler.go, folderHandler.go
```

**UseCase 파일:**
```
[feature]UseCase.go
예: uploadCloudUseCase.go, folderUseCase.go
```

**Repository 파일:**
```
[feature]Repository.go
예: uploadCloudRepository.go, folderRepository.go
```

**Entity 파일:**
```
[entity].go
예: cloudFile.go, folder.go, user.go
```

**Interface 파일:**
```
I[Type][Feature].go
예: ICloudRepositoryHandler.go, IFolderUseCase.go
```

### 6.2 타입 네이밍

**Struct 타입:**
```go
// Handler
type UploadCloudRepositoryHandler struct { ... }

// UseCase
type UploadCloudRepositoryUseCase struct { ... }

// Repository
type UploadCloudRepositoryRepository struct { ... }

// Entity
type CloudFile struct { ... }
type Folder struct { ... }
```

### 6.3 함수 네이밍

**Constructor 함수:**
```go
func NewUploadCloudRepositoryHandler(...) IUploadCloudRepositoryHandler
func NewFolderUseCase(...) IFolderUseCase
```

**Handler 메서드:**
```go
// HTTP 액션에 맞춤
func (h *Handler) RequestUploadURL(c echo.Context) error
func (h *Handler) ListFiles(c echo.Context) error
func (h *Handler) CreateFolder(c echo.Context) error
```

**UseCase 메서드:**
```go
// 비즈니스 액션
func (u *UseCase) RequestUploadURL(ctx context.Context, ...) error
func (u *UseCase) CreateFolder(ctx context.Context, ...) error
```

**Repository 메서드:**
```go
// 데이터 작업
func (r *Repository) CreateFile(ctx context.Context, file *entity.CloudFile) error
func (r *Repository) FindByID(ctx context.Context, id uint) (*entity.CloudFile, error)
func (r *Repository) Update(ctx context.Context, file *entity.CloudFile) error
func (r *Repository) Delete(ctx context.Context, id uint) error
```

### 6.4 패키지 구성

**Import 순서:**
```go
import (
    // 1. 표준 라이브러리
    "context"
    "fmt"
    "time"

    // 2. 외부 패키지
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"

    // 3. Shared 패키지
    "github.com/JokerTrickster/joker_backend/shared/logger"
    "github.com/JokerTrickster/joker_backend/shared/errors"

    // 4. 현재 서비스 패키지
    "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
)
```

**패키지 별칭:**
```go
import (
    _interface "path/to/interface"  // 예약어와 충돌 시
    _aws "github.com/aws/aws-sdk-go-v2"
)
```

### 6.5 에러 처리

**커스텀 에러 타입 (shared/errors):**
```go
errors.BadRequest("invalid file type")
errors.Unauthorized("invalid token")
errors.Forbidden("no permission")
errors.NotFound("file not found")
errors.ValidationError("file size too large")
errors.InternalServerError("unexpected error")
errors.DatabaseError(err)
```

**Handler 레이어:**
```go
if err != nil {
    logger.Error("failed to upload file", zap.Error(err))
    return c.JSON(http.StatusInternalServerError, map[string]string{
        "error": err.Error(),
    })
}
```

**UseCase 레이어:**
```go
if err != nil {
    return nil, fmt.Errorf("failed to create file: %w", err)
}
```

**에러 체이닝:**
```go
// fmt.Errorf with %w for error wrapping
err := repo.CreateFile(ctx, file)
if err != nil {
    return fmt.Errorf("failed to create file in database: %w", err)
}
```

### 6.6 로깅

**Zap 구조화 로깅:**
```go
import "github.com/JokerTrickster/joker_backend/shared/logger"
import "go.uber.org/zap"

// Info 레벨
logger.Info("file uploaded",
    zap.String("file_id", fileID),
    zap.String("user_id", userID))

// Error 레벨
logger.Error("upload failed",
    zap.Error(err),
    zap.String("s3_key", s3Key))

// Warning 레벨
logger.Warn("processing delayed",
    zap.Int("queue_size", queueSize))

// Debug 레벨
logger.Debug("processing file",
    zap.String("stage", stage))
```

**로그 레벨:**
- `Debug`: 디버깅 정보
- `Info`: 일반 정보
- `Warn`: 경고 (처리는 계속)
- `Error`: 에러 (처리 실패)
- `Fatal`: 치명적 에러 (프로세스 종료)

### 6.7 Validation

**Request DTO 검증:**
```go
type UploadRequestDTO struct {
    FileName    string `json:"file_name" validate:"required"`
    FileType    string `json:"file_type" validate:"required,oneof=image video"`
    ContentType string `json:"content_type" validate:"required"`
    FileSize    int64  `json:"file_size" validate:"required,gt=0"`
    FolderID    *uint  `json:"folder_id" validate:"omitempty"`
}
```

**Handler에서 검증:**
```go
var req UploadRequestDTO
if err := c.Bind(&req); err != nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": "invalid request format",
    })
}

if err := c.Validate(&req); err != nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": err.Error(),
    })
}
```

**Validation 태그:**
- `required`: 필수 필드
- `omitempty`: 선택 필드
- `oneof=a b c`: 허용 값 제한
- `gt=0`: 0보다 큰 값
- `email`: 이메일 형식
- `min=3,max=100`: 길이 제한

### 6.8 Context & Timeout

**UseCase 타임아웃 패턴:**
```go
type UploadCloudRepositoryUseCase struct {
    Repo           IUploadCloudRepositoryRepository
    ContextTimeout time.Duration
}

func (u *UploadCloudRepositoryUseCase) RequestUploadURL(
    c context.Context,
    req *request.UploadRequestDTO,
) (*response.UploadResponseDTO, error) {
    ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
    defer cancel()

    // ctx를 모든 하위 호출에 전달
    file, err := u.Repo.CreateFile(ctx, &entity.CloudFile{...})
    if err != nil {
        return nil, err
    }

    return &response.UploadResponseDTO{...}, nil
}
```

**표준 타임아웃:** 30초

### 6.9 의존성 주입

**Constructor 패턴:**
```go
// Handler
func NewUploadCloudRepositoryHandler(
    c *echo.Group,
    useCase IUploadCloudRepositoryUseCase,
) IUploadCloudRepositoryHandler {
    handler := &UploadCloudRepositoryHandler{
        UseCase: useCase,
    }

    // 라우트 등록
    c.POST("/files/upload", handler.RequestUploadURL)
    c.POST("/files/upload/batch", handler.BatchUpload)

    return handler
}

// UseCase
func NewUploadCloudRepositoryUseCase(
    repo IUploadCloudRepositoryRepository,
    timeout time.Duration,
) IUploadCloudRepositoryUseCase {
    return &UploadCloudRepositoryUseCase{
        Repo:           repo,
        ContextTimeout: timeout,
    }
}

// Repository
func NewUploadCloudRepositoryRepository(
    db *gorm.DB,
    s3Client *s3.S3Client,
) IUploadCloudRepositoryRepository {
    return &UploadCloudRepositoryRepository{
        DB:       db,
        S3Client: s3Client,
    }
}
```

### 6.10 인터페이스 정의

**Interface 네이밍 및 구조:**
```go
// Handler 인터페이스
type IUploadCloudRepositoryHandler interface {
    RequestUploadURL(c echo.Context) error
    BatchUpload(c echo.Context) error
    CompleteUpload(c echo.Context) error
}

// UseCase 인터페이스
type IUploadCloudRepositoryUseCase interface {
    RequestUploadURL(ctx context.Context, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error)
    BatchUpload(ctx context.Context, req *request.BatchUploadRequestDTO) (*response.BatchUploadResponseDTO, error)
}

// Repository 인터페이스
type IUploadCloudRepositoryRepository interface {
    CreateFile(ctx context.Context, file *entity.CloudFile) error
    FindByID(ctx context.Context, id uint) (*entity.CloudFile, error)
    GeneratePresignedURL(ctx context.Context, s3Key string) (string, error)
}
```

### 6.11 Response 패턴

**성공 응답:**
```go
return c.JSON(http.StatusOK, responseDTO)
```

**에러 응답:**
```go
return c.JSON(http.StatusBadRequest, map[string]string{
    "error": "error message",
})
```

**표준 HTTP 상태 코드:**
- `200 OK`: 성공
- `201 Created`: 생성 성공
- `400 Bad Request`: 잘못된 요청
- `401 Unauthorized`: 인증 실패
- `403 Forbidden`: 권한 없음
- `404 Not Found`: 리소스 없음
- `500 Internal Server Error`: 서버 오류

### 6.12 Database 패턴

**GORM Soft Delete:**
```go
type CloudFile struct {
    ID        uint           `gorm:"primaryKey"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
    // ...
}
```

**트랜잭션:**
```go
err := r.DB.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&file).Error; err != nil {
        return err
    }

    if err := tx.Create(&activityLog).Error; err != nil {
        return err
    }

    return nil
})
```

**Preloading:**
```go
var files []entity.CloudFile
err := r.DB.
    Preload("Tags").
    Preload("Folder").
    Where("user_id = ?", userID).
    Find(&files).Error
```

**Raw SQL (필요시):**
```go
err := r.DB.Raw(`
    SELECT f.* FROM cloud_files f
    INNER JOIN file_shares s ON f.id = s.file_id
    WHERE s.shared_with_id = ?
`, userID).Scan(&files).Error
```

### 6.13 공통 패턴

**사용되는 디자인 패턴:**

1. **Repository Pattern** - 데이터 접근 추상화
2. **Dependency Injection** - Constructor 기반
3. **Interface Segregation** - 작고 명확한 인터페이스
4. **Factory Pattern** - New* 생성자 함수
5. **Clean Architecture** - 계층 분리
6. **DTO Pattern** - Request/Response 분리
7. **Middleware Chain** - Echo 미들웨어 스택
8. **Context Propagation** - 타임아웃 및 취소
9. **Soft Delete** - GORM DeletedAt 필드
10. **Presigned URLs** - S3 직접 업로드/다운로드

---

## 7. 개발 가이드

### 7.1 개발 환경 설정

#### 필수 요구사항

**소프트웨어:**
- Go 1.21 이상
- Docker & Docker Compose
- MySQL 8.0
- Redis
- FFmpeg (비디오 처리용)

**AWS 리소스:**
- S3 버킷
- IAM 사용자 (S3, SES, SSM 권한)
- SSM Parameter Store (시크릿 관리)

#### 환경 변수 설정

각 서비스의 `.env` 파일 또는 환경 변수:

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=joker_backend

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# AWS
AWS_REGION=ap-northeast-2
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
CLOUD_REPOSITORY_BUCKET=your-bucket-name

# JWT
JWT_SECRET=your_jwt_secret
JWT_ACCESS_EXPIRATION=3600  # 1 hour
JWT_REFRESH_EXPIRATION=2592000  # 30 days

# Server
PORT=18080  # cloudRepositoryService
# PORT=18081  # authService
ENV=development  # development, production
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Google OAuth (authService only)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
```

### 7.2 로컬 개발

#### 데이터베이스 설정

```bash
# 1. MySQL 컨테이너 실행
docker-compose up -d mysql redis

# 2. 마이그레이션 실행
./scripts/migrate.sh
```

#### 서비스 실행

```bash
# authService 실행
cd services/authService
go run cmd/main.go

# cloudRepositoryService 실행
cd services/cloudRepositoryService
go run cmd/main.go
```

#### Docker로 실행

```bash
# 전체 서비스 실행
docker-compose up -d

# 특정 서비스만 실행
docker-compose up -d cloudrepository_api

# 로그 확인
docker-compose logs -f cloudrepository_api
```

### 7.3 새 기능 추가 가이드

#### 1. 새 API 엔드포인트 추가

**Step 1: Entity 정의 (필요시)**
```go
// features/[feature]/model/entity/[entity].go
package entity

type NewEntity struct {
    ID        uint           `gorm:"primaryKey"`
    UserID    uint           `gorm:"not null;index"`
    Name      string         `gorm:"size:255;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

**Step 2: Request/Response DTO 정의**
```go
// features/[feature]/model/request/[feature]Req.go
package request

type CreateNewEntityDTO struct {
    Name string `json:"name" validate:"required,min=3"`
}

// features/[feature]/model/response/[feature]Res.go
package response

type NewEntityResponseDTO struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```

**Step 3: Interface 정의**
```go
// features/[feature]/model/interface/I[Feature].go
package _interface

type INewFeatureRepository interface {
    Create(ctx context.Context, entity *entity.NewEntity) error
    FindByID(ctx context.Context, id uint) (*entity.NewEntity, error)
}

type INewFeatureUseCase interface {
    Create(ctx context.Context, req *request.CreateNewEntityDTO) (*response.NewEntityResponseDTO, error)
}

type INewFeatureHandler interface {
    Create(c echo.Context) error
}
```

**Step 4: Repository 구현**
```go
// features/[feature]/repository/[feature]Repository.go
package repository

type NewFeatureRepository struct {
    DB *gorm.DB
}

func NewNewFeatureRepository(db *gorm.DB) _interface.INewFeatureRepository {
    return &NewFeatureRepository{DB: db}
}

func (r *NewFeatureRepository) Create(ctx context.Context, entity *entity.NewEntity) error {
    return r.DB.WithContext(ctx).Create(entity).Error
}
```

**Step 5: UseCase 구현**
```go
// features/[feature]/usecase/[feature]UseCase.go
package usecase

type NewFeatureUseCase struct {
    Repo           _interface.INewFeatureRepository
    ContextTimeout time.Duration
}

func NewNewFeatureUseCase(repo _interface.INewFeatureRepository, timeout time.Duration) _interface.INewFeatureUseCase {
    return &NewFeatureUseCase{
        Repo:           repo,
        ContextTimeout: timeout,
    }
}

func (u *NewFeatureUseCase) Create(c context.Context, req *request.CreateNewEntityDTO) (*response.NewEntityResponseDTO, error) {
    ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
    defer cancel()

    entity := &entity.NewEntity{
        Name: req.Name,
    }

    if err := u.Repo.Create(ctx, entity); err != nil {
        return nil, fmt.Errorf("failed to create entity: %w", err)
    }

    return &response.NewEntityResponseDTO{
        ID:        entity.ID,
        Name:      entity.Name,
        CreatedAt: entity.CreatedAt,
    }, nil
}
```

**Step 6: Handler 구현**
```go
// features/[feature]/handler/[feature]Handler.go
package handler

type NewFeatureHandler struct {
    UseCase _interface.INewFeatureUseCase
}

func NewNewFeatureHandler(c *echo.Group, useCase _interface.INewFeatureUseCase) _interface.INewFeatureHandler {
    handler := &NewFeatureHandler{UseCase: useCase}

    c.POST("/entities", handler.Create)

    return handler
}

func (h *NewFeatureHandler) Create(c echo.Context) error {
    var req request.CreateNewEntityDTO
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
    }

    if err := c.Validate(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    resp, err := h.UseCase.Create(c.Request().Context(), &req)
    if err != nil {
        logger.Error("failed to create entity", zap.Error(err))
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusCreated, resp)
}
```

**Step 7: main.go에서 DI 설정**
```go
// cmd/main.go
func main() {
    // ... 기존 설정 ...

    // Repository
    newFeatureRepo := repository.NewNewFeatureRepository(db)

    // UseCase
    newFeatureUseCase := usecase.NewNewFeatureUseCase(newFeatureRepo, 30*time.Second)

    // Handler
    handler.NewNewFeatureHandler(apiV1, newFeatureUseCase)

    // ... 서버 시작 ...
}
```

#### 2. 데이터베이스 마이그레이션

**마이그레이션 파일 생성:**
```bash
# migrations/ 디렉토리에 새 파일 생성
# 형식: YYYYMMDDHHMMSS_description.sql
# 예: 20251211120000_create_new_entity_table.sql
```

**마이그레이션 SQL 작성:**
```sql
-- migrations/20251211120000_create_new_entity_table.sql

-- Up Migration
CREATE TABLE new_entities (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,

    INDEX idx_user_id (user_id),
    INDEX idx_deleted_at (deleted_at),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Down Migration (주석으로 작성 또는 별도 파일)
-- DROP TABLE IF EXISTS new_entities;
```

**마이그레이션 실행:**
```bash
./scripts/migrate.sh
```

### 7.4 테스트

#### 단위 테스트 작성

```go
// features/[feature]/usecase/[feature]UseCase_test.go
package usecase_test

import (
    "testing"
    "context"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, entity *entity.NewEntity) error {
    args := m.Called(ctx, entity)
    return args.Error(0)
}

func TestCreate(t *testing.T) {
    mockRepo := new(MockRepository)
    useCase := usecase.NewNewFeatureUseCase(mockRepo, 30*time.Second)

    req := &request.CreateNewEntityDTO{
        Name: "test",
    }

    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    resp, err := useCase.Create(context.Background(), req)

    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, "test", resp.Name)
    mockRepo.AssertExpectations(t)
}
```

#### 통합 테스트 실행

```bash
# 전체 테스트
go test ./...

# 특정 패키지
go test ./features/cloudRepository/usecase

# 커버리지
go test -cover ./...
```

### 7.5 배포

#### 로컬 배포 테스트

```bash
# 서비스 배포 스크립트 실행
./scripts/deploy-service.sh cloudRepositoryService

# 배포 확인
./scripts/test-deployment.sh
```

#### Docker 이미지 빌드

```bash
# cloudRepositoryService 이미지 빌드
cd services/cloudRepositoryService
docker build -t joker-backend-cloudrepository:latest .

# authService 이미지 빌드
cd services/authService
docker build -t joker-backend-auth:latest .
```

#### 프로덕션 배포

```bash
# 1. 환경 변수 설정 확인
# 2. 데이터베이스 마이그레이션
DB_HOST=prod-db-host ./scripts/migrate.sh

# 3. Docker Compose로 배포
docker-compose -f docker-compose.prod.yml up -d

# 4. 헬스 체크
curl http://localhost:18080/health
curl http://localhost:18081/health
```

### 7.6 모니터링 & 디버깅

#### 로그 확인

```bash
# 실시간 로그
docker logs -f cloudrepositoryservice_api

# 최근 100줄
docker logs --tail 100 cloudrepositoryservice_api

# 에러만 필터링
docker logs cloudrepositoryservice_api 2>&1 | grep -i error
```

#### 헬스 체크

```bash
# cloudRepositoryService
curl http://localhost:18080/health

# authService
curl http://localhost:18081/health
```

#### 데이터베이스 디버깅

```bash
# MySQL 접속
docker exec -it joker_mysql mysql -uroot -p

# 쿼리 실행
USE joker_backend;
SELECT * FROM cloud_files WHERE processing_status = 'failed';
```

### 7.7 문제 해결

#### 자주 발생하는 문제

**1. 데이터베이스 연결 실패**
```bash
# MySQL 컨테이너 상태 확인
docker ps | grep mysql

# 연결 테스트
mysql -h localhost -P 3306 -u root -p
```

**2. Redis 연결 실패**
```bash
# Redis 컨테이너 상태 확인
docker ps | grep redis

# Redis ping
redis-cli ping
```

**3. S3 업로드 실패**
- AWS 자격 증명 확인
- S3 버킷 권한 확인
- CORS 설정 확인

**4. 비디오 처리 실패**
```bash
# FFmpeg 설치 확인
ffmpeg -version

# 로그에서 에러 확인
docker logs cloudrepositoryservice_api 2>&1 | grep -i ffmpeg
```

**5. WebSocket 연결 실패**
- JWT 토큰 유효성 확인
- CORS 설정 확인
- 프록시 설정 확인 (Nginx, etc)

### 7.8 성능 최적화

#### 데이터베이스 인덱스

```sql
-- 자주 조회되는 컬럼에 인덱스 추가
CREATE INDEX idx_user_file_type ON cloud_files(user_id, file_type);
CREATE INDEX idx_processing_status_created ON cloud_files(processing_status, created_at);
```

#### 쿼리 최적화

```go
// Preload 사용
db.Preload("Tags").Preload("Folder").Find(&files)

// 필요한 필드만 선택
db.Select("id", "file_name", "created_at").Find(&files)

// 페이지네이션
db.Offset(offset).Limit(limit).Find(&files)
```

#### Redis 캐싱

```go
// 사용자 통계 캐싱
key := fmt.Sprintf("user:stats:%d", userID)
cachedStats, err := redisClient.Get(ctx, key).Result()
if err == nil {
    // 캐시 hit
    return cachedStats, nil
}

// 캐시 miss - DB 조회 후 캐싱
stats := getStatsFromDB(userID)
redisClient.Set(ctx, key, stats, 5*time.Minute)
```

### 7.9 보안 고려사항

#### JWT 보안

- Access Token 만료 시간: 1시간
- Refresh Token 만료 시간: 30일
- 토큰은 HTTPS로만 전송
- Refresh Token은 httpOnly 쿠키 사용 권장

#### S3 보안

- Presigned URL 만료 시간: 15분
- 버킷 정책으로 직접 접근 차단
- CloudFront 사용 권장

#### 입력 검증

- 모든 사용자 입력은 검증
- SQL Injection 방지 (GORM 사용)
- XSS 방지 (출력 이스케이핑)
- 파일 타입 검증
- 파일 크기 제한

#### Rate Limiting

- 기본: 10 req/s, burst 20
- 필요시 IP별, 사용자별 제한 조정

---

## 8. 부록

### 8.1 유용한 명령어

```bash
# Go 모듈 관리
go mod tidy
go mod download
go mod verify

# 코드 포맷팅
go fmt ./...
gofmt -s -w .

# 정적 분석
go vet ./...
golangci-lint run

# 의존성 업데이트
go get -u ./...

# Docker 정리
docker system prune -a
docker volume prune
```

### 8.2 참고 자료

**공식 문서:**
- [Echo Framework](https://echo.labstack.com/)
- [GORM](https://gorm.io/)
- [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/)
- [Go 표준 라이브러리](https://pkg.go.dev/std)

**프로젝트 문서:**
- [README.md](../README.md)
- [API 문서](./API.md) (작성 예정)
- [배포 가이드](./DEPLOYMENT.md) (작성 예정)

### 8.3 연락처 & 지원

**이슈 리포팅:**
- GitHub Issues 사용

**개발팀:**
- Backend Team

---

**문서 버전:** 1.0
**마지막 업데이트:** 2025-12-11
**다음 리뷰 예정:** 2026-01-11
