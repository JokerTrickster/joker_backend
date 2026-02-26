# Unity Mobile Game Integration Guide 🎮

이 가이드는 Unity 모바일 게임과 Joker Backend를 연동하는 방법을 설명합니다.

## 📋 목차
1. [백엔드 준비 상태](#백엔드-준비-상태)
2. [빠른 시작](#빠른-시작)
3. [Unity 설정](#unity-설정)
4. [API 연동](#api-연동)
5. [멀티플레이 구현](#멀티플레이-구현)
6. [문제 해결](#문제-해결)

## 백엔드 준비 상태

### ✅ 완료된 기능
- **REST API**: 인증, 게임 데이터, 멀티플레이
- **WebSocket**: 실시간 통신 지원
- **파일 업로드**: S3 연동 멀티파트 업로드
- **CORS**: Unity WebGL/모바일 지원
- **JWT 인증**: Bearer Token 방식
- **Swagger 문서**: `/swagger/*`

### 🚀 서비스 엔드포인트
| 서비스 | 포트 | 설명 |
|--------|------|------|
| Auth Service | 18081 | 인증 및 사용자 관리 |
| Tower Defense | 18082 | 게임 로직 |
| Cloud Repository | 18080 | 파일 업로드/WebSocket |

## 빠른 시작

### 1. 백엔드 서버 실행

```bash
# 1. Docker 인프라 실행 (MySQL, Redis)
docker-compose up -d

# 2. 데이터베이스 마이그레이션
./scripts/migrate.sh

# 3. 서비스 실행
./scripts/deploy-service.sh
```

### 2. 환경 변수 설정

```bash
export CORS_ALLOWED_ORIGINS="http://localhost:*,file://*"  # Unity 개발용
export JWT_SECRET="your-secret-key"
export IS_LOCAL=true
export DB_USER=root
export DB_PASSWORD=luxrobo1!
export REDIS_PASSWORD=""
```

## Unity 설정

### 1. 필수 패키지 설치

Unity Package Manager에서 설치:
- `com.unity.nuget.newtonsoft-json` (JSON 파싱)
- `com.unity.modules.unitywebrequest` (HTTP 통신)

### 2. GameAPI 매니저 생성

```csharp
using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;
using UnityEngine;
using UnityEngine.Networking;
using Newtonsoft.Json;

public class GameAPI : MonoBehaviour
{
    // 서버 설정
    private string authServiceURL = "http://localhost:18081";
    private string gameServiceURL = "http://localhost:18082";
    private string fileServiceURL = "http://localhost:18080";

    // 인증 토큰
    private string accessToken;
    private string refreshToken;
    private int userId;

    // Singleton
    private static GameAPI instance;
    public static GameAPI Instance
    {
        get
        {
            if (instance == null)
            {
                GameObject go = new GameObject("GameAPI");
                instance = go.AddComponent<GameAPI>();
                DontDestroyOnLoad(go);
            }
            return instance;
        }
    }

    void Awake()
    {
        if (instance != null && instance != this)
        {
            Destroy(gameObject);
            return;
        }
        instance = this;
        DontDestroyOnLoad(gameObject);
    }
}
```

## API 연동

### 1. 회원가입

```csharp
[System.Serializable]
public class SignupRequest
{
    public string email;
    public string password;
    public string name;
    public string service_type = "game";
}

[System.Serializable]
public class AuthResponse
{
    public bool success;
    public AuthData data;
    public string error;
}

[System.Serializable]
public class AuthData
{
    public int user_id;
    public string access_token;
    public string refresh_token;
}

public IEnumerator Signup(string email, string password, string name, Action<bool> callback)
{
    SignupRequest request = new SignupRequest
    {
        email = email,
        password = password,
        name = name
    };

    string json = JsonConvert.SerializeObject(request);
    byte[] bodyRaw = Encoding.UTF8.GetBytes(json);

    using (UnityWebRequest www = new UnityWebRequest($"{authServiceURL}/v0.1/auth/signup", "POST"))
    {
        www.uploadHandler = new UploadHandlerRaw(bodyRaw);
        www.downloadHandler = new DownloadHandlerBuffer();
        www.SetRequestHeader("Content-Type", "application/json");

        yield return www.SendWebRequest();

        if (www.result == UnityWebRequest.Result.Success)
        {
            AuthResponse response = JsonConvert.DeserializeObject<AuthResponse>(www.downloadHandler.text);
            if (response.success)
            {
                accessToken = response.data.access_token;
                refreshToken = response.data.refresh_token;
                userId = response.data.user_id;

                Debug.Log($"Signup successful! UserID: {userId}");
                callback?.Invoke(true);
            }
            else
            {
                Debug.LogError($"Signup failed: {response.error}");
                callback?.Invoke(false);
            }
        }
        else
        {
            Debug.LogError($"Request failed: {www.error}");
            callback?.Invoke(false);
        }
    }
}
```

### 2. 로그인

```csharp
public IEnumerator Login(string email, string password, Action<bool> callback)
{
    var request = new
    {
        email = email,
        password = password,
        service_type = "game"
    };

    string json = JsonConvert.SerializeObject(request);
    byte[] bodyRaw = Encoding.UTF8.GetBytes(json);

    using (UnityWebRequest www = new UnityWebRequest($"{authServiceURL}/v0.1/auth/signin", "POST"))
    {
        www.uploadHandler = new UploadHandlerRaw(bodyRaw);
        www.downloadHandler = new DownloadHandlerBuffer();
        www.SetRequestHeader("Content-Type", "application/json");

        yield return www.SendWebRequest();

        if (www.result == UnityWebRequest.Result.Success)
        {
            AuthResponse response = JsonConvert.DeserializeObject<AuthResponse>(www.downloadHandler.text);
            if (response.success)
            {
                accessToken = response.data.access_token;
                refreshToken = response.data.refresh_token;
                userId = response.data.user_id;

                Debug.Log($"Login successful! UserID: {userId}");
                callback?.Invoke(true);
            }
            else
            {
                Debug.LogError($"Login failed: {response.error}");
                callback?.Invoke(false);
            }
        }
        else
        {
            Debug.LogError($"Request failed: {www.error}");
            callback?.Invoke(false);
        }
    }
}
```

### 3. 인증이 필요한 API 호출

```csharp
public IEnumerator GetGameHistory(Action<string> callback)
{
    using (UnityWebRequest www = UnityWebRequest.Get($"{gameServiceURL}/api/v1/td/game/history"))
    {
        www.SetRequestHeader("Authorization", $"Bearer {accessToken}");
        www.SetRequestHeader("Content-Type", "application/json");

        yield return www.SendWebRequest();

        if (www.result == UnityWebRequest.Result.Success)
        {
            callback?.Invoke(www.downloadHandler.text);
        }
        else if (www.responseCode == 401) // Unauthorized
        {
            Debug.Log("Token expired, refreshing...");
            yield return RefreshToken((success) =>
            {
                if (success)
                {
                    StartCoroutine(GetGameHistory(callback));
                }
            });
        }
        else
        {
            Debug.LogError($"Request failed: {www.error}");
            callback?.Invoke(null);
        }
    }
}
```

### 4. 토큰 갱신

```csharp
public IEnumerator RefreshToken(Action<bool> callback)
{
    var request = new
    {
        refresh_token = refreshToken
    };

    string json = JsonConvert.SerializeObject(request);
    byte[] bodyRaw = Encoding.UTF8.GetBytes(json);

    using (UnityWebRequest www = new UnityWebRequest($"{authServiceURL}/v0.1/auth/refresh", "POST"))
    {
        www.uploadHandler = new UploadHandlerRaw(bodyRaw);
        www.downloadHandler = new DownloadHandlerBuffer();
        www.SetRequestHeader("Content-Type", "application/json");

        yield return www.SendWebRequest();

        if (www.result == UnityWebRequest.Result.Success)
        {
            AuthResponse response = JsonConvert.DeserializeObject<AuthResponse>(www.downloadHandler.text);
            if (response.success)
            {
                accessToken = response.data.access_token;
                Debug.Log("Token refreshed successfully");
                callback?.Invoke(true);
            }
            else
            {
                callback?.Invoke(false);
            }
        }
        else
        {
            callback?.Invoke(false);
        }
    }
}
```

## 멀티플레이 구현

### Co-op 룸 시스템

```csharp
public class CoopManager : MonoBehaviour
{
    private string roomId;
    private bool isHost = false;
    private Coroutine syncCoroutine;

    // 방 생성
    public IEnumerator CreateRoom(string roomName, Action<string> callback)
    {
        var request = new
        {
            room_name = roomName,
            max_players = 4
        };

        string json = JsonConvert.SerializeObject(request);
        byte[] bodyRaw = Encoding.UTF8.GetBytes(json);

        using (UnityWebRequest www = new UnityWebRequest($"{gameServiceURL}/v0.1/coop/rooms", "POST"))
        {
            www.uploadHandler = new UploadHandlerRaw(bodyRaw);
            www.downloadHandler = new DownloadHandlerBuffer();
            www.SetRequestHeader("Authorization", $"Bearer {GameAPI.Instance.AccessToken}");
            www.SetRequestHeader("Content-Type", "application/json");

            yield return www.SendWebRequest();

            if (www.result == UnityWebRequest.Result.Success)
            {
                var response = JsonConvert.DeserializeObject<RoomResponse>(www.downloadHandler.text);
                roomId = response.data.room_id;
                isHost = true;

                // 상태 동기화 시작
                StartStateSync();

                callback?.Invoke(roomId);
            }
        }
    }

    // 방 참가
    public IEnumerator JoinRoom(string roomCode, Action<bool> callback)
    {
        using (UnityWebRequest www = UnityWebRequest.Post($"{gameServiceURL}/v0.1/coop/rooms/{roomCode}/join", ""))
        {
            www.SetRequestHeader("Authorization", $"Bearer {GameAPI.Instance.AccessToken}");

            yield return www.SendWebRequest();

            if (www.result == UnityWebRequest.Result.Success)
            {
                roomId = roomCode;
                isHost = false;

                // 상태 동기화 시작
                StartStateSync();

                callback?.Invoke(true);
            }
            else
            {
                callback?.Invoke(false);
            }
        }
    }

    // 3초마다 상태 동기화 (REST Polling)
    private void StartStateSync()
    {
        if (syncCoroutine != null)
            StopCoroutine(syncCoroutine);

        syncCoroutine = StartCoroutine(SyncGameState());
    }

    private IEnumerator SyncGameState()
    {
        while (!string.IsNullOrEmpty(roomId))
        {
            yield return new WaitForSeconds(3f);

            using (UnityWebRequest www = UnityWebRequest.Get($"{gameServiceURL}/v0.1/coop/rooms/{roomId}/state"))
            {
                www.SetRequestHeader("Authorization", $"Bearer {GameAPI.Instance.AccessToken}");

                yield return www.SendWebRequest();

                if (www.result == UnityWebRequest.Result.Success)
                {
                    var state = JsonConvert.DeserializeObject<GameState>(www.downloadHandler.text);
                    UpdateLocalGameState(state);
                }
            }
        }
    }

    private void UpdateLocalGameState(GameState state)
    {
        // 게임 상태 업데이트 로직
        // 예: 플레이어 위치, 점수, 게임 진행 상황 등
    }
}
```

### WebSocket 실시간 통신 (선택사항)

```csharp
using System;
using System.Collections;
using UnityEngine;
using NativeWebSocket;

public class WebSocketManager : MonoBehaviour
{
    private WebSocket websocket;

    public async void Connect(string token)
    {
        string wsUrl = $"ws://localhost:18080/ws?token={token}";

        websocket = new WebSocket(wsUrl);

        websocket.OnOpen += () =>
        {
            Debug.Log("WebSocket connected!");
        };

        websocket.OnError += (e) =>
        {
            Debug.LogError($"WebSocket error: {e}");
        };

        websocket.OnClose += (e) =>
        {
            Debug.Log("WebSocket closed");
        };

        websocket.OnMessage += (bytes) =>
        {
            var message = System.Text.Encoding.UTF8.GetString(bytes);
            Debug.Log($"Received: {message}");

            // 메시지 처리
            ProcessMessage(message);
        };

        await websocket.Connect();
    }

    public async void SendMessage(string message)
    {
        if (websocket.State == WebSocketState.Open)
        {
            await websocket.SendText(message);
        }
    }

    private void ProcessMessage(string message)
    {
        // 메시지 타입에 따른 처리
        var data = JsonConvert.DeserializeObject<WebSocketMessage>(message);

        switch (data.type)
        {
            case "game_update":
                // 게임 업데이트 처리
                break;
            case "player_action":
                // 플레이어 액션 처리
                break;
        }
    }

    private void Update()
    {
        #if !UNITY_WEBGL || UNITY_EDITOR
            websocket?.DispatchMessageQueue();
        #endif
    }

    private async void OnDestroy()
    {
        if (websocket != null)
        {
            await websocket.Close();
        }
    }
}
```

## 문제 해결

### CORS 에러
Unity WebGL 빌드에서 CORS 에러가 발생하는 경우:

```bash
# 백엔드 환경변수 설정
export CORS_ALLOWED_ORIGINS="http://localhost:*,file://*,null"
```

### 인증 토큰 만료
401 Unauthorized 에러 시 자동으로 토큰 갱신:

```csharp
if (www.responseCode == 401)
{
    yield return RefreshToken((success) =>
    {
        if (success)
        {
            // 원래 요청 재시도
            StartCoroutine(OriginalRequest());
        }
    });
}
```

### 모바일 빌드 설정

**Android:**
- `Player Settings > Other Settings > Internet Access: Require`
- Android 9+ 에서 HTTP 사용 시 `AndroidManifest.xml`에 추가:
  ```xml
  <application android:usesCleartextTraffic="true">
  ```

**iOS:**
- iOS 9+ 에서 HTTP 사용 시 `Info.plist`에 추가:
  ```xml
  <key>NSAppTransportSecurity</key>
  <dict>
      <key>NSAllowsArbitraryLoads</key>
      <true/>
  </dict>
  ```

### 성능 최적화

1. **요청 캐싱**: 자주 사용하는 데이터 로컬 캐싱
2. **배치 요청**: 여러 요청을 하나로 묶기
3. **압축**: 큰 데이터는 gzip 압축 사용
4. **연결 풀링**: HTTP 연결 재사용

## 다음 단계

1. **프로덕션 배포**
   - HTTPS 설정
   - 도메인 설정
   - 로드 밸런서 구성

2. **보안 강화**
   - API 키 관리
   - Rate limiting
   - 암호화 통신

3. **모니터링**
   - 에러 추적
   - 성능 모니터링
   - 사용자 분석

## 지원

문제가 있으시면 GitHub Issues에 등록해주세요:
https://github.com/JokerTrickster/joker_backend/issues