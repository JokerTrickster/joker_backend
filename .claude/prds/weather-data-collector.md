---
name: weather-data-collector
description: 주기적으로 네이버 날씨를 크롤링하여 알람 대상 지역의 날씨 정보를 수집하고 Redis에 캐싱한 후 FCM으로 알람 발송
status: backlog
created: 2025-11-10T11:30:41Z
---

# PRD: Weather Data Collector

## Executive Summary

Weather Data Collector는 사용자가 등록한 날씨 알람을 정확한 시간에 발송하기 위해 주기적으로 날씨 데이터를 수집하고 캐싱하는 백그라운드 시스템입니다. 1분마다 실행되는 고루틴이 DB에서 알람 대상을 조회하고, 알람 시간 1분 전에 네이버 날씨를 크롤링하여 Redis에 저장한 후, FCM을 통해 사용자에게 날씨 알람을 발송합니다.

**핵심 가치:**
- 정확한 시간에 최신 날씨 정보 제공
- Redis 캐싱으로 빠른 데이터 접근
- 자동화된 날씨 알람 발송

## Problem Statement

### 현재 문제
현재 시스템에는 사용자가 날씨 알람을 등록할 수 있지만, 실제로 알람을 발송하는 메커니즘이 없습니다. 날씨 정보를 실시간으로 가져오고 캐싱하며, 정해진 시간에 자동으로 알람을 발송하는 시스템이 필요합니다.

### 왜 지금인가?
- 사용자가 이미 알람을 등록할 수 있는 API가 구현됨
- FCM 토큰이 저장되어 있어 푸시 알림 인프라 준비됨
- 날씨 알람 서비스의 핵심 기능 완성을 위해 필수

### 해결 방안
1분마다 실행되는 고루틴이 다음 작업을 수행:
- DB에서 알람 대상 조회 (현재 시간 + 1분 = 알람 시간)
- 대상 지역의 날씨 데이터 크롤링 (네이버 날씨)
- Redis에 캐싱 (TTL 30분)
- FCM으로 알람 발송
- `last_sent` 업데이트로 중복 방지

## User Stories

### Primary Persona: 날씨 알람 사용자
**프로필:**
- 출근 전 날씨를 확인하고 싶은 직장인
- 특정 시간대의 날씨 정보가 필요한 사용자
- 여러 지역의 날씨를 추적하는 사용자

### User Journey

#### Journey 1: 알람 시간에 날씨 정보 수신
**As a** 날씨 알람 사용자
**I want to** 설정한 시간에 정확한 날씨 정보를 받고 싶다
**So that** 외출 준비를 할 수 있다

**Flow:**
1. 사용자가 "서울시 강남구, 오전 8시" 알람을 등록함
2. 시스템이 매일 07:59에 강남구 날씨를 크롤링
3. 08:00에 사용자에게 FCM 알람 발송
4. 사용자가 푸시 알림으로 날씨 정보 확인

**Acceptance Criteria:**
- [ ] 알람 시간 1분 전에 날씨 데이터 크롤링 완료
- [ ] 알람 시간 정각에 FCM 발송
- [ ] 온도, 습도, 강수확률, 풍속, 날씨 상태 포함
- [ ] 하루에 한 번만 발송 (중복 방지)

#### Journey 2: 여러 지역 동시 알람
**As a** 날씨 알람 사용자
**I want to** 여러 지역의 날씨를 같은 시간에 받고 싶다
**So that** 다양한 지역의 날씨를 비교할 수 있다

**Flow:**
1. 사용자가 "서울 강남구, 부산 해운대구" 두 개 알람 등록 (같은 시간)
2. 시스템이 두 지역의 날씨를 동시 크롤링
3. 각 지역별로 별도 FCM 알람 발송

**Acceptance Criteria:**
- [ ] 여러 지역을 병렬로 크롤링 (고루틴 활용)
- [ ] 각 지역별로 독립적인 Redis 키 저장
- [ ] 크롤링 실패 시 재시도 (최대 3회)

