# Unity Frontend - Backend Integration Guide

## ✅ Backend Implementation Status

### Completed Features

#### 1. Authentication System ✓
- **Endpoint**: `POST /api/v1/td/auth/login`
- **Token Refresh**: `POST /api/v1/td/auth/refresh`
- JWT-based authentication implemented
- Token validation middleware ready

#### 2. Game Session Management ✓
- **Create Session**: `POST /api/v1/td/game/session`
- **Join Session**: `POST /api/v1/td/game/session/join`
- **Get Session Info**: `GET /api/v1/td/game/session/{sessionId}`
- Room code generation for co-op mode
- Session status tracking (waiting, playing, finished)

#### 3. Game Result Saving ✓
- **Save Result**: `POST /api/v1/td/game/result`
- Player statistics tracking
- Experience and level progression system
- Victory/defeat tracking

#### 4. Player Profile ✓
- **Get Profile**: `GET /api/v1/td/profile/{userId}`
- Complete statistics tracking:
  - Games played, victories, scores
  - Highest wave, units placed, enemies defeated
  - Total play time

#### 5. WebSocket Real-time Communication ✓
- **WebSocket URL**: `ws://[server]/ws/game/{sessionId}?token={jwt_token}`
- Message types implemented:
  - Authentication handshake
  - Unit placement/removal sync
  - Game state synchronization
  - Wave start/complete events
  - Player join/leave notifications

## 🔧 Unity Frontend - Required Updates

### 1. Environment Configuration
Create a configuration system to manage different environments:

```csharp
// ScriptableObject for environment configuration
[CreateAssetMenu(fileName = "EnvironmentConfig", menuName = "Config/Environment")]
public class EnvironmentConfig : ScriptableObject
{
    public string apiUrl;
    public string wsUrl;
    public int timeout = 15000;
    public int retryCount = 3;
}

// Environment Manager
public class EnvironmentManager : MonoBehaviour
{
    public EnvironmentConfig development;
    public EnvironmentConfig staging;
    public EnvironmentConfig production;

    private EnvironmentConfig currentConfig;

    public static string ApiUrl => Instance.currentConfig.apiUrl;
    public static string WsUrl => Instance.currentConfig.wsUrl;
}
```

### 2. HTTP Retry Logic Implementation
Add exponential backoff retry mechanism:

```csharp
public class RetryableHttpClient
{
    private const int MaxRetries = 3;
    private readonly int[] retryDelays = { 1000, 2000, 4000 }; // milliseconds

    public async Task<T> SendWithRetry<T>(UnityWebRequest request)
    {
        for (int i = 0; i < MaxRetries; i++)
        {
            using (request)
            {
                await request.SendWebRequest();

                if (request.result == UnityWebRequest.Result.Success)
                {
                    return JsonUtility.FromJson<T>(request.downloadHandler.text);
                }

                if (i < MaxRetries - 1)
                {
                    await Task.Delay(retryDelays[i]);
                }
            }
        }

        throw new Exception($"Request failed after {MaxRetries} retries");
    }
}
```

### 3. Secure Token Storage
Implement encrypted storage for JWT tokens:

```csharp
using System.Security.Cryptography;

public class SecureTokenStorage
{
    private const string TokenKey = "jwt_token_encrypted";
    private const string Salt = "your_salt_here"; // Change this

    public static void SaveToken(string token)
    {
        string encrypted = Encrypt(token, Salt);
        PlayerPrefs.SetString(TokenKey, encrypted);
        PlayerPrefs.Save();
    }

    public static string GetToken()
    {
        string encrypted = PlayerPrefs.GetString(TokenKey, "");
        return string.IsNullOrEmpty(encrypted) ? null : Decrypt(encrypted, Salt);
    }

    // Implement AES encryption/decryption methods
}
```

### 4. WebSocket Connection Manager
Update WebSocket implementation for proper authentication:

```csharp
public class GameWebSocketClient : MonoBehaviour
{
    private WebSocket websocket;
    private string sessionId;
    private Queue<NetworkMessage> messageQueue = new Queue<NetworkMessage>();

    public async Task Connect(string sessionId, string token)
    {
        this.sessionId = sessionId;
        string wsUrl = $"{EnvironmentManager.WsUrl}/game/{sessionId}?token={token}";

        websocket = new WebSocket(wsUrl);

        websocket.OnOpen += OnOpen;
        websocket.OnMessage += OnMessage;
        websocket.OnError += OnError;
        websocket.OnClose += OnClose;

        await websocket.Connect();
    }

    private void OnOpen()
    {
        Debug.Log("WebSocket connected");
        StartCoroutine(HeartbeatCoroutine());
    }

    private IEnumerator HeartbeatCoroutine()
    {
        while (websocket.State == WebSocketState.Open)
        {
            yield return new WaitForSeconds(15f);
            SendHeartbeat();
        }
    }
}
```

