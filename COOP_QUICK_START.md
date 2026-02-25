# 협동 플레이 빠른 시작 가이드

## 1단계: 데이터베이스 업데이트

```sql
cd ~/project/joker_backend
mysql -u root -p lotto_defense < ALTER_TD_ROOMS_FOR_COOP.sql
```

또는:

```sql
ALTER TABLE td_rooms 
ADD COLUMN player1_state JSON,
ADD COLUMN player2_state JSON,
ADD COLUMN last_p1_update TIMESTAMP NULL,
ADD COLUMN last_p2_update TIMESTAMP NULL;
```

## 2단계: 백엔드 재시작

```bash
cd ~/project/joker_backend/services/lottoDefenseService
go run main.go
```

## 3단계: API 테스트

### 방 생성
```bash
curl -X POST http://localhost:8080/v0.1/coop/rooms \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "room_type": "private",
    "max_players": 2
  }'
```

**응답:**
```json
{
  "success": true,
  "data": {
    "room_id": 1,
    "room_code": "AB12",
    "status": "waiting"
  }
}
```

### 상태 업데이트 (Player 1)
```bash
curl -X POST http://localhost:8080/v0.1/coop/rooms/1/state \
  -H "Authorization: Bearer PLAYER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "round": 5,
    "hp": 87,
    "gold": 1200,
    "kills": 45,
    "timestamp": 1234567890
  }'
```

**응답:**
```json
{
  "success": true,
  "data": {
    "status": "updated"
  }
}
```

### 상대방 상태 조회 (Player 2)
```bash
curl -X GET http://localhost:8080/v0.1/coop/rooms/1/opponent-state \
  -H "Authorization: Bearer PLAYER2_TOKEN"
```

**응답:**
```json
{
  "success": true,
  "data": {
    "opponent_id": 1,
    "opponent_name": "Player1",
    "round": 5,
    "hp": 87,
    "gold": 1200,
    "kills": 45,
    "last_update": 1234567890,
    "is_alive": true
  }
}
```

## 4단계: Unity 클라이언트 통합

### 파일 생성
```
Assets/Scripts/Networking/CoopStateSync.cs
```

**내용:** `UNITY_COOP_EXAMPLE.cs` 파일 참고

### GameplayManager 수정

```csharp
public void StartCoopGame(uint roomID, GameDifficulty difficulty)
{
    SetDifficulty(difficulty);
    CurrentState = GameState.Countdown;
    
    // 협동 플레이 동기화 시작
    CoopStateSync.Instance.StartSync(roomID);
    
    StartCountdown();
}

private void OnDestroy()
{
    CoopStateSync.Instance.StopSync();
}
```

### MainGameBootstrapper 수정

```csharp
private void OnCoopPlayClicked()
{
    DifficultySelectionUI.Show(true, (difficulty) => {
        // 방 생성 또는 참가
        CreateOrJoinRoom(difficulty);
    });
}

private void CreateOrJoinRoom(GameDifficulty difficulty)
{
    // 방 생성 API 호출
    StartCoroutine(APIClient.Post("/coop/rooms", 
        new { room_type = "private" },
        onSuccess: (json) => {
            var response = JsonUtility.FromJson<RoomResponse>(json);
            uint roomID = response.data.room_id;
            
            // 게임 시작
            SceneManager.LoadScene("GameScene");
            GameplayManager.Instance.StartCoopGame(roomID, difficulty);
        }
    ));
}
```

## 5단계: 테스트

### 시나리오 1: 로컬 테스트
1. Unity 실행 (Player 1)
2. 협동 플레이 → 방 생성
3. 두 번째 Unity 실행 (Player 2)
4. 방 코드로 참가
5. 양쪽 플레이어 Ready
6. 게임 시작
7. Console에서 동기화 로그 확인

### 시나리오 2: 빌드 테스트
1. Unity 빌드 (2개)
2. 각각 실행
3. 한쪽에서 방 생성
4. 다른 쪽에서 참가
5. 게임 진행하면서 상대방 상태 확인

## 디버깅

### 백엔드 로그 확인
```bash
tail -f ~/project/joker_backend/logs/app.log
```

### Unity Console 로그
```
[CoopStateSync] Started syncing for room 1
[CoopStateSync] State updated
[CoopStateSync] Opponent state: round=5, hp=90
```

### 데이터베이스 확인
```sql
SELECT 
    id, 
    room_code, 
    status, 
    player1_state, 
    player2_state,
    last_p1_update,
    last_p2_update
FROM td_rooms 
WHERE id = 1;
```

## 문제 해결

### 1. "opponent not found" 에러
- 상대방이 아직 상태를 전송하지 않았음
- 3초 후 다시 시도됨

### 2. 동기화가 안 됨
- JWT 토큰 확인
- roomID가 올바른지 확인
- 두 플레이어가 같은 방에 있는지 확인

### 3. 지연이 심함
- `syncInterval`을 2초로 줄이기
- 네트워크 상태 확인

## 성능 최적화

### 동적 동기화 간격
```csharp
// 라운드 시작 시 1초로 단축
if (isRoundStarting)
{
    syncInterval = 1f;
}
// 평상시 3초
else
{
    syncInterval = 3f;
}
```

### 중요 이벤트 즉시 동기화
```csharp
// HP가 변경되었을 때
public void OnHPChanged(int newHP)
{
    if (newHP < GameplayManager.Instance.CurrentLife)
    {
        // 즉시 동기화
        StartCoroutine(CoopStateSync.Instance.UpdateMyState());
    }
}
```

## 다음 단계

### Phase 2: WebSocket 업그레이드
- 실시간 동기화
- 지연 시간 0.1초 이하
- 더 나은 사용자 경험

### Phase 3: 서버 검증
- 치트 감지
- 게임 로직 검증
- 리플레이 시스템

---

**완료! 이제 2인 협동 타워 디펜스를 플레이할 수 있습니다!** 🎮