### Pain Points Being Addressed
- ❌ 수동으로 날씨를 확인해야 하는 번거로움
- ❌ 외출 직전 날씨 확인을 깜빡하는 문제
- ❌ 여러 지역 날씨를 일일이 검색해야 하는 불편함
- ✅ 자동화된 시간 기반 날씨 알림
- ✅ 최신 날씨 정보 보장
- ✅ 간편한 푸시 알림 수신

## Requirements

### Functional Requirements

#### FR-1: 주기적 스케줄러 실행
- **Priority:** P0 (Critical)
- **Description:** 1분마다 실행되는 고루틴 스케줄러
- **Details:**
  - Go의 `time.Ticker` 사용
  - 서버 시작 시 즉시 실행
  - 서버 종료 시 graceful shutdown

#### FR-2: 알람 대상 조회
- **Priority:** P0 (Critical)
- **Description:** 현재 시간 + 1분에 해당하는 알람 조회
- **Details:**
  - `user_alarms` 테이블 쿼리
  - `is_enabled = true` AND `deleted_at IS NULL`
  - `alarm_time = CURRENT_TIME + 1분`
  - `last_sent IS NULL OR last_sent < TODAY` (중복 방지)

#### FR-3: 네이버 날씨 크롤링
- **Priority:** P0 (Critical)
- **Description:** 지역별 날씨 정보 크롤링
- **Details:**
  - 네이버 날씨 페이지 HTML 파싱
  - 추출 정보: 온도, 습도, 강수확률, 풍속, 날씨 상태
  - User-Agent 헤더 설정 (봇 차단 방지)
  - 타임아웃 설정 (10초)

#### FR-4: Redis 캐싱
- **Priority:** P0 (Critical)
- **Description:** 날씨 데이터를 Redis Hash로 저장
- **Details:**
  - 키 형식: `weather:서울시:강남구`
  - 데이터 형식: Hash (field-value)
    - `temperature`: 온도 (°C)
    - `humidity`: 습도 (%)
    - `precipitation`: 강수확률 (%)
    - `wind_speed`: 풍속 (m/s)
    - `condition`: 날씨 상태 (맑음, 흐림, 비 등)
    - `updated_at`: 업데이트 시간 (ISO 8601)
  - TTL: 30분

#### FR-5: FCM 알람 발송
- **Priority:** P0 (Critical)
- **Description:** 사용자에게 날씨 알람 푸시 발송
- **Details:**
  - `weather_service_tokens` 테이블에서 FCM 토큰 조회
  - 메시지 템플릿:
    ```
    [지역] 날씨 알람
    🌡️ 온도: 15°C
    💧 습도: 60%
    🌧️ 강수확률: 30%
    💨 풍속: 2.5m/s
    ☀️ 맑음
    ```
  - FCM payload 구성 및 발송
  - 발송 실패 시 로깅 (재시도 없음)

#### FR-6: last_sent 업데이트
- **Priority:** P0 (Critical)
- **Description:** 중복 발송 방지를 위한 타임스탬프 업데이트
- **Details:**
  - FCM 발송 성공 후 `user_alarms.last_sent` 업데이트
  - 현재 날짜 시간 저장
  - 다음 날 같은 시간에 다시 발송 가능

#### FR-7: 에러 처리 및 재시도
- **Priority:** P1 (High)
- **Description:** 크롤링 실패 시 재시도 로직
- **Details:**
  - 최대 3회 재시도
  - 지수 백오프 (1초, 2초, 4초)
  - 3회 실패 시 로그 기록 후 스킵

### Non-Functional Requirements

#### NFR-1: Performance
- 스케줄러는 1분 이내에 모든 작업 완료
- 크롤링 응답 시간: 평균 3초 이하
- Redis 저장 시간: 100ms 이하
- 동시 크롤링 가능: 최대 50개 지역

