# 비동기 처리 구현 가이드

## 📋 개요

대용량 파일 업로드 후 썸네일 생성과 Duration 추출을 백엔드에서 비동기로 처리하고, 완료 시 WebSocket으로 알림을 보내는 시스템입니다.

## 🎯 주요 기능

- ✅ S3 업로드 완료 후 백그라운드에서 비디오 처리
- ✅ FFmpeg를 이용한 썸네일 생성 및 Duration 추출
- ✅ Redis 기반 작업 큐 (Asynq)
- ✅ WebSocket 실시간 알림
- ✅ 재시도 메커니즘 (최대 3번)
- ✅ 브라우저를 닫아도 처리 계속 진행

## 🏗️ 아키텍처

```
Frontend (S3 Upload)
    ↓
POST /api/v1/files/:id/complete-upload
    ↓
Enqueue Video Processing Task (Redis)
    ↓
Background Worker (Asynq)
    ├─ Extract Duration (FFmpeg)
    ├─ Generate Thumbnail (FFmpeg)
    └─ Upload Thumbnail to S3
    ↓
Update Database
    ↓
WebSocket Notification → Frontend
```

## 📦 설치된 패키지

```bash
go get github.com/hibiken/asynq        # Redis 기반 작업 큐
go get github.com/gorilla/websocket    # WebSocket 서버
go get github.com/u2takey/ffmpeg-go    # FFmpeg Go 바인딩
```

## 🗂️ 프로젝트 구조

```
services/cloudRepositoryService/
├── pkg/
│   ├── queue/
│   │   ├── config.go      # Redis 설정
│   │   ├── tasks.go       # 작업 정의
│   │   └── worker.go      # 비디오 처리 워커
│   ├── websocket/
│   │   ├── hub.go         # WebSocket Hub
│   │   ├── client.go      # WebSocket Client
│   │   └── handler.go     # WebSocket Handler
│   └── ffmpeg/
│       └── processor.go   # FFmpeg 유틸리티
├── features/cloudRepository/
│   ├── handler/
│   │   ├── completeUploadHandler.go  # 업로드 완료 핸들러
│   │   └── routes.go                 # 라우트 등록 (업데이트)
│   └── model/entity/
│       └── cloudFile.go   # CloudFile 엔티티 (업데이트)
└── cmd/
    └── main.go            # 메인 서버 (업데이트)
```

## 📝 데이터베이스 스키마 변경

### CloudFile 엔티티에 추가된 필드

```go
type CloudFile struct {
    // ... 기존 필드들
    ProcessingStatus ProcessingStatus `gorm:"size:20;default:'pending';index"`
    ProcessingError  string           `gorm:"size:512"`
}

type ProcessingStatus string

const (
    ProcessingStatusPending   ProcessingStatus = "pending"
    ProcessingStatusProcessing ProcessingStatus = "processing"
    ProcessingStatusCompleted  ProcessingStatus = "completed"
    ProcessingStatusFailed     ProcessingStatus = "failed"
)
```

## 🚀 사용 방법

### 1️⃣ Redis 설치 및 실행

#### macOS
```bash
brew install redis
brew services start redis

# 확인
redis-cli ping
# 응답: PONG
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

#### Docker
```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

### 2️⃣ FFmpeg 설치

#### macOS
```bash
brew install ffmpeg

# 확인
ffmpeg -version
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install ffmpeg

# 확인
ffmpeg -version
```

### 3️⃣ 환경 변수 설정

`.env` 파일에 다음 항목 추가:

```bash
# Redis (for async queue)
REDIS_HOST=localhost:6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key-here

# AWS (기존)
AWS_REGION=ap-northeast-2
CLOUD_REPOSITORY_BUCKET=your-bucket-name

# Local development
IS_LOCAL=true
```

### 4️⃣ 서버 실행

```bash
cd services/cloudRepositoryService
go run cmd/main.go
```

실행 시 다음 로그가 출력되어야 합니다:

```
✅ Database migration completed successfully
✅ Async queue initialized successfully
✅ WebSocket hub initialized successfully
✅ Async worker started successfully
🚀 Server running on port 18080
```

## 📡 API 엔드포인트

### POST /api/v1/files/:id/complete-upload

파일 업로드 완료를 알리고 비디오 처리를 시작합니다.

#### Request

```bash
POST http://localhost:18080/api/v1/files/123/complete-upload
Authorization: Bearer <JWT_TOKEN>
```

#### Response (Video)

```json
{
  "success": true,
  "message": "Video processing started",
  "file": {
    "id": 123,
    "file_name": "video.mp4",
    "processing_status": "pending",
    "s3_key": "files/video.mp4",
    "file_type": "video"
  }
}
```

#### Response (Image)

```json
{
  "success": true,
  "message": "Upload completed",
  "file": {
    "id": 123,
    "file_name": "image.jpg",
    "processing_status": "completed",
    "s3_key": "files/image.jpg",
    "file_type": "image"
  }
}
```

## 🔌 WebSocket 연결

### 연결

```javascript
const token = localStorage.getItem('access_token');
const socket = new WebSocket(`ws://localhost:18080/ws?token=${token}`);

socket.onopen = () => {
  console.log('✅ WebSocket connected');
};

