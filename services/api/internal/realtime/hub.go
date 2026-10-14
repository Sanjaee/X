package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Origin checked in router/CORS layer
	},
}

// InboundEvent represents incoming client WebSocket messages
type InboundEvent struct {
	EventType string          `json:"event_type"`
	Type      string          `json:"type"` // Fallback alias
	Payload   json.RawMessage `json:"payload"`
}

// OutboundEvent represents outgoing server WebSocket events
type OutboundEvent struct {
	EventID    string      `json:"event_id"`
	EventType  string      `json:"event_type"`
	Type       string      `json:"type"` // Dual compatibility for frontend
	OccurredAt string      `json:"occurred_at"`
	Payload    interface{} `json:"payload"`
}

// Hub maintains the set of active clients and handles routing of events
type Hub struct {
	mu           sync.RWMutex
	clients      map[uuid.UUID]map[*Client]bool
	rooms        map[uuid.UUID]map[*Client]bool // conversation subscriptions
	register     chan *Client
	unregister   chan *Client
	rdb          *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rdb:        rdb,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; !ok {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true

			// Set presence in Redis if available
			if h.rdb != nil {
				_ = h.rdb.Set(ctx, fmt.Sprintf("presence:user:%s", client.userID), "online", 60*time.Second).Err()
			}
			h.mu.Unlock()

			slog.Info("ws_client_connected", "user_id", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if userClients, ok := h.clients[client.userID]; ok {
				delete(userClients, client)
				if len(userClients) == 0 {
					delete(h.clients, client.userID)
					if h.rdb != nil {
						_ = h.rdb.Del(ctx, fmt.Sprintf("presence:user:%s", client.userID)).Err()
					}
				}
			}

			// Clean up conversation room memberships
			for convID, members := range h.rooms {
				delete(members, client)
				if len(members) == 0 {
					delete(h.rooms, convID)
				}
			}
			close(client.send)
			h.mu.Unlock()

			slog.Info("ws_client_disconnected", "user_id", client.userID)
		}
	}
}

// SubscribeToConversation subscribes a client to a conversation room
func (h *Hub) SubscribeToConversation(client *Client, convID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[convID]; !ok {
		h.rooms[convID] = make(map[*Client]bool)
	}
	h.rooms[convID][client] = true
}

// UnsubscribeFromConversation removes a client from a conversation room
func (h *Hub) UnsubscribeFromConversation(client *Client, convID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if members, ok := h.rooms[convID]; ok {
		delete(members, client)
		if len(members) == 0 {
			delete(h.rooms, convID)
		}
	}
}

