# 즐겨찾기 API 프론트엔드 연동 가이드

## 개요
사용자가 자주 사용하는 파일을 즐겨찾기에 추가하고 빠르게 접근할 수 있는 기능입니다.

## API 엔드포인트

### 베이스 URL
```
http://localhost:18080/api/v1
```

### 인증
모든 API는 JWT 인증이 필수입니다.
```
Authorization: Bearer <access_token>
```

---

## 1. 즐겨찾기 추가

### 요청
```http
POST /api/v1/favorites
Content-Type: application/json
Authorization: Bearer <token>

{
  "fileId": 123
}
```

### 응답
**성공 (200 OK)**
```json
{
  "success": true,
  "favoritedAt": "2025-01-26T15:30:00Z"
}
```

**에러**
- `400 Bad Request`: 잘못된 요청 (fileId 누락 또는 유효하지 않음)
  ```json
  {
    "error": "invalid request body"
  }
  ```
- `401 Unauthorized`: 인증 토큰 없음 또는 만료
  ```json
  {
    "error": "unauthorized"
  }
  ```
- `403 Forbidden`: 다른 사용자의 파일
  ```json
  {
    "error": "access denied: you do not own this file"
  }
  ```
- `404 Not Found`: 파일이 존재하지 않음
  ```json
  {
    "error": "file not found"
  }
  ```
- `500 Internal Server Error`: 서버 오류

### 특징
- **Idempotent**: 동일한 파일을 여러 번 추가해도 안전 (중복 생성 안됨)
- 이미 즐겨찾기에 있는 파일을 다시 추가하면 200 응답과 기존 favoritedAt 반환

---

## 2. 즐겨찾기 제거

### 요청
```http
DELETE /api/v1/favorites/:fileId
Authorization: Bearer <token>
```

### 응답
**성공 (204 No Content)**
- 응답 본문 없음

**에러**
- `401 Unauthorized`: 인증 토큰 없음 또는 만료
- `500 Internal Server Error`: 서버 오류

### 특징
- **Idempotent**: 존재하지 않는 즐겨찾기를 제거해도 204 응답 (에러 없음)

---

## 3. 즐겨찾기 목록 조회

### 요청
```http
GET /api/v1/favorites?page=1&size=20&sort=uploadDate&order=desc&q=image&ext=png&tag=work
Authorization: Bearer <token>
```

### 쿼리 파라미터
| 파라미터 | 타입 | 필수 | 기본값 | 설명 |
|---------|------|------|--------|------|
| page | int | 선택 | 1 | 페이지 번호 (1부터 시작) |
| size | int | 선택 | 20 | 페이지당 항목 수 (최소 1, 최대 100) |
| sort | string | 선택 | uploadDate | 정렬 기준 (`uploadDate` 또는 `fileName`) |
| order | string | 선택 | desc | 정렬 순서 (`asc` 또는 `desc`) |
| q | string | 선택 | - | 파일명 검색 (부분 일치) |
| ext | string | 선택 | - | 확장자 필터 (예: `png`, `jpg`, `mp4`) |
| tag | string | 선택 | - | 태그 필터 (정확히 일치) |

### 응답
**성공 (200 OK)**
```json
{
  "data": [
    {
      "id": 123,
      "fileName": "example.png",
      "fileSize": 2048576,
      "contentType": "image/png",
      "extension": "png",
      "s3Key": "users/1/files/abc123.png",
      "downloadUrl": "https://s3.amazonaws.com/bucket/users/1/files/abc123.png?X-Amz-...",
      "thumbnailUrl": "https://s3.amazonaws.com/bucket/users/1/thumbnails/abc123_thumb.jpg?X-Amz-...",
      "uploadedAt": "2025-01-25T10:00:00Z",
      "tags": [
        {
          "id": 1,
          "name": "work"
        },
        {
          "id": 2,
          "name": "important"
        }
      ]
    }
  ],
  "pagination": {
    "total": 45,
    "page": 1,
    "size": 20,
    "totalPages": 3
  }
}
```

