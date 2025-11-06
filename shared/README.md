# Shared Code

공통 코드 라이브러리 - 모든 서비스에서 재사용 가능

## 구조

- `models/` - 공통 데이터 모델
- `utils/` - 유틸리티 함수
- `middleware/` - 공통 미들웨어

## 사용 방법

각 서비스에서 shared 코드를 import하여 사용:

```go
import "joker_backend/shared/models"
import "joker_backend/shared/utils"
```
