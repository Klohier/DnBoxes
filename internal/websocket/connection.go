package websocket

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"log/slog"

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

		}

		slog.Info("got message", "message", string(payload))
		

		topic := event.Topic 
		if topic == "" {
			
			topic = "global:lobbies"
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

	//grabs full user data from database
	// user, err := m.userService.FindByID(c.Request().Context(), userID)
	// if err != nil {
	// 	slog.Error("Error querying database for user: " + err.Error())
	// 	return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user info")
	// }

	// creates new connection with user info
	connection := NewConnection(ws, m, userID, "mandy_hardcode")
	slog.Info("WebSocket connection established for UserID:", "userID", userID)

	// Handle existing connections before adding new one
	connection.manager.register <- connection
	slog.Info("Connection added to manager")

	// Subscribe to personal messages
	m.Subscribe(fmt.Sprintf("user:%d", userID), connection)
	m.Subscribe("global:lobbies", connection)
	// m.Subscribe("game:10001", connection)


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