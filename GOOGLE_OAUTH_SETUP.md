# 구글 OAuth 2.0 클라이언트 ID 설정 가이드

구글 로그인을 사용하기 위해 Google Cloud Console에서 OAuth 2.0 클라이언트 ID를 생성하고 설정하는 방법입니다.

## 1. Google Cloud Console에서 클라이언트 ID 생성

### 1.1 프로젝트 생성/선택

1. [Google Cloud Console](https://console.cloud.google.com/)에 접속
2. 상단의 프로젝트 선택 드롭다운 클릭
3. 새 프로젝트 생성 또는 기존 프로젝트 선택

### 1.2 OAuth 동의 화면 설정

1. 왼쪽 메뉴에서 **"API 및 서비스"** > **"OAuth 동의 화면"** 클릭
2. 사용자 유형 선택:
   - **외부**: 일반 사용자도 사용 가능 (테스트 사용자 추가 필요)
   - **내부**: Google Workspace 조직 내부만 사용
3. 앱 정보 입력:
   - **앱 이름**: 예) "Joker Backend"
   - **사용자 지원 이메일**: 본인 이메일
   - **앱 로고**: (선택사항)
   - **앱 도메인**: (선택사항)
4. **저장 후 계속** 클릭

### 1.3 범위(Scopes) 설정

1. **범위 추가 또는 삭제** 화면에서:
   - 기본적으로 `email`, `profile`, `openid`가 포함됨
   - 추가 범위가 필요하면 추가 (일반적으로 기본값으로 충분)
2. **저장 후 계속** 클릭

### 1.4 테스트 사용자 추가 (외부 앱인 경우)

1. **테스트 사용자** 섹션에서 **"사용자 추가"** 클릭
2. 테스트할 구글 이메일 주소 추가
3. **저장 후 계속** 클릭

### 1.5 OAuth 2.0 클라이언트 ID 생성

1. 왼쪽 메뉴에서 **"API 및 서비스"** > **"사용자 인증 정보"** 클릭
2. 상단의 **"+ 사용자 인증 정보 만들기"** 클릭
3. **"OAuth 클라이언트 ID"** 선택
4. 애플리케이션 유형 선택:
   - **웹 애플리케이션**: 백엔드 서버용
   - **iOS**: iOS 앱용
   - **Android**: Android 앱용
   - **기타**: 데스크톱 앱 등

5. **웹 애플리케이션** 선택 시:
   - **이름**: 예) "Joker Backend Web Client"
   - **승인된 JavaScript 원본**: ⚠️ **필수** - 프론트엔드가 실행되는 도메인
     - 로컬 개발: `http://localhost:3000`, `http://localhost:3001` 등
     - 프로덕션: `https://yourdomain.com`, `https://app.yourdomain.com` 등
     - **프로토콜(http/https)과 포트까지 정확히 입력해야 합니다**
   - **승인된 리디렉션 URI**: ⚠️ **필수** - OAuth 콜백을 받을 프론트엔드 URL
     - 로컬 개발: `http://localhost:3000/auth/callback`, `http://localhost:3000/auth/google/callback` 등
     - 프로덕션: `https://yourdomain.com/auth/callback`, `https://yourdomain.com/auth/google/callback` 등
     - **프론트엔드에서 구글 로그인 후 리디렉션되는 URL을 정확히 입력해야 합니다**

6. **만들기** 클릭

**⚠️ 중요**: 
- 여러 도메인을 사용하는 경우 각각 추가해야 합니다
- 로컬 개발과 프로덕션 도메인을 모두 추가하는 것을 권장합니다
- URL 끝에 슬래시(/)가 있으면 없으면 다른 것으로 인식되므로 주의하세요

7. 생성된 클라이언트 ID 확인:
   - 팝업 창에 **클라이언트 ID**와 **클라이언트 보안 비밀번호**가 표시됨
   - **클라이언트 ID**를 복사 (예: `123456789-abcdefghijklmnop.apps.googleusercontent.com`)

## 2. 환경변수 설정

### 2.1 로컬 개발 환경 (.env 파일)

프로젝트 루트에 `.env` 파일을 생성하거나 수정:

```bash
# .env 파일
GOOGLE_CLIENT_ID=123456789-abcdefghijklmnop.apps.googleusercontent.com
```

### 2.2 Docker Compose 사용 시