#### NFR-2: Reliability
- 스케줄러 가동률: 99.9% 이상
- 크롤링 성공률: 95% 이상
- FCM 발송 성공률: 98% 이상

#### NFR-3: Scalability
- 단일 인스턴스에서 1,000명 사용자 지원
- 알람 수 증가 시 고루틴 풀 크기 조정 가능
- Redis 메모리 효율적 사용 (TTL로 자동 정리)

#### NFR-4: Security
- 크롤링 시 적절한 User-Agent 사용
- Rate limiting (네이버 서버 부하 방지)
- Redis 접근 인증

#### NFR-5: Observability
- 구조화된 로깅 (JSON 형식)
- 주요 이벤트 로깅:
  - 스케줄러 시작/종료
  - 알람 대상 조회 (건수)
  - 크롤링 성공/실패 (지역, 소요시간)
  - FCM 발송 성공/실패 (user_id, alarm_id)
  - 에러 로깅 (스택 트레이스 포함)

## Success Criteria

### Key Metrics

1. **알람 정확도**
   - Target: 99% 이상의 알람이 설정 시간 ±1분 내 발송
   - Measurement: FCM 발송 시간과 alarm_time 비교

2. **크롤링 성공률**
   - Target: 95% 이상
   - Measurement: 성공 크롤링 수 / 총 시도 수

3. **시스템 가동 시간**
   - Target: 99.9% uptime
   - Measurement: 스케줄러 실행 시간 / 전체 시간

4. **응답 속도**
   - Target: 크롤링부터 FCM 발송까지 10초 이내
   - Measurement: 각 단계별 소요 시간 측정

### Measurable Outcomes

- 사용자 1,000명 기준 일일 알람 발송 성공률 98% 이상
- Redis 캐시 히트율 80% 이상 (같은 지역 중복 조회 시)
- 크롤링 에러율 5% 이하
- 평균 알람 발송 지연 시간 5초 이하

## Technical Design

### Architecture

```
┌─────────────────┐
│  weatherService │
│   (Main Server) │
└────────┬────────┘
         │
         │ Start Goroutine
         ▼
┌─────────────────────────────┐
│  Weather Data Collector     │
│  (Background Goroutine)     │
│                             │
│  ┌──────────────────────┐  │
│  │  1분 Ticker          │  │
│  └──────────────────────┘  │
│           │                 │
│           ▼                 │
│  ┌──────────────────────┐  │
│  │  Query user_alarms   │  │
│  │  (alarm_time = now+1)│  │
│  └──────────────────────┘  │
│           │                 │
│           ▼                 │
│  ┌──────────────────────┐  │
│  │ Crawl Naver Weather  │  │
│  │  (Parallel, Retry)   │  │
│  └──────────────────────┘  │
│           │                 │
│           ▼                 │
│  ┌──────────────────────┐  │
│  │  Cache to Redis      │  │
│  │  (Hash, TTL 30min)   │  │
│  └──────────────────────┘  │
│           │                 │
│           ▼                 │
│  ┌──────────────────────┐  │
│  │  Send FCM Alarm      │  │
│  └──────────────────────┘  │
│           │                 │
│           ▼                 │
│  ┌──────────────────────┐  │
│  │  Update last_sent    │  │
│  └──────────────────────┘  │
└─────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  MySQL (3306)   │
│  - user_alarms  │
│  - tokens       │
└─────────────────┘

         │
         ▼
┌─────────────────┐
│  Redis (6379)   │
│  - weather:*    │
└─────────────────┘

         │
         ▼
┌─────────────────┐
│  FCM Server     │
└─────────────────┘
```

### Data Models

#### Redis Weather Data (Hash)
```
Key: weather:서울시:강남구
Fields:
  temperature: "15"
  humidity: "60"
  precipitation: "30"
  wind_speed: "2.5"
  condition: "맑음"
  updated_at: "2025-11-10T08:59:00Z"
TTL: 1800 seconds (30분)
```