socket.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('📨 Received:', message);

  if (message.type === 'file:processed') {
    const { file_id, status, thumbnail_key, duration, error } = message.data;

    if (status === 'completed') {
      console.log(`✅ File ${file_id} processed successfully`);
      console.log(`   Thumbnail: ${thumbnail_key}`);
      console.log(`   Duration: ${duration}s`);
    } else if (status === 'failed') {
      console.error(`❌ File ${file_id} processing failed: ${error}`);
    }
  }
};

socket.onerror = (error) => {
  console.error('❌ WebSocket error:', error);
};

socket.onclose = () => {
  console.log('❌ WebSocket disconnected');
};
```

### 알림 메시지 형식

#### 성공

```json
{
  "type": "file:processed",
  "data": {
    "file_id": 123,
    "status": "completed",
    "thumbnail_key": "thumbnails/123_video.mp4.jpg",
    "duration": 125.5
  }
}
```

#### 실패

```json
{
  "type": "file:processed",
  "data": {
    "file_id": 123,
    "status": "failed",
    "error": "failed to extract duration: invalid video format"
  }
}
```

## 🧪 테스트

### 1️⃣ Redis 연결 테스트

```bash
redis-cli ping
# 응답: PONG
```

### 2️⃣ 서버 시작 확인

서버 로그에서 다음을 확인:

```
✅ Async queue initialized successfully
✅ WebSocket hub initialized successfully
✅ Async worker started successfully
```

### 3️⃣ WebSocket 연결 테스트

브라우저 콘솔에서:

```javascript
const ws = new WebSocket('ws://localhost:18080/ws?token=YOUR_TOKEN');
ws.onopen = () => console.log('Connected!');
```

### 4️⃣ 전체 플로우 테스트

1. **파일 업로드**
   ```bash
   # 프론트엔드에서 S3에 업로드
   ```

2. **업로드 완료 알림**
   ```bash
   curl -X POST http://localhost:18080/api/v1/files/123/complete-upload \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

3. **WebSocket 알림 확인**
   ```
   브라우저 콘솔에서 file:processed 이벤트 확인
   ```

4. **데이터베이스 확인**
   ```sql
   SELECT id, file_name, processing_status, duration, thumbnail_key
   FROM cloud_files
   WHERE id = 123;
   ```

## 🔧 트러블슈팅

### Redis 연결 실패

```bash
# Redis 실행 확인
redis-cli ping

# Redis 서비스 재시작 (macOS)
brew services restart redis

# Redis 서비스 재시작 (Linux)
sudo systemctl restart redis-server
```

### FFmpeg 오류

```bash
# FFmpeg 설치 확인
ffmpeg -version

# FFmpeg 재설치 (macOS)
brew reinstall ffmpeg
```

### WebSocket 연결 실패

- JWT 토큰이 유효한지 확인
- CORS 설정 확인
- 서버 로그에서 에러 메시지 확인

### 비디오 처리 실패

서버 로그 확인:

```bash
# 처리 실패 로그 확인
grep "Video processing failed" logs/*.log

# 데이터베이스에서 에러 확인
SELECT id, file_name, processing_status, processing_error
FROM cloud_files
WHERE processing_status = 'failed';
```

## 📊 모니터링

### Asynq 대시보드 (선택사항)

Asynq Inspector로 작업 모니터링:

```bash
go install github.com/hibiken/asynq/tools/asynq@latest

asynq dash --redis-addr=localhost:6379
```

브라우저에서 `http://localhost:8080` 접속

### 로그 모니터링

```bash
# 실시간 로그 확인
tail -f logs/app.log

# 비디오 처리 관련 로그만 확인
tail -f logs/app.log | grep "Processing video"
```

## 🚀 배포 시 고려사항

### 1️⃣ Redis 설정

개발: 로컬 Redis
프로덕션: AWS ElastiCache, Redis Cloud 등

```go
// 프로덕션 환경
REDIS_HOST=your-redis-cluster.cache.amazonaws.com:6379
REDIS_PASSWORD=your-secure-password
```

### 2️⃣ 워커 스케일링

- 워커 프로세스를 여러 개 실행하여 병렬 처리
- Redis를 통해 작업 분산
- Kubernetes HPA로 자동 스케일링

### 3️⃣ S3 권한

워커가 S3에 접근할 수 있도록 IAM 권한 설정:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket/*"
      ]
    }
  ]
}
```

### 4️⃣ 에러 알림

Sentry, Slack 등으로 처리 실패 알림 설정

## 📚 참고 자료

- [Asynq Documentation](https://github.com/hibiken/asynq)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [FFmpeg Go](https://github.com/u2takey/ffmpeg-go)
- [Echo Framework](https://echo.labstack.com/)

## ✅ 체크리스트

- [ ] Redis 설치 및 실행 확인
- [ ] FFmpeg 설치 확인
- [ ] Go 의존성 설치 완료
- [ ] 환경 변수 설정 완료
- [ ] 서버 시작 성공
- [ ] WebSocket 연결 테스트 성공
- [ ] 비디오 처리 테스트 성공
- [ ] 데이터베이스 스키마 업데이트 확인

## 🔗 다음 단계

1. 프론트엔드 WebSocket 연동
2. 업로드 플로우 수정 (썸네일 생성 제거)
3. 통합 테스트
4. 성능 최적화
5. 프로덕션 배포

궁금한 점이나 문제가 발생하면 이슈를 생성해주세요!
