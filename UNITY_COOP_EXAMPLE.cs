// CoopStateSync.cs
// Unity 클라이언트에서 협동 플레이 상태 동기화를 위한 예제 코드
// Assets/Scripts/Networking/ 폴더에 추가하세요

using System.Collections;
using UnityEngine;
using LottoDefense.Backend;
using LottoDefense.Gameplay;

namespace LottoDefense.Networking
{
    /// <summary>
    /// REST polling을 통한 협동 플레이 상태 동기화
    /// 3초마다 내 상태를 서버에 전송하고 상대방 상태를 수신
    /// </summary>
    public class CoopStateSync : MonoBehaviour
    {
        #region Singleton
        private static CoopStateSync _instance;
        public static CoopStateSync Instance
        {
            get
            {
                if (_instance == null)
                {
                    GameObject go = new GameObject("CoopStateSync");
                    _instance = go.AddComponent<CoopStateSync>();
                    DontDestroyOnLoad(go);
                }
                return _instance;
            }
        }
        #endregion

        #region Properties
        private uint roomID;
        private float syncInterval = 3f; // 3초마다 동기화
        private Coroutine syncCoroutine;
        private bool isActive = false;

        public OpponentStateResponse OpponentState { get; private set; }
        #endregion

        #region Public Methods
        /// <summary>
        /// 협동 플레이 동기화 시작
        /// </summary>
        public void StartSync(uint roomID)
        {
            this.roomID = roomID;
            isActive = true;
            
            if (syncCoroutine != null)
            {
                StopCoroutine(syncCoroutine);
            }
            
            syncCoroutine = StartCoroutine(SyncLoop());
            Debug.Log($"[CoopStateSync] Started syncing for room {roomID}");
        }

        /// <summary>
        /// 협동 플레이 동기화 중지
        /// </summary>
        public void StopSync()
        {
            isActive = false;
            
            if (syncCoroutine != null)
            {
                StopCoroutine(syncCoroutine);
                syncCoroutine = null;
            }
            
            Debug.Log("[CoopStateSync] Stopped syncing");
        }
        #endregion

        #region Private Methods
        private IEnumerator SyncLoop()
        {
            while (isActive)
            {
                // 1. 내 상태 전송
                yield return UpdateMyState();
                
                // 2. 0.5초 대기
                yield return new WaitForSeconds(0.5f);
                
                // 3. 상대방 상태 수신
                yield return GetOpponentState();
                
                // 4. 다음 동기화까지 대기
                yield return new WaitForSeconds(syncInterval - 0.5f);
            }
        }

        private IEnumerator UpdateMyState()
        {
            if (GameplayManager.Instance == null)
            {
                yield break;
            }

            var state = new UpdateGameStateRequest
            {
                round = GameplayManager.Instance.CurrentRound,
                hp = GameplayManager.Instance.CurrentLife,
                gold = GameplayManager.Instance.CurrentGold,
                kills = MonsterManager.Instance != null ? MonsterManager.Instance.TotalKills : 0,
                timestamp = System.DateTimeOffset.UtcNow.ToUnixTimeSeconds()
            };

            yield return APIClient.Post(
                $"/coop/rooms/{roomID}/state",
                state,
                onSuccess: (json) => {
                    // 상태 업데이트 성공
                },
                onError: (error) => {
                    Debug.LogWarning($"[CoopStateSync] State update failed: {error}");
                }
            );
        }

        private IEnumerator GetOpponentState()
        {
            yield return APIClient.Get(
                $"/coop/rooms/{roomID}/opponent-state",
                onSuccess: (json) => {
                    try
                    {
                        var response = JsonUtility.FromJson<APIResponse<OpponentStateResponse>>(json);
                        if (response.success && response.data != null)
                        {
                            OpponentState = response.data;
                            UpdateOpponentUI(OpponentState);
                        }
                    }
                    catch (System.Exception e)
                    {
                        Debug.LogError($"[CoopStateSync] Failed to parse opponent state: {e.Message}");
                    }
                },
                onError: (error) => {
                    // 상대방이 아직 상태를 보내지 않았을 수 있음
                    Debug.LogWarning($"[CoopStateSync] Opponent state not available: {error}");
                }
            );
        }

        private void UpdateOpponentUI(OpponentStateResponse state)
        {
            if (OpponentStatusUI.Instance != null)
            {
                OpponentStatusUI.Instance.SetHP(state.hp);
                OpponentStatusUI.Instance.SetRound(state.round);
                OpponentStatusUI.Instance.SetGold(state.gold);
                OpponentStatusUI.Instance.SetKills(state.kills);
                OpponentStatusUI.Instance.SetName(state.opponent_name);
                
                if (!state.is_alive)
                {
                    OpponentStatusUI.Instance.ShowDefeat();
                }
            }
        }
        #endregion
    }

    #region Data Models
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

    [System.Serializable]
    public class APIResponse<T>
    {
        public bool success;
        public T data;
    }
    #endregion
}

// ==============================================================================
// 사용 방법
// ==============================================================================

/*
1. GameplayManager에서 게임 시작 시:

public void StartCoopGame(uint roomID, GameDifficulty difficulty)
{
    StartGame(difficulty);
    CoopStateSync.Instance.StartSync(roomID);
}

2. GameplayManager에서 게임 종료 시:

private void OnGameEnd()
{
    CoopStateSync.Instance.StopSync();
    // 결과 전송...
}

3. UI에서 상대방 상태 표시:

void Update()
{
    if (CoopStateSync.Instance.OpponentState != null)
    {
        var state = CoopStateSync.Instance.OpponentState;
        opponentHPText.text = $"HP: {state.hp}";
        opponentRoundText.text = $"Round: {state.round}";
    }
}

4. MainGameBootstrapper에서 협동 플레이 시작:

private void OnCoopPlayClicked()
{
    // 난이도 선택 후...
    uint roomID = 123; // 방 참가 시 받은 roomID
    GameplayManager.Instance.StartCoopGame(roomID, selectedDifficulty);
}
*/
