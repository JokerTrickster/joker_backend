# 협동 플레이 구현 가이드 (REST Polling 방식)

## 개요
WebSocket 없이 REST API polling으로 간단하게 협동 플레이 구현

## 구현 순서

### 1. 데이터베이스 테이블 수정

기존 `td_rooms` 테이블에 컬럼 추가:

```sql
ALTER TABLE td_rooms 
ADD COLUMN player1_state JSON,
ADD COLUMN player2_state JSON,
ADD COLUMN last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
```

**player_state JSON 구조:**
```json
{
  "round": 5,
  "hp": 87,
  "gold": 1200,
  "kills": 45,
  "timestamp": 1234567890,
  "is_alive": true
}
```

### 2. Repository 메서드 추가

`room_repository.go`에 추가:

```go
func (r *TDRoomRepository) UpdatePlayerState(
    ctx context.Context,
    roomID uint,
    userID uint,
    state *request.UpdateGameStateRequest,
) error {
    stateJSON, _ := json.Marshal(map[string]interface{}{
        "round":     state.Round,
        "hp":        state.HP,
        "gold":      state.Gold,
        "kills":     state.Kills,
        "timestamp": state.Timestamp,
        "is_alive":  state.HP > 0,
    })

    // Find which player position
    var room entity.TDRoom
    if err := r.db.WithContext(ctx).First(&room, roomID).Error; err != nil {
        return err
    }

    if room.HostUserID == userID {
        return r.db.WithContext(ctx).
            Model(&entity.TDRoom{}).
            Where("id = ?", roomID).
            Update("player1_state", stateJSON).Error
    } else {
        return r.db.WithContext(ctx).
            Model(&entity.TDRoom{}).
            Where("id = ?", roomID).
            Update("player2_state", stateJSON).Error
    }
}

func (r *TDRoomRepository) GetOpponentState(
    ctx context.Context,
    roomID uint,
    userID uint,
) (*response.OpponentStateResponse, error) {
    var room entity.TDRoom
    if err := r.db.WithContext(ctx).
        Preload("HostUser").
        Preload("RoomPlayers.User").
        First(&room, roomID).Error; err != nil {
        return nil, err
    }

    var opponentState string
    var opponentID uint
    var opponentName string

    if room.HostUserID == userID {
        // Current user is host, get player2 state
        opponentState = room.Player2State
        if len(room.RoomPlayers) > 0 {
            opponentID = room.RoomPlayers[0].UserID
            opponentName = room.RoomPlayers[0].User.Name
        }
    } else {
        // Current user is player2, get host state
        opponentState = room.Player1State
        opponentID = room.HostUserID
        opponentName = room.HostUser.Name
    }

    if opponentState == "" {
        return nil, errors.New("opponent not found")
    }

    var state map[string]interface{}
    json.Unmarshal([]byte(opponentState), &state)

    return &response.OpponentStateResponse{
        OpponentID:   opponentID,
        OpponentName: opponentName,
        Round:        int(state["round"].(float64)),
        HP:           int(state["hp"].(float64)),
        Gold:         int(state["gold"].(float64)),
        Kills:        int(state["kills"].(float64)),
        LastUpdate:   int64(state["timestamp"].(float64)),
        IsAlive:      state["is_alive"].(bool),
    }, nil
}
```

### 3. UseCase 메서드 추가

`room_usecase.go`에 추가:

```go
func (u *TDRoomUseCase) UpdatePlayerState(
    ctx context.Context,
    roomID uint,
    userID uint,
    req *request.UpdateGameStateRequest,
) error {
    return u.roomRepo.UpdatePlayerState(ctx, roomID, userID, req)
}

func (u *TDRoomUseCase) GetOpponentState(
    ctx context.Context,
    roomID uint,
    userID uint,
) (*response.OpponentStateResponse, error) {
    return u.roomRepo.GetOpponentState(ctx, roomID, userID)
}
```

### 4. Interface 추가

`model/interface/room_repository.go`에 추가:

```go
type ITDRoomRepository interface {
    // ... existing methods
    UpdatePlayerState(ctx context.Context, roomID uint, userID uint, state *request.UpdateGameStateRequest) error
    GetOpponentState(ctx context.Context, roomID uint, userID uint) (*response.OpponentStateResponse, error)
}
```

`model/interface/usecase.go`에 추가:

```go
type ITDRoomUseCase interface {
    // ... existing methods
    UpdatePlayerState(ctx context.Context, roomID uint, userID uint, req *request.UpdateGameStateRequest) error
    GetOpponentState(ctx context.Context, roomID uint, userID uint) (*response.OpponentStateResponse, error)
}
```

### 5. Routes 등록

`handler/routes.go`에 추가:

```go
func RegisterTowerDefenseRoutes(e *echo.Echo, usecases *Usecases) {
    // ... existing routes
    
    // Co-op state sync
    NewTDCoopStateHandler(authGroup, usecases.RoomUseCase)
}
```