// BroadcastToUser sends an event to all active sessions of a user
func (h *Hub) BroadcastToUser(userID uuid.UUID, eventType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userClients, ok := h.clients[userID]
	if !ok || len(userClients) == 0 {
		return
	}

	event := OutboundEvent{
		EventID:    uuid.New().String(),
		EventType:  eventType,
		Type:       eventType,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		Payload:    payload,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	for client := range userClients {
		select {
		case client.send <- bytes:
		default:
			// Buffer full, close and let unregister handle it
			close(client.send)
			delete(userClients, client)
		}
	}
}

// BroadcastToConversation broadcasts an event to all subscribed members of a conversation
func (h *Hub) BroadcastToConversation(convID uuid.UUID, excludeUser *uuid.UUID, eventType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomMembers, ok := h.rooms[convID]
	if !ok || len(roomMembers) == 0 {
		return
	}

	event := OutboundEvent{
		EventID:    uuid.New().String(),
		EventType:  eventType,
		Type:       eventType,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		Payload:    payload,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	for client := range roomMembers {
		if excludeUser != nil && client.userID == *excludeUser {
			continue
		}
		select {
		case client.send <- bytes:
		default:
		}
	}
}

// BroadcastMessage dispatches a new message to conversation subscribers AND direct user sessions without duplicates
func (h *Hub) BroadcastMessage(convID uuid.UUID, recipientID *uuid.UUID, senderID *uuid.UUID, msg interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	event := OutboundEvent{
		EventID:    uuid.New().String(),
		EventType:  "message.created",
		Type:       "message.created",
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		Payload:    msg,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	sentClients := make(map[*Client]bool)

	// 1. Send to all clients subscribed to this conversation room
	if roomMembers, ok := h.rooms[convID]; ok {
		for client := range roomMembers {
			if !sentClients[client] {
				sentClients[client] = true
				select {
				case client.send <- bytes:
				default:
				}
			}
		}
	}

	// 2. Send to recipient user's active sessions (for notification/badge when outside conversation)
	if recipientID != nil {
		if userClients, ok := h.clients[*recipientID]; ok {
			for client := range userClients {
				if !sentClients[client] {
					sentClients[client] = true
					select {
					case client.send <- bytes:
					default:
					}
				}
			}
		}
	}

	// 3. Send to sender user's active sessions (multi-tab sync)
	if senderID != nil {
		if userClients, ok := h.clients[*senderID]; ok {
			for client := range userClients {
				if !sentClients[client] {
					sentClients[client] = true
					select {
					case client.send <- bytes:
					default:
					}
				}
			}
		}
	}
}

// Client represents a single connected WebSocket session
type Client struct {
	hub    *Hub
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(65536)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var event InboundEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		c.handleInboundEvent(event)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Drain queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleInboundEvent(event InboundEvent) {
	evtType := event.EventType
	if evtType == "" {
		evtType = event.Type
	}

	switch evtType {
	case "ping":
		pongEvent := OutboundEvent{
			EventID:    uuid.New().String(),
			EventType:  "pong",
			Type:       "pong",
			OccurredAt: time.Now().UTC().Format(time.RFC3339),
			Payload:    map[string]string{"status": "ok"},
		}
		bytes, _ := json.Marshal(pongEvent)
		c.send <- bytes

	case "conversation.subscribe", "subscribe":
		var p struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(event.Payload, &p); err == nil {
			if convUUID, err := uuid.Parse(p.ConversationID); err == nil {
				c.hub.SubscribeToConversation(c, convUUID)
			}
		}

	case "conversation.unsubscribe", "unsubscribe":
		var p struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(event.Payload, &p); err == nil {
			if convUUID, err := uuid.Parse(p.ConversationID); err == nil {
				c.hub.UnsubscribeFromConversation(c, convUUID)
			}
		}

	case "typing.start", "typing":
		var p struct {
			ConversationID string `json:"conversation_id"`
			IsTyping       bool   `json:"is_typing"`
		}
		if err := json.Unmarshal(event.Payload, &p); err == nil {
			if convUUID, err := uuid.Parse(p.ConversationID); err == nil {
				isTyping := true
				if evtType == "typing" {
					isTyping = p.IsTyping
				}

				if c.hub.rdb != nil {
					key := fmt.Sprintf("typing:conversation:%s:user:%s", convUUID, c.userID)
					if isTyping {
						_ = c.hub.rdb.Set(context.Background(), key, "1", 4*time.Second).Err()
					} else {
						_ = c.hub.rdb.Del(context.Background(), key).Err()
					}
				}
				c.hub.BroadcastToConversation(convUUID, &c.userID, "typing.updated", map[string]interface{}{
					"conversation_id": convUUID.String(),
					"user_id":         c.userID.String(),
					"is_typing":       isTyping,
				})
			}
		}

	case "typing.stop":
		var p struct {
			ConversationID string `json:"conversation_id"`
		}
		if err := json.Unmarshal(event.Payload, &p); err == nil {
			if convUUID, err := uuid.Parse(p.ConversationID); err == nil {
				if c.hub.rdb != nil {
					key := fmt.Sprintf("typing:conversation:%s:user:%s", convUUID, c.userID)
					_ = c.hub.rdb.Del(context.Background(), key).Err()
				}
				c.hub.BroadcastToConversation(convUUID, &c.userID, "typing.updated", map[string]interface{}{
					"conversation_id": convUUID.String(),
					"user_id":         c.userID.String(),
					"is_typing":       false,
				})
			}
		}
	}
}

// ServeWS handles WebSocket upgrade requests from authenticated HTTP clients
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws_upgrade_failed", "error", err)
		return
	}

	client := &Client{
		hub:    hub,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}
