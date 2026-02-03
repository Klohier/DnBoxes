package websocket

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second

	pingInterval = (pongWait * 9) / 10
)

type ConnectionList map[*Connection]bool

// Connection for a single websocket user
type Connection struct {
	ws       *websocket.Conn
	manager  *Manager
	egress   chan BroadcastEvent
	userID   int
	username string
}

// NewConnection creates a new WebSocket connection.
func NewConnection(ws *websocket.Conn, manager *Manager, userID int, username string) *Connection {
	return &Connection{
		ws:       ws,
		manager:  manager,
		egress:   make(chan BroadcastEvent, 100),
		userID:   userID,
		username: username,
	}
}

func (c *Connection) Send(event BroadcastEvent) {
	slog.Info("Sending event to WS", "userID", c.userID, "type", event.Type)
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("send on closed egress channel (stale connection)", "userID", c.userID)
		}
	}()
	select {
	case c.egress <- event:
		// message enqueued successfully
	default:
		// egress channel full, drop the connection
		slog.Warn("dropping connection: egress channel full", "userID", c.userID)
		c.manager.unregister <- c
	}
}

func (c *Connection) readMessage() {
	defer func() {
		c.manager.unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(512)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := c.ws.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	c.ws.SetPongHandler(c.pongHandler)
	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		_, payload, err := c.ws.ReadMessage()
		if err != nil {
			// If Connection is closed, we will Recieve an error here
			// We only want to log Strange errors, but not simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break // Break the loop to close conn & Cleanup
		}

		var event events.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("error marshalling message: %v", err)
			continue
		}

		slog.Info("got message", "message", string(payload))

		topic := event.Topic
		if topic == "" {
			topic = "global:lobbies"
		}

		// Publish chat messages to a save topic for persistence.
		// The ChatService subscribes to "chat:save" independently via the EventBus.
		if event.Type == events.EventMessage {
			saveEvent := events.Event{
				Topic:   topic,
				Type:    events.EventMessage,
				Payload: event.Payload,
			}
			if err := c.manager.eventBus.Publish(context.Background(), "chat:save", saveEvent); err != nil {
				slog.Error("failed to publish chat save event", "error", err)
			}
		}

		c.manager.eventBus.Publish(context.Background(), topic, event)
	}
}

// writeMessage is a process that listens for new messages to output to the Client
func (c *Connection) writeMessage() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.egress:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				log.Println("Egress channel closed")
				if err := c.ws.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			// Write a Regular text message to the connection
			if err := c.ws.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			slog.Info("sent message", "message", string(data))
		case <-ticker.C:
			slog.Debug("ping")
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return // return to break this goroutine triggeing cleanup
			}
		}
	}
}

func (c *Connection) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	slog.Debug("pong")
	return c.ws.SetReadDeadline(time.Now().Add(pongWait))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs handles WebSocket connections.
func (m *Manager) ServeWs(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", slog.Any("error", err))
		return err
	}

	// Extract the JWT token from Echo context
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
	}

	// Extract claims
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok || !userToken.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
	}

	// Get user ID from claims
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
	}
	userID := int(userIDFloat)

	// Get username from claims
	username, _ := claims["username"].(string)
	if username == "" {
		username = fmt.Sprintf("user_%d", userID)
	}

	// creates new connection with user info
	connection := NewConnection(ws, m, userID, username)
	slog.Info("WebSocket connection established", "userID", userID, "username", username)

	// Handle existing connections before adding new one
	connection.manager.register <- connection
	slog.Info("Connection added to manager")

	// Subscribe to personal messages and global topics
	m.Subscribe(fmt.Sprintf("user:%d", userID), connection)
	m.Subscribe("global:lobbies", connection)
	m.Subscribe("chat:global", connection)

	// go routine for read message
	go func() {
		slog.Info("Starting readMessage goroutine")
		connection.readMessage()
	}()

	// go routine for write message
	go func() {
		slog.Info("Starting writeMessage goroutine")
		connection.writeMessage()
	}()

	return nil
}