### 5. API Client Updates
Update the APIClient with proper error handling:

```csharp
public class TDApiClient : MonoBehaviour
{
    private RetryableHttpClient httpClient = new RetryableHttpClient();

    public async Task<LoginResponse> Login(string username, string password)
    {
        var request = new LoginRequest { username = username, password = password };
        var json = JsonUtility.ToJson(request);

        using (var webRequest = UnityWebRequest.Post($"{EnvironmentManager.ApiUrl}/auth/login", json))
        {
            webRequest.SetRequestHeader("Content-Type", "application/json");

            var response = await httpClient.SendWithRetry<LoginResponse>(webRequest);

            // Save token securely
            SecureTokenStorage.SaveToken(response.token);

            return response;
        }
    }

    // Implement other API methods...
}
```

## 📝 Testing Checklist

### Local Testing (Development)
- [ ] Update BASE_URL to `http://localhost:18082/api/v1/td`
- [ ] Test login flow with mock credentials
- [ ] Create single-player session
- [ ] Create co-op session and get room code
- [ ] Join co-op session with room code
- [ ] Test WebSocket connection
- [ ] Test unit placement synchronization
- [ ] Save game results
- [ ] Verify player stats update

### Staging Testing
- [ ] Update URLs to staging server
- [ ] Test with real user accounts
- [ ] Test cross-device co-op play
- [ ] Verify WebSocket reconnection logic
- [ ] Test under poor network conditions
- [ ] Verify token refresh flow

### Production Deployment
- [ ] Enable certificate pinning
- [ ] Verify all environment variables
- [ ] Test load balancing
- [ ] Monitor error rates
- [ ] Check WebSocket scaling

## 🚀 Quick Start Guide

### 1. Start Backend Services
```bash
# Clone and setup backend
git clone https://github.com/JokerTrickster/joker_backend
cd joker_backend

# Start services with Docker
docker-compose up -d

# Verify services are running
curl http://localhost:18082/health
# Should return: {"status":"healthy"}
```

### 2. Test Authentication
```bash
# Login
curl -X POST http://localhost:18082/api/v1/td/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}'

# Response:
{
  "token": "eyJhbGc...",
  "userId": "123",
  "profile": {
    "nickname": "testuser",
    "avatarId": "default_avatar",
    "level": 1,
    "experience": 0
  }
}
```

### 3. Create Game Session
```bash
# Create co-op session
curl -X POST http://localhost:18082/api/v1/td/game/session \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"mode":"coop","playerCount":2}'

# Response:
{
  "sessionId": "uuid-here",
  "roomCode": "ABC123",
  "webSocketUrl": "ws://localhost:18082/ws/game/uuid-here"
}
```

### 4. Connect WebSocket
```javascript
// JavaScript example for testing
const ws = new WebSocket('ws://localhost:18082/ws/game/SESSION_ID?token=YOUR_TOKEN');

ws.onopen = () => {
  console.log('Connected to game session');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

// Send unit placed message
ws.send(JSON.stringify({
  type: 'unit_placed',
  payload: {
    playerNumber: 1,
    position: { x: 2, y: 3 },
    unitType: 'tower_basic'
  },
  timestamp: Date.now()
}));
```

## ⚠️ Important Notes

### Security Considerations
1. **Token Storage**: Never store tokens in plain text
2. **Certificate Pinning**: Implement for production
3. **API Rate Limiting**: Backend implements rate limiting
4. **Input Validation**: All inputs are validated server-side

### Performance Optimization
1. **WebSocket Reconnection**: Implement exponential backoff
2. **Message Batching**: Batch multiple updates when possible
3. **State Compression**: Large state updates are compressed
4. **Connection Pooling**: Reuse HTTP connections

### Error Handling
1. **Network Errors**: Show user-friendly messages
2. **Token Expiry**: Auto-refresh before expiry
3. **WebSocket Disconnect**: Auto-reconnect with state recovery
4. **Server Errors**: Log and report to analytics

## 📞 Support

For backend issues or API questions, please contact the backend team or create an issue in the repository.

Backend Repository: https://github.com/JokerTrickster/joker_backend

## 🔄 API Version History

- **v1.0.0** (2024-02-26): Initial release with all core features
  - Authentication system
  - Game session management
  - WebSocket real-time sync
  - Player profiles and statistics