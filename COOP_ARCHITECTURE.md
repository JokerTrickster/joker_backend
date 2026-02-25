# 협동 타워 디펜스 아키텍처 설계

## 개요
2인 협동 타워 디펜스를 위한 백엔드 아키텍처 설계 문서

## 현재 상태

### ✅ 구현됨
- REST API 방 매칭 시스템
- Unity WebSocket 클라이언트
- 방 생성/참가/준비 로직

### ❌ 미구현
- WebSocket 서버
- 실시간 게임 동기화
- 이벤트 브로드캐스팅

---

## 권장 아키텍처: Relaxed Hybrid

### 왜 이 방식인가?

**타워 디펜스 특징:**
- PvE (협동)
- 라운드 기반
- 결과 중심
- 치트가 큰 문제 아님 (친구끼리 플레이)

**선택 이유:**
- ✅ 서버 부하 낮음
- ✅ 구현 간단
- ✅ 지연 시간 낮음
- ✅ 확장 가능

---

## 시스템 플로우

### Phase 1: 매칭 (REST API)
```
Player 1                    Server                    Player 2
   │                          │                          │
   ├─ POST /coop/rooms ──────►│                          │
   │◄─ room_id: 123 ──────────┤                          │
   │                          │                          │
   │                          │◄─ POST /rooms/join ──────┤
   │                          ├─ room_id: 123 ──────────►│
   │                          │                          │
   ├─ POST /rooms/:id/ready ─►│                          │
   │                          │◄─ POST /rooms/:id/ready ─┤
   │                          │                          │
   │◄─ both_ready ────────────┼─ both_ready ────────────►│
```

### Phase 2: 게임 시작 (WebSocket)
```
Player 1                    Server                    Player 2
   │                          │                          │
   ├─ ws://server/game/123 ──►│                          │
   │◄─ connected ─────────────┤                          │
   │                          │◄─ ws://server/game/123 ──┤
   │                          ├─ connected ─────────────►│
   │                          │                          │
   │◄─ game_start ────────────┼─ game_start ────────────►│
   │   {difficulty, seed}     │   {difficulty, seed}     │
```

### Phase 3: 게임 진행 (경량 동기화)
```
각 클라이언트가 독립적으로 실행
서버는 주요 이벤트만 중계

Player 1                    Server                    Player 2
   │                          │                          │
   ├─ round_complete: 1 ─────►│                          │
   │                          ├─ round_complete: 1 ─────►│
   │                          │                          │
   │◄─ opponent_hp: 95 ───────┤◄─ hp_update: 95 ─────────┤
   │                          │                          │
   ├─ monster_kill: 5 ────────►│                          │
   │                          ├─ opponent_kills: 5 ──────►│
```

**동기화 항목 (최소):**
- ✅ 라운드 진행 (round_start, round_complete)
- ✅ HP 변화 (player_hp)
- ✅ 주요 이벤트 (보스 처치, 게임 오버)

**동기화 안 하는 것:**
- ❌ 개별 유닛 위치
- ❌ 몬스터 위치
- ❌ 스킬 사용
→ 각 클라이언트가 같은 seed로 독립 실행

### Phase 4: 결과 전송 (REST API)
```
Player 1                    Server                    Player 2
   │                          │                          │
   ├─ POST /game/result ─────►│                          │
   │   {rounds: 30, hp: 5}    │                          │
   │                          │◄─ POST /game/result ─────┤
   │                          │   {rounds: 30, hp: 8}    │
   │                          │                          │
   │◄─ ranking_update ────────┼─ ranking_update ─────────►│
```

---

## WebSocket 메시지 프로토콜

### Client → Server

#### 1. 라운드 완료
```json
{
  "type": "round_complete",
  "round": 5,
  "hp": 87,
  "gold": 1200,
  "kills": 45
}
```

#### 2. HP 변화
```json
{
  "type": "hp_update",
  "hp": 75,
  "timestamp": 1234567890
}
```

#### 3. 게임 오버
```json
{
  "type": "game_over",
  "reason": "hp_zero",
  "final_round": 18,
  "final_hp": 0
}
```

### Server → Client

#### 1. 게임 시작
```json
{
  "type": "game_start",
  "difficulty": "hard",
  "seed": 12345,
  "player1_id": 1,
  "player2_id": 2
}
```

#### 2. 상대방 상태
```json
{
  "type": "opponent_state",
  "round": 5,
  "hp": 90,
  "gold": 1100,
  "kills": 42
}
```

