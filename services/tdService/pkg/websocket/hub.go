package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Sessions mapped by session ID
	sessions map[string]*GameRoom

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// GameRoom represents a game session with connected clients
type GameRoom struct {
	SessionID string
	Clients   map[*Client]bool
	State     *entity.GameStatePayload
	mu        sync.RWMutex
}

// Client represents a WebSocket connection
type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	SessionID string
	UserID    string
	PlayerNum int
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		sessions:   make(map[string]*GameRoom),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)
		}
	}
}

// registerClient registers a new client to a room
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.sessions[client.SessionID]
	if !exists {
		room = &GameRoom{
			SessionID: client.SessionID,
			Clients:   make(map[*Client]bool),
			State:     &entity.GameStatePayload{
				Wave:    0,
				Enemies: []entity.EnemyState{},
				Players: []entity.PlayerState{},
			},
		}
		h.sessions[client.SessionID] = room
	}

	room.mu.Lock()
	room.Clients[client] = true
	room.mu.Unlock()

	// Notify other players
	h.broadcastToRoom(client.SessionID, entity.WSTypePlayerJoined, map[string]string{
		"userId":   client.UserID,
		"playerNum": fmt.Sprintf("%d", client.PlayerNum),
	}, client)

	log.Printf("Client %s joined session %s", client.UserID, client.SessionID)
}

// unregisterClient removes a client from a room
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.sessions[client.SessionID]; exists {
		room.mu.Lock()
		if _, ok := room.Clients[client]; ok {
			delete(room.Clients, client)
			close(client.Send)

			// Notify other players
			h.broadcastToRoom(client.SessionID, entity.WSTypePlayerLeft, map[string]string{
				"userId":   client.UserID,
				"playerNum": fmt.Sprintf("%d", client.PlayerNum),
			}, client)

			// Clean up empty rooms
			if len(room.Clients) == 0 {
				delete(h.sessions, client.SessionID)
			}
		}
		room.mu.Unlock()

		log.Printf("Client %s left session %s", client.UserID, client.SessionID)
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(sessionID string, msgType entity.WSMessageType, payload interface{}) {
	h.broadcastToRoom(sessionID, msgType, payload, nil)
}

// broadcastToRoom sends a message to all clients in a room except the sender
func (h *Hub) broadcastToRoom(sessionID string, msgType entity.WSMessageType, payload interface{}, sender *Client) {
	h.mu.RLock()
	room, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	msg, err := entity.NewWSMessage(msgType, payload)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		return
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for client := range room.Clients {
		if sender != nil && client == sender {
			continue // Skip sender
		}

		select {
		case client.Send <- msgBytes:
		default:
			// Client's send channel is full, close it
			close(client.Send)
			delete(room.Clients, client)
		}
	}
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, msgType entity.WSMessageType, payload interface{}) {
	msg, err := entity.NewWSMessage(msgType, payload)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		return
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case client.Send <- msgBytes:
	default:
		// Client's send channel is full
		log.Printf("Client %s send channel is full", client.UserID)
	}
}

// UpdateRoomState updates the game state for a room
func (h *Hub) UpdateRoomState(sessionID string, state *entity.GameStatePayload) {
	h.mu.RLock()
	room, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if exists {
		room.mu.Lock()
		room.State = state
		room.mu.Unlock()

		// Broadcast state to all clients
		h.BroadcastToRoom(sessionID, entity.WSTypeGameState, state)
	}
}

// GetRoomState gets the current game state for a room
func (h *Hub) GetRoomState(sessionID string) *entity.GameStatePayload {
	h.mu.RLock()
	room, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if !exists {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.State
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process message
		c.processMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage handles incoming WebSocket messages
func (c *Client) processMessage(message []byte) {
	var msg entity.WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		c.Hub.SendToClient(c, entity.WSTypeError, entity.ErrorPayload{
			Code:    "INVALID_MESSAGE",
			Message: "Invalid message format",
		})
		return
	}

	switch msg.Type {
	case entity.WSTypeHeartbeat:
		// Send ACK
		c.Hub.SendToClient(c, entity.WSTypeAck, map[string]string{"type": "heartbeat"})

	case entity.WSTypeUnitPlaced:
		var payload entity.UnitActionPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshaling unit placed: %v", err)
			return
		}
		// Broadcast to other players
		c.Hub.broadcastToRoom(c.SessionID, entity.WSTypeUnitPlaced, payload, c)

	case entity.WSTypeUnitRemoved:
		var payload entity.UnitActionPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshaling unit removed: %v", err)
			return
		}
		// Broadcast to other players
		c.Hub.broadcastToRoom(c.SessionID, entity.WSTypeUnitRemoved, payload, c)

	case entity.WSTypeGameStart:
		// Broadcast game start to all players
		c.Hub.BroadcastToRoom(c.SessionID, entity.WSTypeGameStart, map[string]string{"initiator": c.UserID})

	case entity.WSTypeGamePause:
		// Broadcast pause to all players
		c.Hub.BroadcastToRoom(c.SessionID, entity.WSTypeGamePause, map[string]string{"initiator": c.UserID})

	case entity.WSTypeGameResume:
		// Broadcast resume to all players
		c.Hub.BroadcastToRoom(c.SessionID, entity.WSTypeGameResume, map[string]string{"initiator": c.UserID})

	default:
		log.Printf("Unknown message type: %s", msg.Type)
		c.Hub.SendToClient(c, entity.WSTypeError, entity.ErrorPayload{
			Code:    "UNKNOWN_TYPE",
			Message: fmt.Sprintf("Unknown message type: %s", msg.Type),
		})
	}
}