`.env` 파일에 추가하거나, `docker-compose.yml`에서 직접 설정:

```yaml
# docker-compose.yml의 auth-service 환경변수에 이미 추가됨
GOOGLE_CLIENT_ID: ${GOOGLE_CLIENT_ID:-}
```

`.env` 파일에 설정하면 자동으로 적용됩니다:

```bash
# .env
GOOGLE_CLIENT_ID=your-client-id-here
```

### 2.3 직접 환경변수로 설정

```bash
# Linux/Mac
export GOOGLE_CLIENT_ID=123456789-abcdefghijklmnop.apps.googleusercontent.com

# Windows (PowerShell)
$env:GOOGLE_CLIENT_ID="123456789-abcdefghijklmnop.apps.googleusercontent.com"
```

## 3. 클라이언트 ID 없이 사용하기

클라이언트 ID를 설정하지 않아도 구글 ID 토큰 검증이 가능합니다. 
하지만 프로덕션 환경에서는 보안을 위해 클라이언트 ID를 설정하는 것을 **강력히 권장**합니다.

## 4. 클라이언트 ID 확인 방법

생성한 클라이언트 ID를 다시 확인하려면:

1. [Google Cloud Console](https://console.cloud.google.com/) 접속
2. **"API 및 서비스"** > **"사용자 인증 정보"** 클릭
3. 생성한 OAuth 2.0 클라이언트 ID 클릭
4. **클라이언트 ID** 필드에서 확인

## 5. 주의사항

### 보안
- 클라이언트 ID는 공개되어도 안전하지만, 클라이언트 보안 비밀번호는 절대 공개하지 마세요
- `.env` 파일은 `.gitignore`에 추가되어 있는지 확인하세요
- 프로덕션 환경에서는 환경변수나 시크릿 관리 시스템을 사용하세요

### 테스트
- 외부 앱으로 설정한 경우, 테스트 사용자로 추가된 이메일만 로그인 가능합니다
- 프로덕션 배포 전에 OAuth 동의 화면을 검토 상태로 제출해야 합니다

### 도메인 설정
- **승인된 JavaScript 원본**과 **승인된 리디렉션 URI**는 **반드시 설정해야 합니다**
- 프로덕션 환경에서는 정확한 도메인을 설정해야 합니다
- 로컬 개발 시에는 `http://localhost:포트번호`를 추가하세요
- 여러 환경(로컬, 스테이징, 프로덕션)을 사용하는 경우 모두 추가하세요

**예시 설정:**
```
승인된 JavaScript 원본:
- http://localhost:3000
- http://localhost:3001
- https://yourdomain.com

승인된 리디렉션 URI:
- http://localhost:3000/auth/callback
- http://localhost:3000/auth/google/callback
- https://yourdomain.com/auth/callback
- https://yourdomain.com/auth/google/callback
```

## 6. 문제 해결

### "Invalid Google ID token" 에러
- 클라이언트 ID가 올바른지 확인
- 토큰이 만료되지 않았는지 확인 (일반적으로 1시간)
- OAuth 동의 화면이 올바르게 설정되었는지 확인
- **승인된 JavaScript 원본**에 프론트엔드 도메인이 추가되었는지 확인

### "Access blocked" 에러
- 테스트 사용자로 추가되었는지 확인 (외부 앱인 경우)
- OAuth 동의 화면이 검토 완료되었는지 확인
- **승인된 리디렉션 URI**에 콜백 URL이 정확히 추가되었는지 확인

### "redirect_uri_mismatch" 에러
- **승인된 리디렉션 URI**에 프론트엔드의 실제 콜백 URL이 정확히 추가되었는지 확인
- URL에 슬래시(/)가 있는지 없는지 확인 (예: `/auth/callback` vs `/auth/callback/`)
- 프로토콜(http/https)이 정확한지 확인

### "origin_mismatch" 에러
- **승인된 JavaScript 원본**에 프론트엔드 도메인이 정확히 추가되었는지 확인
- 프로토콜(http/https)과 포트 번호가 정확한지 확인

## 7. 참고 자료

- [Google OAuth 2.0 문서](https://developers.google.com/identity/protocols/oauth2)
- [Google Identity Platform](https://developers.google.com/identity)
- [OAuth 동의 화면 설정](https://support.google.com/cloud/answer/10311615)

