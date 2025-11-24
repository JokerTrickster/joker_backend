# 구글 로그인 테스트 가이드

구글 로그인 API를 테스트하는 방법입니다.

## 1. 서버 실행

### Docker Compose로 실행
```bash
docker-compose up auth-service
```

### 로컬에서 직접 실행
```bash
cd services/authService
go run cmd/main.go
```

서버가 정상적으로 실행되면 `http://localhost:8080`에서 접근 가능합니다.

## 2. 구글 ID 토큰 얻기

### 방법 1: 프론트엔드에서 구글 로그인 구현

프론트엔드에서 Google Sign-In을 구현하여 ID 토큰을 받습니다.

**React 예시:**
```javascript
import { GoogleLogin } from '@react-oauth/google';

function App() {
  const handleGoogleSuccess = async (credentialResponse) => {
    const idToken = credentialResponse.credential;
    
    // 백엔드로 ID 토큰 전송
    const response = await fetch('http://localhost:8080/v0.1/auth/google/signin', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        idToken: idToken
      })
    });
    
    const data = await response.json();
    console.log('Access Token:', data.accessToken);
    console.log('Refresh Token:', data.refreshToken);
  };

  return (
    <GoogleLogin
      onSuccess={handleGoogleSuccess}
      onError={() => {
        console.log('Login Failed');
      }}
    />
  );
}
```

### 방법 2: Google OAuth Playground 사용 (테스트용)

1. [Google OAuth 2.0 Playground](https://developers.google.com/oauthplayground/) 접속
2. 왼쪽에서 "Google OAuth2 API v2" 선택
3. `https://www.googleapis.com/auth/userinfo.email` 체크
4. "Authorize APIs" 클릭
5. 구글 계정으로 로그인
6. "Exchange authorization code for tokens" 클릭
7. 받은 `id_token` 값을 복사

### 방법 3: curl로 테스트

구글 ID 토큰을 받은 후:

```bash
curl -X POST http://localhost:8080/v0.1/auth/google/signin \
  -H "Content-Type: application/json" \
  -d '{
    "idToken": "여기에_구글_ID_토큰_입력"
  }'
```

**성공 응답 예시:**
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**에러 응답 예시:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid Google ID token"
  }
}
```

## 3. API 엔드포인트

- **URL**: `POST /v0.1/auth/google/signin`
- **Content-Type**: `application/json`
- **Request Body**:
  ```json
  {
    "idToken": "구글_ID_토큰"
  }
  ```
- **Response**:
  ```json
  {
    "accessToken": "JWT_액세스_토큰",
    "refreshToken": "JWT_리프레시_토큰"
  }
  ```

## 4. 확인 사항

### 환경변수 확인
```bash
# Docker Compose 사용 시
docker-compose exec auth-service env | grep GOOGLE_CLIENT_ID

# 로컬 실행 시
echo $GOOGLE_CLIENT_ID
```

### 로그 확인
서버 로그에서 다음을 확인:
- 환경변수가 제대로 로드되었는지
- 구글 토큰 검증이 성공했는지
- 유저 생성/조회가 정상적으로 이루어졌는지

## 5. 문제 해결

### "Invalid Google ID token" 에러
1. ID 토큰이 만료되지 않았는지 확인 (일반적으로 1시간 유효)
2. 클라이언트 ID가 올바른지 확인
3. 구글 클라우드 콘솔에서 승인된 JavaScript 원본과 리디렉션 URI가 올바르게 설정되었는지 확인

### "Email not found in token" 에러
1. 구글 OAuth 동의 화면에서 `email` 스코프가 포함되었는지 확인
2. 사용자가 이메일 공개를 허용했는지 확인

### 데이터베이스 에러
1. MySQL이 정상적으로 실행 중인지 확인
2. users 테이블이 존재하는지 확인
3. 데이터베이스 연결 정보가 올바른지 확인

## 6. Swagger 문서 확인

서버 실행 후 Swagger 문서에서 API를 확인할 수 있습니다:
- Swagger UI: `http://localhost:8080/swagger/index.html` (설정된 경우)
- 또는 API 문서 경로 확인

## 7. 다음 단계

1. 프론트엔드에서 구글 로그인 버튼 구현
2. 받은 JWT 토큰을 로컬 스토리지나 쿠키에 저장
3. 이후 API 요청 시 `Authorization: Bearer {accessToken}` 헤더로 인증