#### 3. 라운드 동기화
```json
{
  "type": "round_sync",
  "round": 6,
  "both_ready": true
}
```

---

## 백엔드 구현 필요 사항

### 1. WebSocket 서버 (Go + gorilla/websocket)

```go
// services/lottoDefenseService/features/towerDefense/websocket/server.go

type GameSession struct {
    RoomID    uint
    Player1   *websocket.Conn
    Player2   *websocket.Conn
    State     GameState
    Seed      int64
}

type WebSocketServer struct {
    sessions map[uint]*GameSession
    mu       sync.RWMutex
}

func (s *WebSocketServer) HandleConnection(c echo.Context) error {
    ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
    if err != nil {
        return err
    }
    
    roomID := c.QueryParam("room_id")
    userID := getUserIDFromContext(c)
    
    // 세션 찾기 또는 생성
    session := s.getOrCreateSession(roomID)
    
    // 플레이어 등록
    if session.Player1 == nil {
        session.Player1 = ws
    } else {
        session.Player2 = ws
    }
    
    // 양쪽 다 연결되면 게임 시작
    if session.Player1 != nil && session.Player2 != nil {
        s.startGame(session)
    }
    
    // 메시지 루프
    go s.handleMessages(ws, session, userID)
    
    return nil
}
```

### 2. 메시지 브로드캐스팅

```go
func (s *WebSocketServer) broadcastToOpponent(
    session *GameSession, 
    sender *websocket.Conn, 
    message []byte,
) {
    var target *websocket.Conn
    
    if sender == session.Player1 {
        target = session.Player2
    } else {
        target = session.Player1
    }
    
    if target != nil {
        target.WriteMessage(websocket.TextMessage, message)
    }
}
```

### 3. 라우트 추가

```go
// handler/routes.go

func RegisterTowerDefenseRoutes(e *echo.Echo, usecases *Usecases) {
    // ... existing REST routes
    
    // WebSocket route
    wsServer := websocket.NewWebSocketServer()
    e.GET("/ws/game", wsServer.HandleConnection, middleware.JWT())
}
```

---

## 대안: 더 간단한 방법 (REST Only)

WebSocket 없이 REST API polling으로도 가능:

### 폴링 방식
```
Player 1                    Server                    Player 2
   │                          │                          │
   ├─ POST /game/state ──────►│                          │
   │   {round: 5, hp: 87}     │                          │
   │                          │                          │
   ├─ GET /room/:id/state ───►│                          │
   │◄─ opponent: {hp: 90} ────┤                          │
   │                          │◄─ POST /game/state ──────┤
   │                          │   {round: 5, hp: 90}     │
   │                          │                          │
   (3초마다 반복)              │                          │
```

**장점:**
- ✅ 구현 매우 간단
- ✅ WebSocket 서버 불필요
- ✅ 방화벽/NAT 문제 없음

**단점:**
- ❌ 3초 지연
- ❌ 서버 요청 많음
- ❌ 실시간성 떨어짐

**타워 디펜스에는 충분히 사용 가능!**

---

## 추천 구현 순서

### Phase 1: REST Only (가장 간단)
1. ✅ 기존 방 매칭 API 사용
2. ✅ `/room/:id/state` 엔드포인트 추가
3. ✅ Unity에서 3초마다 폴링
4. ✅ 결과만 저장

**예상 작업 시간: 2-4시간**

### Phase 2: WebSocket 추가 (더 나은 경험)
1. WebSocket 서버 구현
2. 라운드 동기화
3. 실시간 상태 공유

**예상 작업 시간: 1-2일**

### Phase 3: Advanced (선택)
1. 서버 검증 로직
2. 치트 감지
3. 리플레이 시스템

**예상 작업 시간: 3-5일**

---

## 결론

**현재 상황:**
- Unity에 WebSocket 클라이언트는 구현됨
- 백엔드에 WebSocket 서버 없음

**권장 사항:**

### 빠른 프로토타입: REST Polling
```go
// 추가 엔드포인트 1개만
GET /coop/rooms/:id/opponent-state
→ {round: 5, hp: 90, gold: 1200}
```

### 정식 버전: WebSocket
- Go + gorilla/websocket
- 실시간 이벤트 브로드캐스팅
- 더 나은 사용자 경험

**타워 디펜스는 라운드 기반이라 3초 폴링도 충분합니다!**