---

## Unity 클라이언트 구현

### CoopStateSync.cs

```csharp
using System.Collections;
using UnityEngine;
using LottoDefense.Backend;
using LottoDefense.Gameplay;

public class CoopStateSync : MonoBehaviour
{
    private uint roomID;
    private float syncInterval = 3f; // 3초마다 동기화
    private Coroutine syncCoroutine;

    public void StartSync(uint roomID)
    {
        this.roomID = roomID;
        if (syncCoroutine != null) StopCoroutine(syncCoroutine);
        syncCoroutine = StartCoroutine(SyncLoop());
    }

    public void StopSync()
    {
        if (syncCoroutine != null)
        {
            StopCoroutine(syncCoroutine);
            syncCoroutine = null;
        }
    }

    private IEnumerator SyncLoop()
    {
        while (true)
        {
            // 1. 내 상태 전송
            yield return UpdateMyState();
            
            // 2. 상대방 상태 수신
            yield return GetOpponentState();
            
            // 3. 대기
            yield return new WaitForSeconds(syncInterval);
        }
    }

    private IEnumerator UpdateMyState()
    {
        var state = new UpdateGameStateRequest
        {
            round = GameplayManager.Instance.CurrentRound,
            hp = GameplayManager.Instance.CurrentLife,
            gold = GameplayManager.Instance.CurrentGold,
            kills = MonsterManager.Instance.TotalKills,
            timestamp = DateTimeOffset.UtcNow.ToUnixTimeSeconds()
        };

        yield return APIClient.Post(
            $"/coop/rooms/{roomID}/state",
            state,
            onSuccess: (json) => {
                Debug.Log("State updated");
            },
            onError: (error) => {
                Debug.LogError($"State update failed: {error}");
            }
        );
    }

    private IEnumerator GetOpponentState()
    {
        yield return APIClient.Get(
            $"/coop/rooms/{roomID}/opponent-state",
            onSuccess: (json) => {
                var state = JsonUtility.FromJson<OpponentStateResponse>(json);
                UpdateOpponentUI(state);
            },
            onError: (error) => {
                // 상대방이 아직 상태를 보내지 않았을 수 있음
                Debug.LogWarning($"Opponent state not available: {error}");
            }
        );
    }

    private void UpdateOpponentUI(OpponentStateResponse state)
    {
        // UI 업데이트
        OpponentStatusUI.Instance.SetHP(state.hp);
        OpponentStatusUI.Instance.SetRound(state.round);
        OpponentStatusUI.Instance.SetGold(state.gold);
        
        if (!state.is_alive)
        {
            OpponentStatusUI.Instance.ShowDefeat();
        }
    }
}

[System.Serializable]
public class UpdateGameStateRequest
{
    public int round;
    public int hp;
    public int gold;
    public int kills;
    public long timestamp;
}

[System.Serializable]
public class OpponentStateResponse
{
    public uint opponent_id;
    public string opponent_name;
    public int round;
    public int hp;
    public int gold;
    public int kills;
    public long last_update;
    public bool is_alive;
}
```

---

## 사용 흐름

### 1. 방 생성 & 참가
```
Player 1: POST /coop/rooms (roomID: 123)
Player 2: POST /coop/rooms/join (roomID: 123)
Both: POST /coop/rooms/:id/ready
```

### 2. 게임 시작
```csharp
// Unity
GameplayManager.Instance.StartGame(difficulty);
CoopStateSync.Instance.StartSync(roomID);
```

### 3. 게임 진행 (3초마다)
```
POST /coop/rooms/123/state {round:5, hp:87, gold:1200}
GET /coop/rooms/123/opponent-state → {round:5, hp:90, gold:1100}
```

### 4. 게임 종료
```csharp
CoopStateSync.Instance.StopSync();
POST /game/single/result {rounds:30, hp:5} // 각자 결과 전송
```

---

## 장단점

### ✅ 장점
- 구현 매우 간단 (2-4시간)
- WebSocket 서버 불필요
- 방화벽/NAT 문제 없음
- 타워 디펜스에 충분

### ❌ 단점
- 3초 지연 (실시간성 낮음)
- 서버 요청 많음 (초당 0.33 req/user)
- 확장성 낮음

### 💡 최적화 팁
- 라운드 변경 시에만 즉시 동기화
- 평상시 5초, 이벤트 발생 시 1초로 동적 조정
- Redis 캐시로 DB 부하 감소

---

## 다음 단계: WebSocket 업그레이드

나중에 사용자가 많아지면:

1. Go + gorilla/websocket 서버 구현
2. 라운드 동기화, 이벤트 브로드캐스팅
3. 실시간 협동 플레이 경험 향상

**하지만 지금은 REST polling으로 충분합니다!**
