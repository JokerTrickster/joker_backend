---
name: backend-api-helper
description: 백엔드 API 개발을 위한 도우미 스킬. RESTful API 설계, 에러 핸들링, 데이터베이스 쿼리 등을 지원합니다.
---

# Backend API Helper

이 스킬은 백엔드 API 개발 시 Claude가 자동으로 따라야 할 지침과 모범 사례를 제공합니다.

## 사용 목적

- RESTful API 엔드포인트 설계
- 에러 핸들링 및 응답 형식 표준화
- 데이터베이스 쿼리 최적화
- 인증 및 보안 처리

## Workflow

### 1. API 엔드포인트 설계
- RESTful 원칙 준수
- 적절한 HTTP 메서드 사용 (GET, POST, PUT, DELETE, PATCH)
- 명확한 URL 패턴 사용 (`/api/v1/resource`)

### 2. 요청/응답 처리
- 요청 데이터 검증
- 일관된 JSON 응답 형식
- 적절한 HTTP 상태 코드 사용

### 3. 에러 핸들링
- 명확한 에러 메시지 제공
- 적절한 에러 코드 반환
- 로깅 및 모니터링 고려

### 4. 보안
- 인증 토큰 검증
- 입력 데이터 검증 및 sanitization
- SQL Injection 방지

## 예시 응답 형식

### 성공 응답
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

### 에러 응답
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

## 참고 사항

- `references/` 폴더의 문서를 참고하여 구체적인 구현 방법 확인
- `scripts/` 폴더의 코드 템플릿 활용 가능

