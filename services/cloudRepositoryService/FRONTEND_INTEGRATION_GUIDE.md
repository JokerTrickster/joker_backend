# 프론트엔드 통합 가이드 - 비동기 비디오 처리

## 📋 백엔드 작업 완료 내역

### ✅ 구현된 기능

1. **비동기 비디오 처리**
   - S3 업로드 후 백엔드에서 썸네일 생성 및 Duration 추출
   - Redis 기반 백그라운드 작업 큐
   - 재시도 메커니즘 (최대 3번)

2. **WebSocket 실시간 알림**
   - 비디오 처리 완료 시 실시간 알림
   - JWT 기반 인증
   - 자동 재연결 지원

3. **새로운 API 엔드포인트**
   - `POST /api/v1/files/:id/complete-upload` - 업로드 완료 알림

4. **데이터베이스 스키마 변경**
   - `processing_status` 필드 추가 (pending, processing, completed, failed)
   - `processing_error` 필드 추가

---

## 🔄 변경된 업로드 플로우

### 기존 플로우 (프론트엔드에서 모든 처리)
```
1. 파일 선택
2. S3 Presigned URL 요청
3. S3 직접 업로드
4. 프론트엔드에서 썸네일 생성 ❌ (제거됨)
5. 프론트엔드에서 Duration 추출 ❌ (제거됨)
6. 백엔드에 파일 정보 전송
```

### 새로운 플로우 (백엔드에서 비동기 처리)
```
1. 파일 선택
2. S3 Presigned URL 요청
3. S3 직접 업로드
4. 백엔드에 업로드 완료 알림 ✅ (새로 추가)
5. WebSocket으로 처리 완료 대기 ✅ (새로 추가)
6. 처리 완료 시 UI 업데이트 ✅ (새로 추가)
```

---

## 📝 프론트엔드 작업 지시

### 1️⃣ WebSocket 연결 구현 (필수)

#### 위치: `src/services/websocket.ts` (새로 생성)

```typescript
class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 3000;
  private listeners: Map<string, Function[]> = new Map();

  connect(token: string) {
    const wsUrl = `ws://localhost:18080/ws?token=${token}`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('✅ WebSocket connected');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.notifyListeners(message.type, message.data);
    };

    this.ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('❌ WebSocket disconnected');
      this.handleReconnect(token);
    };
  }

  on(eventType: string, callback: Function) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, []);
    }
    this.listeners.get(eventType)!.push(callback);
  }

  off(eventType: string, callback: Function) {
    const callbacks = this.listeners.get(eventType);
    if (callbacks) {
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    }
  }

  private notifyListeners(eventType: string, data: any) {
    const callbacks = this.listeners.get(eventType);
    if (callbacks) {
      callbacks.forEach(callback => callback(data));
    }
  }

  private handleReconnect(token: string) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`🔄 Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

      setTimeout(() => {
        this.connect(token);
      }, this.reconnectDelay);
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const wsService = new WebSocketService();
```

### 2️⃣ 업로드 완료 API 호출 추가 (필수)

#### 위치: `src/api/files.ts` (기존 파일 수정)

```typescript
// 새로운 API 함수 추가
export const completeFileUpload = async (fileId: number) => {
  const response = await fetch(`/api/v1/files/${fileId}/complete-upload`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('access_token')}`
    }
  });

  if (!response.ok) {
    throw new Error('Failed to complete upload');
  }

  return response.json();
};
```

### 3️⃣ 파일 업로드 컴포넌트 수정 (필수)

#### 위치: `src/components/FileUpload.tsx` (기존 파일 수정)

```typescript
import { wsService } from '../services/websocket';
import { completeFileUpload } from '../api/files';

// 컴포넌트 마운트 시 WebSocket 연결
useEffect(() => {
  const token = localStorage.getItem('access_token');
  if (token) {
    wsService.connect(token);
  }

  return () => {
    wsService.disconnect();
  };
}, []);

