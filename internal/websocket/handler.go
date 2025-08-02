package websocket

import (
	"context"
	"dango/internal/auth/token"
	"dango/internal/session"
	"dango/internal/user"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

// Manager holds connections and Events possible
type Manager struct {
	connections ConnectionList
	rooms       map[int]ConnectionList
	sync.RWMutex
	userService    *user.UserService
	sessionService *session.SessionService
	handlers       map[string]EventHandler
	redisClient    *redis.Client
	deps           *HandlerDeps
	subscriptions  map[string]map[*Connection]bool
}

func NewManager(UserService *user.UserService, SessionService *session.SessionService, RedisClient *redis.Client, deps *HandlerDeps) *Manager {
	m := &Manager{
		connections:    make(ConnectionList),
		rooms:          make(map[int]ConnectionList),
		handlers:       make(map[string]EventHandler),
		userService:    UserService,
		sessionService: SessionService,
		redisClient:    RedisClient,
		deps:           deps,
	}
	m.setupEventHandlers(deps)
	return m
}

// setupEventHandlers is where we add different Events
func (m *Manager) setupEventHandlers(deps *HandlerDeps) {
	m.handlers[EventMessage] = func(e Event, c *Connection) error {
		return MessageHandler(e, c, deps)
	}

	m.handlers[EventGameState] = func(e Event, c *Connection) error {
		return GameStateHandler(e, c, deps)
	}

	m.handlers[EventMakeMove] = func(e Event, c *Connection) error {
		return MoveHandler(e, c, deps)
	}

	m.handlers[EventQuitGame] = func(e Event, c *Connection) error {
		return QuitGameHandler(e, c, deps)
	}
	m.handlers[EventAcceptInvite] = func(e Event, c *Connection) error {
		return AcceptInviteHandler(e, c, deps)
	}

	m.handlers[EventGetPlayers] = PlayerHandler
	m.handlers[EventSendInvite] = InviteHandler
	m.handlers[EventJoinPage] = HandlePageJoin
	m.handlers[EventLeavePage] = HandlePageLeave
}

// routeEvent is how we send events to proper handler
func (m *Manager) routeEvent(event Event, c *Connection) error {
	// Check if Handler is present in Map
	if handler, ok := m.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

// ServeWs handles WebSocket connections.
func (m *Manager) ServeWs(c echo.Context) error {

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", slog.Any("error", err))
		return err
	}
	//grabs user data from session
	cookie, err := c.Cookie("DnB-Session")
	if err != nil {
		slog.Error("Error getting session from cookie: " + err.Error())
		return echo.NewHTTPError(http.StatusUnauthorized, "Session not found in cookie")
	}
	userID, err := token.VerifyToken(cookie.Value)

	if err != nil {
		slog.Error("Error decoding the cookie:", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session token")
	}
	//grabs full user data from datanase
	user, err := m.userService.FindByID(c.Request().Context(), userID)
	if err != nil {
		slog.Error("Error querying database for user: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user info")
	}
	// creates new connection with user info
	connection := NewConnection(ws, m, userID, user.Username)
	slog.Info("WebSocket connection established for UserID:", "userID", userID)
	m.addConnection(connection)
	slog.Info("Connection added to manager")

	// Subscribe to personal messages
	m.Subscribe(fmt.Sprintf("user:%d", userID), connection)

	m.Subscribe("lobby", connection)

	// go routine for read message
	go func() {
		slog.Info("Starting readMessage goroutine")
		defer m.cleanupConnection(connection)
		connection.readMessage()

	}()

	// go routine for write message
	go func() {
		slog.Info("Starting writeMessage goroutine")
		defer m.cleanupConnection(connection)
		connection.writeMessage()
	}()

	return nil
}

func (m *Manager) Subscribe(topic string, conn *Connection) {
	m.Lock()

	// Init the map if needed
	if m.subscriptions == nil {
		m.subscriptions = make(map[string]map[*Connection]bool)
	}

	if m.subscriptions[topic] == nil {
		m.subscriptions[topic] = make(map[*Connection]bool)

		// Start Redis listener for topic
		go m.SubscribeToRedis(topic)
	}

	m.subscriptions[topic][conn] = true
	m.Unlock()

}

func (m *Manager) Unsubscribe(topic string, conn *Connection) {
	m.Lock()

	if conns, ok := m.subscriptions[topic]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
		}
	}
	m.Unlock()

}

