package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/JokerTrickster/joker_backend/shared/logger"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients mapped by userID
	clients map[uint]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[uint]map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]bool)
	}
	h.clients[client.userID][client] = true

	logger.Info("WebSocket client registered",
		zap.Uint("user_id", client.userID),
		zap.Int("total_connections", len(h.clients[client.userID])),
	)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.userID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			if len(clients) == 0 {
				delete(h.clients, client.userID)
			}

			logger.Info("WebSocket client unregistered",
				zap.Uint("user_id", client.userID),
			)
		}
	}
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uint, eventType string, data interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	if !ok || len(clients) == 0 {
		logger.Debug("No WebSocket clients for user", zap.Uint("user_id", userID))
		return nil
	}

	// Create message
	message := Message{
		Type: eventType,
		Data: data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send to all connections for this user
	for client := range clients {
		select {
		case client.send <- messageBytes:
		default:
			logger.Warn("Failed to send message to client",
				zap.Uint("user_id", userID),
			)
		}
	}

	logger.Debug("Message sent to user",
		zap.Uint("user_id", userID),
		zap.String("event_type", eventType),
		zap.Int("num_clients", len(clients)),
	)

	return nil
}

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// FileProcessedData represents data sent when file processing completes
type FileProcessedData struct {
	FileID       uint    `json:"file_id"`
	Status       string  `json:"status"` // "completed" or "failed"
	ThumbnailKey string  `json:"thumbnail_key,omitempty"`
	Duration     float64 `json:"duration,omitempty"`
	Error        string  `json:"error,omitempty"`
}

// NotifyFileProcessed sends a file processed notification to a user
func (h *Hub) NotifyFileProcessed(userID uint, fileID uint, status string, thumbnailKey string, duration float64, err error) {
	data := FileProcessedData{
		FileID:       fileID,
		Status:       status,
		ThumbnailKey: thumbnailKey,
		Duration:     duration,
	}

	if err != nil {
		data.Error = err.Error()
	}

	if sendErr := h.SendToUser(userID, "file:processed", data); sendErr != nil {
		logger.Error("Failed to send file processed notification",
			zap.Uint("user_id", userID),
			zap.Uint("file_id", fileID),
			zap.Error(sendErr),
		)
	}
}

// GetConnectedUsers returns the number of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetConnectionsForUser returns the number of connections for a specific user
func (h *Hub) GetConnectionsForUser(userID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}