// 파일 업로드 처리 함수 수정
const handleFileUpload = async (file: File) => {
  try {
    // 1. 기존: Presigned URL 요청
    const { uploadUrl, fileId } = await getPresignedUploadUrl(file);

    // 2. 기존: S3에 직접 업로드
    await uploadToS3(uploadUrl, file);

    // 3. 새로 추가: 업로드 완료 알림
    const result = await completeFileUpload(fileId);

    // 4. 새로 추가: 비디오인 경우 WebSocket으로 처리 완료 대기
    if (result.file.file_type === 'video') {
      // UI에 "처리 중..." 상태 표시
      setProcessingStatus(fileId, 'processing');

      // WebSocket 이벤트 리스너 등록
      const handleProcessed = (data: any) => {
        if (data.file_id === fileId) {
          if (data.status === 'completed') {
            // 처리 완료: UI 업데이트
            updateFileInfo(fileId, {
              thumbnailUrl: data.thumbnail_key,
              duration: data.duration,
              status: 'completed'
            });
            showNotification('✅ 비디오 처리 완료!');
          } else if (data.status === 'failed') {
            // 처리 실패: 에러 표시
            showError(`❌ 처리 실패: ${data.error}`);
            setProcessingStatus(fileId, 'failed');
          }

          // 리스너 제거
          wsService.off('file:processed', handleProcessed);
        }
      };

      wsService.on('file:processed', handleProcessed);
    } else {
      // 이미지는 즉시 완료 처리
      setProcessingStatus(fileId, 'completed');
    }

  } catch (error) {
    console.error('Upload failed:', error);
    showError('업로드 실패');
  }
};
```

### 4️⃣ UI 상태 관리 추가 (필수)

#### 파일 상태 타입 정의

```typescript
interface FileStatus {
  id: number;
  name: string;
  type: 'image' | 'video';
  processingStatus: 'pending' | 'processing' | 'completed' | 'failed';
  thumbnailUrl?: string;
  duration?: number;
  error?: string;
}
```

#### 처리 중 UI 예시

```typescript
const FileCard = ({ file }: { file: FileStatus }) => {
  return (
    <div className="file-card">
      {file.processingStatus === 'processing' && (
        <div className="processing-overlay">
          <Spinner />
          <p>비디오 처리 중...</p>
        </div>
      )}

      {file.processingStatus === 'completed' && (
        <>
          {file.thumbnailUrl && (
            <img src={file.thumbnailUrl} alt={file.name} />
          )}
          {file.duration && (
            <span className="duration">{formatDuration(file.duration)}</span>
          )}
        </>
      )}

      {file.processingStatus === 'failed' && (
        <div className="error">
          <p>❌ 처리 실패</p>
          {file.error && <small>{file.error}</small>}
        </div>
      )}
    </div>
  );
};
```

### 5️⃣ 제거할 코드 (필수)

#### ❌ 프론트엔드 썸네일 생성 코드 제거

```typescript
// 이 함수들 삭제
// - generateVideoThumbnail()
// - extractVideoDuration()
// - createThumbnailFromVideo()
```

이제 백엔드에서 자동으로 처리하므로 **완전히 제거**해야 합니다.

---

## 🧪 테스트 시나리오

### 1. 이미지 업로드 테스트
```
1. 이미지 파일 선택
2. 업로드 진행
3. 즉시 완료 표시 확인 ✅
```

### 2. 비디오 업로드 테스트
```
1. 비디오 파일 선택
2. 업로드 진행
3. "처리 중..." 표시 확인 ✅
4. WebSocket 이벤트 수신 확인 ✅
5. 썸네일 및 Duration 표시 확인 ✅
```

### 3. WebSocket 재연결 테스트
```
1. 비디오 업로드 중 네트워크 끊기
2. 자동 재연결 확인 ✅
3. 처리 완료 알림 수신 확인 ✅
```

### 4. 에러 처리 테스트
```
1. 잘못된 비디오 파일 업로드
2. 에러 메시지 표시 확인 ✅
3. UI에서 재시도 가능 확인 ✅
```

---

## 📊 API 응답 예시

### POST /api/v1/files/:id/complete-upload

#### 비디오 파일 응답
```json
{
  "success": true,
  "message": "Video processing started",
  "file": {
    "id": 123,
    "file_name": "video.mp4",
    "file_type": "video",
    "processing_status": "pending",
    "s3_key": "files/123/video.mp4",
    "created_at": "2025-12-01T10:30:00Z"
  }
}
```

#### 이미지 파일 응답
```json
{
  "success": true,
  "message": "Upload completed",
  "file": {
    "id": 124,
    "file_name": "image.jpg",
    "file_type": "image",
    "processing_status": "completed",
    "s3_key": "files/124/image.jpg",
    "created_at": "2025-12-01T10:31:00Z"
  }
}
```

### WebSocket 이벤트 (file:processed)

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
    "error": "Invalid video format: unsupported codec"
  }
}
```