// TODO: Remove Soon
func (m *Manager) broadcastPlayers(topic string) {
	m.RLock()
	conns, ok := m.subscriptions[topic]
	m.RUnlock()
	if !ok {
		return
	}

	var players []Player
	for conn := range conns {
		players = append(players, Player{
			UserID:   conn.userID,
			Username: conn.username,
		})
	}

	responsePayload, err := json.Marshal(players)
	if err != nil {
		slog.Error("failed to marshal players", "err", err)
		return
	}

	responseEvent := Event{
		Type:    EventGetPlayers,
		Payload: responsePayload,
	}

	for conn := range conns {
		select {
		case conn.egress <- responseEvent:
		default:
			slog.Warn("dropped player update for client", "userID", conn.userID)
		}
	}
}

func (m *Manager) broadcast(topic string, eventType string, data any) {
	m.RLock()
	conns, ok := m.subscriptions[topic]
	m.RUnlock()
	if !ok {
		return
	}

	responsePayload, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal broadcast data", "topic", topic, "err", err)
		return
	}

	responseEvent := Event{
		Type:    eventType,
		Payload: responsePayload,
	}

	for conn := range conns {
		select {
		case conn.egress <- responseEvent:
		default:
			slog.Warn("dropped message to client", "userID", conn.userID, "eventType", eventType)
		}
	}
}

// cleanupConnection closes websocket connection and removes from manager
func (m *Manager) cleanupConnection(connection *Connection) {
	slog.Info("Closing WebSocket connection")

	m.removeConnection(connection)

	connection.ws.Close()

	slog.Info("WebSocket connection closed", "userID", connection.userID)
}

// addConnection adds new connection and broadcast updated connections to connected clients
func (m *Manager) addConnection(connection *Connection) {
	m.Lock()

	var existingConn *Connection
	for conn := range m.connections {
		if conn.userID == connection.userID {
			existingConn = conn
			break
		}
	}
	m.Unlock()

	if existingConn != nil {
		slog.Info("Closing existing connection for UserID:", "userID", existingConn.userID)
		m.cleanupConnection(existingConn)
	}
	m.Lock()
	m.connections[connection] = true
	m.Unlock()
}

func (m *Manager) removeConnection(connection *Connection) {
	slog.Info("Removing connection", "userID", connection.userID)
	m.Lock()
	defer m.Unlock()
	delete(m.connections, connection)
	m.UnsubscribeAll(connection)
}

func findConnectionByUserID(m *Manager, userID int) *Connection {
	m.RLock()
	defer m.RUnlock()

	for client := range m.connections {
		if client.userID == userID {
			return client
		}
	}
	return nil
}

func (m *Manager) SubscribeToRedis(topic string) {
	ctx := context.Background()
	slog.Info("Started Redis subscription for topic:", "topic", topic)

	pubsub := m.redisClient.Subscribe(ctx, topic)
	ch := pubsub.Channel()

	for msg := range ch {

		var event Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			slog.Error("Failed to unmarshal Redis event:", "error ", err)
			continue
		}

		m.RLock()
		conns := m.subscriptions[topic]
		for conn := range conns {
			select {
			case conn.egress <- event:
			default:
				slog.Warn("Egress full, dropping event for user:", "userID", conn.userID)
			}
		}
		m.RUnlock()

	}
}

func (m *Manager) UnsubscribeAll(c *Connection) {

	for topic, conns := range m.subscriptions {
		delete(conns, c)
		if len(conns) == 0 {
			delete(m.subscriptions, topic)
		}
	}

}
