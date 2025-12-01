# 비동기 처리 시스템 구현 완료

## ✅ 구현 완료 항목

### 1. 데이터베이스 스키마 업데이트
- [x] CloudFile 엔티티에 `processing_status` 필드 추가
- [x] CloudFile 엔티티에 `processing_error` 필드 추가
- [x] ProcessingStatus enum 타입 정의 (pending, processing, completed, failed)

### 2. Go 패키지 설치
- [x] `github.com/hibiken/asynq` - Redis 기반 작업 큐
- [x] `github.com/gorilla/websocket` - WebSocket 서버
- [x] `github.com/u2takey/ffmpeg-go` - FFmpeg Go 바인딩

### 3. 비동기 처리 인프라
- [x] Redis 큐 설정 (`pkg/queue/config.go`)
- [x] 비디오 처리 작업 정의 (`pkg/queue/tasks.go`)
- [x] 비디오 처리 워커 구현 (`pkg/queue/worker.go`)

### 4. FFmpeg 유틸리티
- [x] Duration 추출 기능 (`pkg/ffmpeg/processor.go`)
- [x] 썸네일 생성 기능 (중간 프레임)
- [x] S3 업로드/다운로드 기능

### 5. WebSocket 서버
- [x] WebSocket Hub 구현 (`pkg/websocket/hub.go`)
- [x] WebSocket Client 구현 (`pkg/websocket/client.go`)
- [x] WebSocket Handler 구현 (`pkg/websocket/handler.go`)
- [x] JWT 기반 인증

### 6. API 엔드포인트
- [x] `POST /api/v1/files/:id/complete-upload` - 업로드 완료 알림
- [x] `GET /ws?token={jwt}` - WebSocket 연결

### 7. 서버 통합
- [x] main.go에 Redis 큐 초기화 추가
- [x] main.go에 WebSocket Hub 초기화 추가
- [x] main.go에 비디오 처리 워커 시작 추가
- [x] routes.go에 queueClient 파라미터 추가

### 8. 환경 변수 설정
- [x] `.env` 파일에 Redis 설정 추가
- [x] `.env.example` 파일 업데이트

### 9. 문서화
- [x] 구현 가이드 작성 (`ASYNC_PROCESSING_GUIDE.md`)
- [x] README 업데이트 (이 파일)

## 📋 시작하기 전 준비사항

### 1. Redis 설치 및 실행

#### macOS
```bash
brew install redis
brew services start redis
redis-cli ping  # 응답: PONG
```

#### Linux
```bash
sudo apt-get install redis-server
sudo systemctl start redis-server
```

#### Docker
```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

### 2. FFmpeg 설치

#### macOS
```bash
brew install ffmpeg
ffmpeg -version
```

#### Linux
```bash
sudo apt-get install ffmpeg
ffmpeg -version
```

### 3. 환경 변수 확인

`.env` 파일에 다음 설정이 있는지 확인:

```bash
# Redis
REDIS_HOST=localhost:6379
REDIS_PASSWORD=

# AWS
AWS_REGION=ap-northeast-2
CLOUD_REPOSITORY_BUCKET=joker-cloud-repository-dev

# JWT
JWT_SECRET=your-secret-key

# Local
IS_LOCAL=true
```

## 🚀 서버 실행

```bash
cd services/cloudRepositoryService
go run cmd/main.go
```

### 실행 확인

서버 시작 시 다음 로그가 출력되어야 합니다:

```
✅ Database migration completed successfully
✅ Async queue initialized successfully
✅ WebSocket hub initialized successfully
✅ Async worker started successfully
🚀 Server running on port 18080
```

## 📡 사용 방법

### 1. 파일 업로드 완료 알림

프론트엔드에서 S3 업로드 완료 후:

```bash
curl -X POST http://localhost:18080/api/v1/files/123/complete-upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**응답 (비디오 파일):**
```json
{
  "success": true,
  "message": "Video processing started",
  "file": {
    "id": 123,
    "processing_status": "pending"
  }
}
```

**응답 (이미지 파일):**
```json
{
  "success": true,
  "message": "Upload completed",
  "file": {
    "id": 123,
    "processing_status": "completed"
  }
}
```

### 2. WebSocket 연결

```javascript
const token = localStorage.getItem('access_token');
const ws = new WebSocket(`ws://localhost:18080/ws?token=${token}`);

