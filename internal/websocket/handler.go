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
	cancelFuncs    map[string]context.CancelFunc
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
		subscriptions:  make(map[string]map[*Connection]bool),
		cancelFuncs:    make(map[string]context.CancelFunc),
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

	//grabs full user data from database
	user, err := m.userService.FindByID(c.Request().Context(), userID)
	if err != nil {
		slog.Error("Error querying database for user: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user info")
	}

	// creates new connection with user info
	connection := NewConnection(ws, m, userID, user.Username)
	slog.Info("WebSocket connection established for UserID:", "userID", userID)

	// FIXED: Handle existing connections properly BEFORE adding new one
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
	defer m.Unlock()

	// Init the map if needed
	if m.subscriptions == nil {
		m.subscriptions = make(map[string]map[*Connection]bool)
	}

	if m.subscriptions[topic] == nil {
		m.subscriptions[topic] = make(map[*Connection]bool)

		ctx, cancel := context.WithCancel(context.Background())
		m.cancelFuncs[topic] = cancel

		// Start Redis listener for topic
		go m.SubscribeToRedis(ctx, topic)
		slog.Info("Started new Redis subscription", "topic", topic)
	}

	m.subscriptions[topic][conn] = true
	slog.Info("Connection subscribed to topic", "topic", topic, "userID", conn.userID)
}

func (m *Manager) Unsubscribe(topic string, conn *Connection) {
	m.Lock()
	defer m.Unlock()

	if conns, ok := m.subscriptions[topic]; ok {
		delete(conns, conn)
		slog.Info("Connection unsubscribed from topic", "topic", topic, "userID", conn.userID)

		if len(conns) == 0 {
			delete(m.subscriptions, topic)
			if cancel, ok := m.cancelFuncs[topic]; ok {
				cancel() // Cancel Redis subscription
				delete(m.cancelFuncs, topic)
				slog.Info("Canceled Redis subscription for topic", "topic", topic)
			}
		}
	}
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
		slog.Info("broadcast topic has no subscribers", "topic", topic)
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

	slog.Info("broadcasting message", "topic", topic, "connections", len(conns), "eventType", eventType)
	for conn := range conns {
		slog.Info("sending to connection", "userID", conn.userID)
		select {
		case conn.egress <- responseEvent:
		default:
			slog.Warn("dropped message to client", "userID", conn.userID, "eventType", eventType)
		}
	}
}

// cleanupConnection closes websocket connection and removes from manager
func (m *Manager) cleanupConnection(connection *Connection) {
	slog.Info("Cleaning up connection", "userID", connection.userID)

	// Close the websocket first
	connection.ws.Close()

	// Remove from manager (which handles unsubscriptions)
	m.removeConnection(connection)

	slog.Info("WebSocket connection closed", "userID", connection.userID)
}

// addConnection adds user connection to memory
func (m *Manager) addConnection(connection *Connection) {
	m.Lock()
	defer m.Unlock()

	// Find existing connection for this user
	var existingConn *Connection
	for conn := range m.connections {
		if conn.userID == connection.userID {
			existingConn = conn
			break
		}
	}

	// If there's an existing connection, clean it up first
	if existingConn != nil {
		slog.Info("Found existing connection for UserID, cleaning up", "userID", existingConn.userID)

		// Remove from connections map immediately
		delete(m.connections, existingConn)

		// Unsubscribe from all topics (this is safe to call within the lock)
		m.unsubscribeAllUnsafe(existingConn)

		// Close the websocket connection in a goroutine to avoid blocking
		go func() {
			existingConn.ws.Close()
			slog.Info("Closed existing WebSocket connection", "userID", existingConn.userID)
		}()
	}

	// Add the new connection
	m.connections[connection] = true
	slog.Info("Added new connection", "userID", connection.userID, "totalConnections", len(m.connections))
}

func (m *Manager) removeConnection(connection *Connection) {
	slog.Info("Removing connection", "userID", connection.userID)

	m.Lock()
	delete(m.connections, connection)
	m.Unlock()

	m.UnsubscribeAll(connection)
}

// Added unsafe version for use within existing locks
func (m *Manager) unsubscribeAllUnsafe(c *Connection) {
	for topic, conns := range m.subscriptions {
		if _, exists := conns[c]; exists {
			delete(conns, c)
			slog.Info("Unsubscribed connection from topic", "topic", topic, "userID", c.userID)

			// If no more connections for this topic, cancel Redis subscription
			if len(conns) == 0 {
				delete(m.subscriptions, topic)
				if cancel, ok := m.cancelFuncs[topic]; ok {
					cancel()
					delete(m.cancelFuncs, topic)
					slog.Info("Canceled Redis subscription for topic", "topic", topic)
				}
			}
		}
	}
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

func (m *Manager) SubscribeToRedis(ctx context.Context, topic string) {
	slog.Info("Started Redis subscription for topic:", "topic", topic)

	pubsub := m.redisClient.Subscribe(ctx, topic)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Redis subscription canceled for topic:", "topic", topic)
			return
		case msg, ok := <-ch:
			if !ok {
				slog.Info("Redis subscription channel closed for topic:", "topic", topic)
				return
			}

			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Error("Failed to unmarshal Redis event:", "error", err)
				continue
			}

			//  Pass the raw payload instead of double-marshaling
			var rawPayload json.RawMessage
			if err := json.Unmarshal(event.Payload, &rawPayload); err != nil {
				slog.Error("Failed to unmarshal event payload:", "error", err)
				continue
			}

			m.broadcast(topic, event.Type, rawPayload)
		}
	}
}

func (m *Manager) UnsubscribeAll(c *Connection) {
	m.Lock()
	defer m.Unlock()
	m.unsubscribeAllUnsafe(c)
}