#### Alarm Query Result
```go
type AlarmTarget struct {
    AlarmID  int
    UserID   int
    Region   string
    FCMToken string
    DeviceID string
}
```

#### Weather Data
```go
type WeatherData struct {
    Temperature   float64 // °C
    Humidity      int     // %
    Precipitation int     // %
    WindSpeed     float64 // m/s
    Condition     string  // 맑음, 흐림, 비, 눈 등
    UpdatedAt     time.Time
}
```

### Components

#### 1. Scheduler
- `scheduler/scheduler.go`
- 1분마다 Tick
- Graceful shutdown 지원

#### 2. Alarm Query Service
- `service/alarm_query.go`
- DB에서 알람 대상 조회
- 중복 방지 로직 포함

#### 3. Weather Crawler
- `crawler/naver_weather.go`
- HTML 파싱 (goquery 사용)
- 재시도 로직
- User-Agent 설정

#### 4. Redis Cache
- `cache/weather_cache.go`
- Hash 저장/조회
- TTL 관리

#### 5. FCM Sender
- `notification/fcm_sender.go`
- FCM 메시지 구성
- 발송 로직
- 에러 처리

#### 6. Database Repository
- `repository/alarm_repository.go`
- 알람 조회
- last_sent 업데이트

### Configuration

```yaml
weather_collector:
  interval: "1m"
  crawler:
    timeout: "10s"
    max_retries: 3
    retry_delays: ["1s", "2s", "4s"]
    user_agent: "Mozilla/5.0 (compatible; WeatherBot/1.0)"
  redis:
    ttl: "30m"
    key_prefix: "weather:"
  fcm:
    timeout: "10s"
  concurrency:
    max_goroutines: 50
```

## Constraints & Assumptions

### Constraints
- 네이버 날씨 크롤링은 네이버의 robots.txt 및 이용약관을 준수해야 함
- 크롤링 빈도 제한 (rate limiting 필요)
- 단일 인스턴스만 실행 (중복 실행 방지 메커니즘 없음)
- Redis가 반드시 사용 가능해야 함 (fallback 없음)

### Assumptions
- 네이버 날씨 HTML 구조가 크게 변하지 않음
- FCM 서버가 안정적으로 동작함
- MySQL 연결이 안정적임
- 사용자 알람 수가 급격히 증가하지 않음 (1,000개 이하)
- 한국 시간대(KST) 기준으로 동작

### Technical Limitations
- Go 기반 구현
- 네이버 날씨 페이지 변경 시 크롤러 업데이트 필요
- HTML 파싱 오류 가능성 존재
- 지역명 매핑 필요 (DB 지역명 → 네이버 URL)

## Out of Scope

명시적으로 이번 구현에서 **제외**되는 항목:

1. ❌ **미세먼지 정보 수집** - 추후 별도 구현
2. ❌ **날씨 예보 데이터** - 현재 날씨만 제공
3. ❌ **다중 인스턴스 지원** - 단일 인스턴스만 고려
4. ❌ **웹 대시보드** - 모니터링은 로그로만
5. ❌ **사용자 알람 설정 변경** - 기존 API 사용
6. ❌ **날씨 데이터 영속화** - Redis 캐시만 사용
7. ❌ **다른 날씨 소스** - 네이버 날씨만 지원
8. ❌ **알람 발송 이력 저장** - last_sent만 업데이트
9. ❌ **실시간 날씨 조회 API** - 크롤링 결과만 사용
10. ❌ **국제화** - 한국어만 지원

## Dependencies

### External Dependencies
- **네이버 날씨**: 크롤링 대상, 페이지 구조 변경 시 영향
- **FCM (Firebase Cloud Messaging)**: 푸시 알림 발송
- **Redis**: 날씨 데이터 캐싱

### Internal Dependencies
- **user_alarms 테이블**: 알람 대상 조회
- **weather_service_tokens 테이블**: FCM 토큰 조회
- **기존 JWT 인증**: 없음 (백그라운드 작업)