**에러**
- `400 Bad Request`: 잘못된 쿼리 파라미터
  ```json
  {
    "error": "invalid query parameters: size must be between 1 and 100"
  }
  ```
- `401 Unauthorized`: 인증 토큰 없음 또는 만료
- `500 Internal Server Error`: 서버 오류

### 파일 정보 필드
| 필드 | 타입 | 설명 |
|------|------|------|
| id | uint | 파일 ID |
| fileName | string | 원본 파일명 |
| fileSize | int64 | 파일 크기 (바이트) |
| contentType | string | MIME 타입 |
| extension | string | 확장자 |
| s3Key | string | S3 저장 경로 |
| downloadUrl | string | 다운로드 URL (1시간 유효) |
| thumbnailUrl | string | 썸네일 URL (동영상만, 1시간 유효) |
| uploadedAt | string | 업로드 일시 (ISO 8601) |
| tags | array | 태그 목록 |

---

## 프론트엔드 통합 가이드

### 1. 즐겨찾기 토글 구현
```javascript
async function toggleFavorite(fileId, isFavorited) {
  const token = localStorage.getItem('accessToken');
  const method = isFavorited ? 'DELETE' : 'POST';
  const url = isFavorited
    ? `http://localhost:18080/api/v1/favorites/${fileId}`
    : 'http://localhost:18080/api/v1/favorites';

  const options = {
    method,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  };

  if (!isFavorited) {
    options.body = JSON.stringify({ fileId });
  }

  try {
    const response = await fetch(url, options);

    if (!response.ok) {
      if (response.status === 401) {
        // 토큰 만료 처리
        redirectToLogin();
        return;
      }
      if (response.status === 403) {
        alert('권한이 없습니다');
        return;
      }
      if (response.status === 404) {
        alert('파일을 찾을 수 없습니다');
        return;
      }
      throw new Error('즐겨찾기 처리 실패');
    }

    // UI 업데이트
    return !isFavorited; // 새로운 상태 반환
  } catch (error) {
    console.error('Error toggling favorite:', error);
    throw error;
  }
}
```

### 2. 즐겨찾기 목록 조회
```javascript
async function fetchFavorites(page = 1, size = 20, filters = {}) {
  const token = localStorage.getItem('accessToken');
  const params = new URLSearchParams({
    page: page.toString(),
    size: size.toString(),
    sort: filters.sort || 'uploadDate',
    order: filters.order || 'desc'
  });

  // 선택적 필터 추가
  if (filters.q) params.append('q', filters.q);
  if (filters.ext) params.append('ext', filters.ext);
  if (filters.tag) params.append('tag', filters.tag);

  try {
    const response = await fetch(
      `http://localhost:18080/api/v1/favorites?${params}`,
      {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      }
    );

    if (!response.ok) {
      if (response.status === 401) {
        redirectToLogin();
        return;
      }
      throw new Error('즐겨찾기 목록 조회 실패');
    }

    const data = await response.json();
    return {
      files: data.data,
      pagination: data.pagination
    };
  } catch (error) {
    console.error('Error fetching favorites:', error);
    throw error;
  }
}
```

### 3. Optimistic UI 업데이트 (권장)
```javascript
async function toggleFavoriteWithOptimisticUpdate(fileId, currentState, onUpdate) {
  // 1. 즉시 UI 업데이트 (낙관적)
  const newState = !currentState;
  onUpdate(fileId, newState);

  try {
    // 2. 서버 요청
    await toggleFavorite(fileId, currentState);
  } catch (error) {
    // 3. 실패시 롤백
    onUpdate(fileId, currentState);
    alert('즐겨찾기 처리에 실패했습니다. 다시 시도해주세요.');
  }
}
```

### 4. 페이지네이션 구현
```javascript
function FavoritesList() {
  const [favorites, setFavorites] = useState([]);
  const [pagination, setPagination] = useState({ page: 1, totalPages: 1 });

  useEffect(() => {
    loadFavorites(pagination.page);
  }, [pagination.page]);

  async function loadFavorites(page) {
    const result = await fetchFavorites(page, 20);
    setFavorites(result.files);
    setPagination(result.pagination);
  }

  return (
    <div>
      {/* 파일 목록 렌더링 */}
      <FileList files={favorites} />

      {/* 페이지네이션 */}
      <Pagination
        current={pagination.page}
        total={pagination.totalPages}
        onChange={(page) => setPagination({...pagination, page})}
      />
    </div>
  );
}
```

---

## 주요 특징 및 주의사항

### 1. Presigned URL 유효시간
- **다운로드 URL**: 1시간 유효
- **업로드 URL**: 12시간 유효 (파일 업로드 API)
- URL 만료시 목록을 다시 조회하여 새 URL 받아야 함

### 2. Idempotent 동작
- **추가**: 이미 즐겨찾기에 있어도 200 OK 응답 (에러 없음)
- **제거**: 없는 항목을 제거해도 204 No Content (에러 없음)
- 네트워크 재시도 안전

### 3. 자동 정리
- 파일 삭제시 즐겨찾기도 자동 삭제 (CASCADE DELETE)
- 별도 처리 불필요

### 4. 성능
- 목록 조회시 최대 100개까지 요청 가능
- 기본 20개 권장
- 인덱스로 최적화됨 (user_id, favorited_at)

### 5. 필터링
- 파일명 검색 (`q`): 부분 일치 (대소문자 구분 없음)
- 확장자 필터 (`ext`): 정확히 일치
- 태그 필터 (`tag`): 정확히 일치 (하나의 태그만 지정 가능)

---

## 에러 처리 권장사항

```javascript
async function handleApiCall(apiFunction) {
  try {
    return await apiFunction();
  } catch (error) {
    // 네트워크 오류
    if (!error.response) {
      alert('네트워크 연결을 확인해주세요');
      return;
    }

    // HTTP 에러
    switch (error.response.status) {
      case 400:
        alert('잘못된 요청입니다');
        break;
      case 401:
        alert('로그인이 필요합니다');
        redirectToLogin();
        break;
      case 403:
        alert('권한이 없습니다');
        break;
      case 404:
        alert('파일을 찾을 수 없습니다');
        break;
      case 500:
        alert('서버 오류가 발생했습니다. 잠시 후 다시 시도해주세요');
        break;
      default:
        alert('오류가 발생했습니다');
    }
  }
}
```

---

## 테스트 시나리오

### 1. 기본 플로우
1. 파일 목록에서 별 아이콘 클릭 → 즐겨찾기 추가
2. 즐겨찾기 탭 이동 → 추가된 파일 확인
3. 별 아이콘 다시 클릭 → 즐겨찾기 제거
4. 즐겨찾기 목록에서 사라진 것 확인

### 2. 필터링 테스트
1. 여러 파일 즐겨찾기 추가 (이미지, 동영상, 다양한 태그)
2. 파일명으로 검색
3. 확장자로 필터링 (예: png만 보기)
4. 태그로 필터링 (예: work 태그만 보기)

### 3. 페이지네이션 테스트
1. 50개 이상의 파일을 즐겨찾기에 추가
2. 페이지 이동 확인
3. size 파라미터 변경 (10, 20, 50)

### 4. 에러 케이스
1. 만료된 토큰으로 요청 → 401 처리 확인
2. 다른 사용자의 파일 즐겨찾기 시도 → 403 처리 확인
3. 존재하지 않는 파일 ID → 404 처리 확인
4. 네트워크 오류 시뮬레이션 → 재시도 로직 확인

---

## 배포 정보

### 데이터베이스 마이그레이션
백엔드 배포 전 다음 마이그레이션 실행 필요:
```bash
migrate -path migrations -database "mysql://user:pass@tcp(host:3306)/dbname" up
```

마이그레이션 파일: `migrations/000004_create_favorites_table.up.sql`

### 환경 변수
기존 환경 변수 사용, 추가 설정 불필요

---

## 문의
백엔드 관련 문의사항은 개발팀에 문의해주세요.

**버전**: 1.0.0
**최종 업데이트**: 2025-01-26