---

## ⚠️ 중요 사항

### 1. 썸네일 URL 변경
**기존**: 프론트엔드에서 생성한 Blob URL
**변경**: S3에서 받은 thumbnail_key로 presigned URL 요청

```typescript
// 기존 (삭제)
const thumbnailUrl = URL.createObjectURL(thumbnailBlob);

// 새로운 방식
const thumbnailUrl = await getPresignedDownloadUrl(file.thumbnail_key);
```

### 2. Duration 포맷
백엔드에서 **초 단위 float**로 전달합니다.

```typescript
// 예: 125.5초 → "02:05"
const formatDuration = (seconds: number) => {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
};
```

### 3. 브라우저를 닫아도 처리 계속
사용자가 업로드 후 브라우저를 닫아도 백엔드에서 계속 처리합니다.
다음 접속 시 `processing_status`로 완료 여부 확인 가능합니다.

```typescript
// 페이지 로드 시 미완료 파일 확인
const checkPendingFiles = async () => {
  const files = await getMyFiles();
  const pendingVideos = files.filter(f =>
    f.file_type === 'video' &&
    f.processing_status === 'processing'
  );

  // WebSocket 리스너 등록하여 완료 대기
  pendingVideos.forEach(file => {
    wsService.on('file:processed', (data) => {
      if (data.file_id === file.id) {
        updateFileStatus(file.id, data);
      }
    });
  });
};
```

---

## 🔧 환경 변수 설정

### .env 파일 추가

```bash
# WebSocket URL
VITE_WS_URL=ws://localhost:18080/ws

# API Base URL
VITE_API_URL=http://localhost:18080/api/v1
```

### 사용 예시

```typescript
const wsUrl = `${import.meta.env.VITE_WS_URL}?token=${token}`;
```

---

## 📱 사용자 경험 개선

### 1. 진행 상태 표시
```typescript
// 업로드 진행률과 처리 진행률 구분
<ProgressBar
  uploadProgress={uploadProgress}    // S3 업로드 진행률
  processingStatus={processingStatus} // 백엔드 처리 상태
/>
```

### 2. 알림 메시지
```typescript
// 처리 완료 시 토스트 알림
toast.success('✅ 비디오 처리 완료!');

// 처리 실패 시 상세 에러
toast.error(`❌ 처리 실패: ${error.message}`);
```

### 3. 재시도 버튼
```typescript
// 실패한 파일 재처리
const retryProcessing = async (fileId: number) => {
  await completeFileUpload(fileId);
  // WebSocket 리스너 다시 등록
};
```

---

## ✅ 작업 체크리스트

프론트엔드 개발자는 다음 순서로 작업하세요:

- [ ] **1단계**: WebSocket 서비스 구현 (`websocket.ts`)
- [ ] **2단계**: `completeFileUpload` API 함수 추가
- [ ] **3단계**: 파일 업로드 플로우에 WebSocket 통합
- [ ] **4단계**: 처리 중 UI 상태 추가
- [ ] **5단계**: 기존 썸네일/Duration 생성 코드 제거
- [ ] **6단계**: 테스트 시나리오 실행
- [ ] **7단계**: 에러 처리 및 재시도 로직 구현

---

## 🚀 배포 시 주의사항

### WebSocket URL 변경
```typescript
// 개발
ws://localhost:18080/ws

// 프로덕션
wss://api.yourdomain.com/ws  // HTTPS에서는 WSS 사용 필수
```

### CORS 설정 확인
백엔드에서 WebSocket CORS가 허용되어 있어야 합니다. (이미 설정됨)

---

## 📞 문의

- 백엔드 API 관련: `/api/v1/files/:id/complete-upload` 엔드포인트 확인
- WebSocket 연결 문제: 토큰 유효성 및 네트워크 확인
- 처리 실패: `processing_error` 필드에서 상세 에러 확인

**질문이 있으면 백엔드 팀에 문의하세요!** 🙌