ws.onopen = () => {
  console.log('✅ WebSocket connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  if (message.type === 'file:processed') {
    const { file_id, status, thumbnail_key, duration, error } = message.data;

    if (status === 'completed') {
      console.log(`✅ File ${file_id} processed`);
      console.log(`   Thumbnail: ${thumbnail_key}`);
      console.log(`   Duration: ${duration}s`);
    } else {
      console.error(`❌ Failed: ${error}`);
    }
  }
};
```

## 🔍 처리 플로우

```
1. Frontend → S3 직접 업로드
   ↓
2. Frontend → POST /api/v1/files/:id/complete-upload
   ↓
3. Backend → Redis Queue에 작업 추가
   ↓
4. Worker → 백그라운드 처리 시작
   ├─ S3에서 비디오 다운로드
   ├─ Duration 추출 (FFmpeg)
   ├─ 썸네일 생성 (FFmpeg)
   └─ 썸네일 S3 업로드
   ↓
5. Worker → Database 업데이트
   ↓
6. Worker → WebSocket 알림 전송
   ↓
7. Frontend → 실시간 알림 수신
```

## 🧪 테스트

### Redis 연결 테스트
```bash
redis-cli ping
# 응답: PONG
```

### 서버 빌드 테스트
```bash
cd services/cloudRepositoryService
go build -o bin/server cmd/main.go
```

### WebSocket 연결 테스트
```javascript
// 브라우저 콘솔에서
const ws = new WebSocket('ws://localhost:18080/ws?token=YOUR_TOKEN');
ws.onopen = () => console.log('Connected!');
```

## 📊 모니터링

### Asynq Inspector 설치 (선택)

```bash
go install github.com/hibiken/asynq/tools/asynq@latest
asynq dash --redis-addr=localhost:6379
```

브라우저에서 `http://localhost:8080` 접속하여 작업 모니터링

### 로그 모니터링

```bash
# 비디오 처리 로그
tail -f logs/app.log | grep "Processing video"

# 에러 로그
tail -f logs/app.log | grep "ERROR"
```

## 🐛 트러블슈팅

### Redis 연결 실패
```bash
# Redis 상태 확인
redis-cli ping

# Redis 재시작
brew services restart redis  # macOS
sudo systemctl restart redis-server  # Linux
```

### FFmpeg 오류
```bash
# FFmpeg 설치 확인
ffmpeg -version

# FFmpeg 재설치
brew reinstall ffmpeg  # macOS
```

### WebSocket 연결 실패
- JWT 토큰 유효성 확인
- 서버 로그에서 에러 확인
- 네트워크 탭에서 WebSocket 업그레이드 확인

### 비디오 처리 실패
```sql
-- 실패한 작업 확인
SELECT id, file_name, processing_status, processing_error
FROM cloud_files
WHERE processing_status = 'failed';
```

## 📚 관련 파일

### 핵심 파일
- `pkg/queue/worker.go` - 비디오 처리 워커
- `pkg/websocket/hub.go` - WebSocket 허브
- `pkg/ffmpeg/processor.go` - FFmpeg 유틸리티
- `features/cloudRepository/handler/completeUploadHandler.go` - 업로드 완료 핸들러
- `cmd/main.go` - 서버 진입점

### 설정 파일
- `.env` - 환경 변수
- `go.mod` - Go 의존성

### 문서
- `ASYNC_PROCESSING_GUIDE.md` - 상세 구현 가이드
- `README_ASYNC.md` - 이 파일

## 🚀 배포 고려사항

### 프로덕션 환경

1. **Redis**: AWS ElastiCache 또는 Redis Cloud 사용
2. **FFmpeg**: 컨테이너 이미지에 포함
3. **워커 스케일링**: 여러 워커 인스턴스 실행
4. **모니터링**: Sentry, CloudWatch 등 설정
5. **S3 권한**: IAM Role로 최소 권한 부여

### Docker 배포

Dockerfile에 FFmpeg 추가:

```dockerfile
FROM golang:1.24-alpine

# Install FFmpeg
RUN apk add --no-cache ffmpeg

# ... 기존 내용
```

## ✅ 다음 단계

1. [ ] 프론트엔드 WebSocket 연동
2. [ ] 업로드 플로우에서 썸네일 생성 제거
3. [ ] 통합 테스트 수행
4. [ ] 성능 최적화
5. [ ] 프로덕션 배포

## 📞 문의

문제가 발생하거나 질문이 있으면 이슈를 생성해주세요.
