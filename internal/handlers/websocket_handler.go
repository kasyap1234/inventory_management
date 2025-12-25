package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"agromart2/internal/common"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

type WebSocketClient struct {
	ID       string
	TenantID uuid.UUID
	UserID   uuid.UUID
	Conn     *websocket.Conn
	Send     chan WebSocketMessage
	Hub      *WebSocketHub
}

type WebSocketHub struct {
	clients    map[string]*WebSocketClient
	broadcast  chan WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]*WebSocketClient),
		broadcast:  make(chan WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("WebSocket client registered: %s (tenant: %s, user: %s)",
				client.ID, client.TenantID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Printf("WebSocket client unregistered: %s", client.ID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			// Collect clients that need to be unregistered to avoid
			// unlocking/re-locking during iteration (race condition fix)
			var toUnregister []*WebSocketClient
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, mark for unregistration
					toUnregister = append(toUnregister, client)
				}
			}
			h.mu.RUnlock()
			// Unregister clients outside the read lock to avoid deadlock
			for _, client := range toUnregister {
				h.unregister <- client
			}
		}
	}
}

func (h *WebSocketHub) BroadcastToTenant(tenantID uuid.UUID, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.TenantID == tenantID {
			select {
			case client.Send <- message:
			default:
				// Skip if channel is full
			}
		}
	}
}

func (h *WebSocketHub) BroadcastToUser(userID uuid.UUID, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				// Skip if channel is full
			}
		}
	}
}

func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg WebSocketMessage
		err := websocket.JSON.Receive(c.Conn, &msg)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("WebSocket read error for client %s: %v", c.ID, err)
			}
			break
		}

		// Handle ping/pong
		if msg.Type == "pong" {
			continue
		}

		// Handle other message types if needed
		log.Printf("Received message from client %s: %s", c.ID, msg.Type)
	}
}

func (c *WebSocketClient) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Channel closed
				return
			}

			message.Timestamp = time.Now().UTC().Format(time.RFC3339)
			err := websocket.JSON.Send(c.Conn, message)
			if err != nil {
				log.Printf("WebSocket write error for client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			// Send ping
			ping := WebSocketMessage{
				Type:      "ping",
				Data:      make(map[string]interface{}),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			err := websocket.JSON.Send(c.Conn, ping)
			if err != nil {
				log.Printf("WebSocket ping error for client %s: %v", c.ID, err)
				return
			}
		}
	}
}

type WebSocketHandlers struct {
	hub *WebSocketHub
}

func NewWebSocketHandlers(hub *WebSocketHub) *WebSocketHandlers {
	return &WebSocketHandlers{hub: hub}
}

func (h *WebSocketHandlers) HandleWebSocket(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// Get user info from context (set by JWT middleware)
		ctx := c.Request().Context()
		userID, ok := common.GetUserIDFromContext(ctx)
		if !ok {
			log.Println("WebSocket: user_id not found in context")
			return
		}

		tenantID, ok := common.GetTenantIDFromContext(ctx)
		if !ok {
			log.Println("WebSocket: tenant_id not found in context")
			return
		}

		client := &WebSocketClient{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			UserID:   userID,
			Conn:     ws,
			Send:     make(chan WebSocketMessage, 256),
			Hub:      h.hub,
		}

		h.hub.register <- client

		// Start goroutines for reading and writing
		go client.WritePump()
		client.ReadPump() // This blocks until connection closes
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

// Helper function to broadcast inventory updates
func (h *WebSocketHub) BroadcastInventoryUpdate(ctx context.Context, tenantID uuid.UUID, productID uuid.UUID, warehouseID uuid.UUID, quantity int) {
	message := WebSocketMessage{
		Type: "inventory_update",
		Data: map[string]interface{}{
			"product_id":   productID.String(),
			"warehouse_id": warehouseID.String(),
			"quantity":     quantity,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	h.BroadcastToTenant(tenantID, message)
}

// Helper function to broadcast order updates
func (h *WebSocketHub) BroadcastOrderUpdate(ctx context.Context, tenantID uuid.UUID, orderID uuid.UUID, status string) {
	message := WebSocketMessage{
		Type: "order_update",
		Data: map[string]interface{}{
			"order_id": orderID.String(),
			"status":   status,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	h.BroadcastToTenant(tenantID, message)
}

// Helper function to send low stock alerts
func (h *WebSocketHub) BroadcastLowStockAlert(ctx context.Context, tenantID uuid.UUID, productName string, currentStock int, threshold int) {
	message := WebSocketMessage{
		Type: "low_stock_alert",
		Data: map[string]interface{}{
			"product_name":  productName,
			"current_stock": currentStock,
			"threshold":     threshold,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	h.BroadcastToTenant(tenantID, message)
}