### Technical Dependencies
- **Go libraries**:
  - `github.com/PuerkitoBio/goquery`: HTML 파싱
  - `github.com/go-redis/redis/v8`: Redis 클라이언트
  - `firebase.google.com/go/v4`: FCM SDK
  - `gorm.io/gorm`: Database ORM
  - `github.com/robfig/cron/v3`: (optional) 더 정교한 스케줄링

### Team Dependencies
- **DevOps**: Redis 설치 및 설정
- **Infrastructure**: FCM 프로젝트 설정 및 서버 키 발급
- **Backend**: 기존 알람 등록 API와 통합 테스트

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1)
- [ ] Scheduler 구현 (1분 Ticker)
- [ ] Alarm Query Service 구현
- [ ] Redis 캐시 레이어 구현
- [ ] 기본 로깅 설정

### Phase 2: Weather Crawling (Week 2)
- [ ] 네이버 날씨 크롤러 구현
- [ ] HTML 파싱 로직
- [ ] 재시도 로직
- [ ] 지역명 매핑 테이블/함수

### Phase 3: FCM Integration (Week 3)
- [ ] FCM SDK 통합
- [ ] 메시지 템플릿 구현
- [ ] FCM 토큰 조회 로직
- [ ] 발송 에러 처리

### Phase 4: Integration & Testing (Week 4)
- [ ] 전체 플로우 통합
- [ ] last_sent 업데이트 로직
- [ ] 중복 방지 테스트
- [ ] 부하 테스트 (1,000개 알람)
- [ ] 에러 케이스 테스트

### Phase 5: Deployment & Monitoring (Week 5)
- [ ] Production 배포
- [ ] 로그 모니터링 설정
- [ ] 성능 메트릭 수집
- [ ] 문서화

## Testing Strategy

### Unit Tests
- [ ] Scheduler 로직
- [ ] Weather data 파싱
- [ ] Redis 캐시 저장/조회
- [ ] FCM 메시지 구성
- [ ] 재시도 로직

### Integration Tests
- [ ] DB 쿼리 → 크롤링 → Redis 저장
- [ ] Redis 조회 → FCM 발송
- [ ] last_sent 업데이트

### E2E Tests
- [ ] 알람 등록 → 스케줄러 실행 → FCM 수신
- [ ] 여러 지역 동시 알람
- [ ] 중복 발송 방지

### Performance Tests
- [ ] 1,000개 알람 동시 처리
- [ ] 크롤링 응답 시간 측정
- [ ] Redis 성능 측정

## Risks & Mitigation

### Risk 1: 네이버 페이지 구조 변경
- **Impact:** High
- **Probability:** Medium
- **Mitigation:**
  - HTML 파싱 실패 감지 및 알림
  - 구조 변경 시 빠른 대응 가능하도록 크롤러 모듈화
  - 백업 데이터 소스 고려 (추후)

### Risk 2: 크롤링 차단
- **Impact:** High
- **Probability:** Low
- **Mitigation:**
  - 적절한 User-Agent 사용
  - Rate limiting 구현
  - IP 차단 시 공식 API로 전환 고려

### Risk 3: Redis 장애
- **Impact:** High
- **Probability:** Low
- **Mitigation:**
  - Redis 연결 재시도 로직
  - Redis 장애 시 크롤링 스킵 (알람 미발송)
  - Redis 모니터링 및 알림

### Risk 4: FCM 발송 실패
- **Impact:** Medium
- **Probability:** Low
- **Mitigation:**
  - 로깅으로 실패 추적
  - 사용자 피드백 채널 (추후)

### Risk 5: 알람 수 급증
- **Impact:** Medium
- **Probability:** Medium
- **Mitigation:**
  - 고루틴 풀 크기 제한
  - 크롤링 우선순위 설정
  - 캐시 활용으로 중복 크롤링 방지

## Monitoring & Alerts

