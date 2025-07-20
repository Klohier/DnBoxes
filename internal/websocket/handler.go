package websocket

import (
	"context"
	"dango/internal/chat"
	"dango/internal/game"
	"dango/internal/session"
	"dango/internal/token"
	"dango/internal/user"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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
	gameService    *game.GameService
	userService    *user.UserService
	chatService    *chat.ChatService
	sessionService *session.SessionService
	handlers       map[string]EventHandler
}

func NewManager(GameService *game.GameService, UserService *user.UserService, ChatService *chat.ChatService, SessionService *session.SessionService) *Manager {
	m := &Manager{
		connections:    make(ConnectionList),
		rooms:          make(map[int]ConnectionList),
		handlers:       make(map[string]EventHandler),
		gameService:    GameService,
		userService:    UserService,
		chatService:    ChatService,
		sessionService: SessionService,
	}
	m.setupEventHandlers()
	return m
}

// setupEventHandlers is where we add different Events
func (m *Manager) setupEventHandlers() {
	m.handlers[EventMessage] = MessageHandler

	m.handlers[EventGameState] = GameStateHandler
	m.handlers[EventGetPlayers] = PlayerHandler
	m.handlers[EventSendInvite] = InviteHandler
	m.handlers[EventAcceptInvite] = AcceptInviteHandler
	m.handlers[EventMakeMove] = MoveHandler
	m.handlers[EventQuitGame] = QuitGameHandler
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
		log.Printf("Error decoding the cookie: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session token")
	}

	//grabs full user data from datanase
	user, err := m.userService.FindByID(c.Request().Context(), userID)
	if err != nil {
		slog.Error("Error querying database for user: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user info")
	}

	sessionID, err := m.sessionService.FindSessionByUserID(c.Request().Context(), userID)
	if err != nil {
		slog.Error("Error querying session for user: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user session")
	}

	// If no active session, assign to default lobby (ID 1)
	finalSessionID := 1
	if sessionID != nil {
		finalSessionID = *sessionID
	}

	// creates new connection with user info
	connection := NewConnection(ws, m, userID, user.Username, finalSessionID)
	slog.Info(fmt.Sprintf("WebSocket connection established for UserID=%d, SessionID=%d", userID, finalSessionID))
	m.addConnection(connection)
	m.JoinRoom(connection, finalSessionID)
	log.Println("Connection added to manager")

	// go routine for read message
	go func() {
		log.Println("Starting readMessage goroutine")
		defer m.cleanupConnection(connection)
		connection.readMessage()

	}()

	// go routine for write message
	go func() {
		log.Println("Starting writeMessage goroutine")
		defer m.cleanupConnection(connection)
		connection.writeMessage()
		// m.cleanupConnection(connection)
	}()

	return nil
}

// cleanupConnection closes websocket connection and removes from manager
func (m *Manager) cleanupConnection(connection *Connection) {
	log.Println("Closing WebSocket connection")

	m.LeaveRoom(connection)

	m.removeConnection(connection)

	connection.ws.Close()

	log.Printf("WebSocket connection closed for UserID=%d", connection.userID)
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
		log.Printf("Closing existing connection for UserID=%d", existingConn.userID)
		m.cleanupConnection(existingConn) // safe to run without deadlock
	}
	m.Lock()
	m.connections[connection] = true
	m.Unlock()

}

func (m *Manager) removeConnection(connection *Connection) {
	m.Lock()
	defer m.Unlock()
	delete(m.connections, connection)
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

func (m *Manager) BroadcastToRoom(event Event, room int) error {
	m.RLock()

	//Check if room exists
	conns, ok := m.rooms[room]
	m.RUnlock()

	if !ok {
		return fmt.Errorf("room %d does not exist", room)
	}

	// Collect the list of players in this room
	for client := range conns {
		select {
		case client.egress <- event:
		default:
			slog.Warn("Dropping message: egress channel full", "userID", client.userID, "room", room)
		}
	}

	return nil
}


func (m *Manager) BroadcastPlayerListToRoom(sessionID int) error {
	m.RLock()
	conns, ok := m.rooms[sessionID]
	m.RUnlock()

	if !ok {
		return fmt.Errorf("room %d does not exist", sessionID)
	}

	var players []Player
	for client := range conns {
		players = append(players, Player{
			UserID:   client.userID,
			Username: client.username,
		})
	}

	payload, err := json.Marshal(players)
	if err != nil {
		return fmt.Errorf("failed to marshal player list: %v", err)
	}

	event := Event{
		Type:    EventGetPlayers,
		Payload: payload,
	}

	return m.BroadcastToRoom(event, sessionID)
}



func (m *Manager) JoinRoom(connection *Connection, sessionID int) {
	m.Lock()

	//  if connection.sessionID == sessionID {
	//     return
	// }

	// Remove from previous session room map if any
	if connection.sessionID != 0 {
		m.Unlock()
		m.LeaveRoom(connection)
		m.Lock()
	}

	// Add connection to new session room
	if m.rooms[sessionID] == nil {
		m.rooms[sessionID] = make(ConnectionList)
	}
	m.rooms[sessionID][connection] = true

	// Update connection's current sessionID
	connection.sessionID = sessionID
	m.Unlock()

	slog.Info("User joined session", "userID", connection.userID, "sessionID", sessionID)
	if err := m.BroadcastPlayerListToRoom(sessionID); err != nil {
		log.Printf("error broadcasting player list to session %d: %v", sessionID, err)
	}

	if err := m.sessionService.AddUserToSession(context.Background(), sessionID, connection.userID); err != nil {
		slog.Error("Failed to update user session in DB", "userID", connection.userID, "sessionID", sessionID, "err", err)
	}

	if err := m.sessionService.SetUserConnectionStatus(context.Background(), sessionID, connection.userID, "connected"); err != nil {
		slog.Error("Failed to set user connected in DB", "userID", connection.userID, "sessionID", sessionID, "err", err)
	}
}

func (m *Manager) LeaveRoom(connection *Connection) {
	m.Lock()
	sessionID := connection.sessionID
	if sessionID == 0 {
		m.Unlock()
		return
	}

	if conns, ok := m.rooms[sessionID]; ok {
		// close(connection.egress)
		delete(conns, connection)
		if len(conns) == 0 && sessionID != 1 {
			delete(m.rooms, sessionID)
		}
	}

	connection.sessionID = 0
	m.Unlock()

	slog.Info("User left session", "userID", connection.userID, "sessionID", sessionID)

	if err := m.sessionService.SetUserConnectionStatus(context.Background(), sessionID, connection.userID, "disconnected"); err != nil {
		slog.Error("Failed to set user disconnected in DB", "userID", connection.userID, "sessionID", sessionID, "err", err)
	}

	slog.Info("User marked disconnected from session", "userID", connection.userID, "sessionID", sessionID)

	if err := m.BroadcastPlayerListToRoom(sessionID); err != nil {
		log.Printf("error broadcasting player list to session %d: %v", sessionID, err)
	}
}
