# Shared Code

공통 인프라 코드 라이브러리 - 모든 서비스에서 재사용

## 구조

### 인프라 계층 (모든 서비스 공통)
- `config/` - 설정 관리 (환경 변수, DB 설정)
- `database/` - DB 연결 및 관리
- `logger/` - 로깅 시스템 (Zap)
- `errors/` - 에러 핸들링 및 커스텀 에러
- `middleware/` - HTTP 미들웨어 (CORS, Rate Limit, Recovery, etc)
- `response/` - 표준 API 응답 포맷

### 공통 유틸리티
- `models/` - 공통 데이터 모델 (BaseModel 등)
- `utils/` - 유틸리티 함수

## 사용 방법

각 서비스에서 shared 패키지를 import:

```go
import (
    "main/shared/config"
    "main/shared/database"
    "main/shared/logger"
    "main/shared/errors"
    "main/shared/middleware"
    "main/shared/response"
)
```

## 서비스별 코드

서비스 고유의 비즈니스 로직은 각 서비스 디렉토리에서 관리:

- `internal/model/` - 서비스별 도메인 모델
- `internal/handler/` - HTTP 핸들러
- `internal/service/` - 비즈니스 로직
- `internal/repository/` - 데이터 접근 계층

## 의존성 관리

- 각 서비스의 `go.mod`에서 `replace` directive로 로컬 shared 참조
- shared 패키지 수정 시 모든 서비스 자동 재배포