### Metrics to Track
1. **Scheduler Health**
   - 실행 횟수 (분당 1회)
   - 실행 소요 시간
   - 스케줄러 중단 여부

2. **Crawling Metrics**
   - 크롤링 성공률
   - 크롤링 응답 시간 (p50, p95, p99)
   - 재시도 횟수
   - 실패한 지역 목록

3. **FCM Metrics**
   - 발송 성공률
   - 발송 소요 시간
   - 실패 원인별 분류

4. **Redis Metrics**
   - 캐시 히트율
   - 저장/조회 시간
   - 메모리 사용량

### Log Levels
- **INFO**: 정상 동작 (스케줄러 실행, 알람 발송 성공)
- **WARN**: 재시도 발생, 일부 실패
- **ERROR**: 크롤링 실패, FCM 발송 실패, DB 에러
- **FATAL**: 스케줄러 중단, Redis 연결 실패

### Sample Log Format
```json
{
  "timestamp": "2025-11-10T08:59:00Z",
  "level": "INFO",
  "component": "weather-collector",
  "action": "crawl_weather",
  "region": "서울시 강남구",
  "duration_ms": 2500,
  "success": true,
  "retry_count": 0
}
```

## Documentation Requirements

- [ ] README: 시스템 개요 및 실행 방법
- [ ] API 문서: (없음, 백그라운드 작업)
- [ ] 운영 가이드: Redis 설정, FCM 설정, 로그 확인
- [ ] 트러블슈팅 가이드: 일반적인 에러 케이스
- [ ] 네이버 크롤링 구조 문서: HTML 구조 변경 시 참고용

## Future Enhancements

향후 고려 사항 (현재 범위 밖):

1. **미세먼지 정보 추가**
   - 에어코리아 API 통합
   - 별도 크롤러 구현

2. **날씨 예보 데이터**
   - 시간대별 예보
   - 주간 예보

3. **다중 데이터 소스**
   - OpenWeatherMap API
   - 기상청 API
   - Fallback 메커니즘

4. **알람 커스터마이징**
   - 사용자 정의 메시지 템플릿
   - 조건부 알람 (비 올 때만, 온도 특정 범위 등)

5. **웹 대시보드**
   - 크롤링 상태 모니터링
   - 알람 발송 이력
   - 메트릭 시각화

6. **다중 인스턴스 지원**
   - Redis 기반 분산 락
   - 인스턴스 간 작업 분배

7. **알람 발송 이력**
   - 별도 테이블에 저장
   - 사용자 발송 이력 조회 API

## Appendix

### A. Region Mapping Example
네이버 날씨 URL 구조 예시:
```
서울시 강남구 → https://weather.naver.com/today/11680
부산시 해운대구 → https://weather.naver.com/today/26530
```

지역 코드 매핑 필요 (CSV 또는 DB 테이블)

### B. Naver Weather HTML Structure (예상)
```html
<div class="temperature_text">
  <strong>15°</strong>
</div>
<dl class="summary_list">
  <dd>60%</dd> <!-- 습도 -->
  <dd>30%</dd> <!-- 강수확률 -->
  <dd>2.5m/s</dd> <!-- 풍속 -->
</dl>
```

### C. FCM Payload Example
```json
{
  "token": "fcm-token-here",
  "notification": {
    "title": "서울시 강남구 날씨 알람",
    "body": "🌡️ 15°C | 💧 60% | 🌧️ 30% | 💨 2.5m/s | ☀️ 맑음"
  },
  "data": {
    "type": "weather_alarm",
    "region": "서울시 강남구",
    "temperature": "15",
    "condition": "맑음"
  }
}
```

### D. References
- [Naver Weather](https://weather.naver.com/)
- [Firebase Cloud Messaging](https://firebase.google.com/docs/cloud-messaging)
- [goquery Documentation](https://github.com/PuerkitoBio/goquery)
- [go-redis Documentation](https://github.com/go-redis/redis)
