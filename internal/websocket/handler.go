package websocket

import (
	"dango/internal/chat"
	"dango/internal/game"
	"dango/internal/user"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
type Manager struct{
	connections ConnectionList
	sync.RWMutex
	gameService *game.GameService
	userService *user.UserService
	chatService *chat.ChatService
	handlers map[string]EventHandler
}

func NewManager(GameService *game.GameService, UserService *user.UserService, ChatService *chat.ChatService) *Manager {
	m := &Manager{
		connections: make(ConnectionList),
		handlers: make(map[string]EventHandler),
		gameService: GameService,
		userService: UserService,
		chatService: ChatService,
	}
	m.setupEventHandlers()
	return m
}

//setupEventHandlers is where we add different Events
func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler

	m.handlers[EventGetGrids] = GetGridsHandler
	m.handlers[EventGetPlayers] = GetPlayersHandler
	m.handlers[EventSendInvite] = SendInviteHandler
	m.handlers[EventAcceptInvite] = AcceptInviteHandler
	m.handlers[EventMakeMove] = MakeMoveHandler
	m.handlers[EventQuitGame] = QuitGameHandler
}

//routeEvent is how we send events to proper handler
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
		slog.Error("Error upgrading to WebSocket: ", err.Error())
		return err
	}

	//grabs user data from session
	cookie, err := c.Cookie("DnB-Session")
	if err != nil {
		slog.Error("Error getting session from cookie: " + err.Error())
		return echo.NewHTTPError(http.StatusUnauthorized, "Session not found in cookie")
	}
	decodedToken, err := base64.StdEncoding.DecodeString(cookie.Value)

	if err != nil {
		log.Printf("Error decoding the cookie: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session token")
	}

	tokenParts := strings.Split(string(decodedToken), "|")

	userIDStr := tokenParts[0]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		slog.Error("Error converting userID to int: " + err.Error())
		return err
	}

	//grabs full user data from datanase
	user, err := m.userService.FindByID(c.Request().Context(), userID)
	if err != nil {
		slog.Error("Error querying database for user: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user info")
	}

	slog.Info("WebSocket connection established")
	

	// creates new connection with user info
	connection := NewConnection(ws, m, userID, user.Username, user.GameID)
	m.addConnection(connection)
	log.Println("Connection added to manager")

	if user.GameID != nil {
		log.Println("New connection: UserID =", userID, "Username =", user.Username, "GameID =", strconv.Itoa(*user.GameID))
	} else {
		log.Println("New connection: UserID =", userID, "Username =", user.Username, "GameID = nil")
	}

	


	// go routine for read message
	go func() {
        log.Println("Starting readMessage goroutine")
        connection.readMessage()
		m.cleanupConnection(connection)
    }()

	// go routine for write message
    go func() {
        log.Println("Starting writeMessage goroutine")
        connection.writeMessage()
		m.cleanupConnection(connection)
    }()

	return nil
}

//cleanupConnection closes websocket connection and removes from manager
func (m *Manager) cleanupConnection(connection *Connection) {
    log.Println("Closing WebSocket connection")
    connection.ws.Close()
    m.removeConnection(connection)
}

//addConnection adds new connection and broadcast updated connections to connected clients
func (m *Manager) addConnection(connection *Connection){
	m.Lock()
	defer m.Unlock()

	m.connections[connection] = true
	BroadcastPlayerList(m)
	
}

func (m *Manager) removeConnection(connection *Connection) {
	m.Lock()
	defer m.Unlock()

	// Check if Client exists, then delete it
	if _, ok := m.connections[connection]; ok {
		// close connection
		connection.ws.Close()
		// remove
		delete(m.connections, connection)
		BroadcastPlayerList(m)
	}
}

func findConnectionByUserID(m *Manager, userID int) *Connection {
    for client := range m.connections {
        if client.userID == userID {
            return client
        }
    }
    return nil 
}


func BroadcastPlayerList(manager *Manager) error {
    var players []Player

    // Collect the list of players who are not in a game
    for client := range manager.connections {
        if client.gameID == nil { 
            players = append(players, Player{
                UserID:   client.userID,
                Username: client.username,
            })
        }
    }

    responsePayload, err := json.Marshal(players)
    if err != nil {
        return fmt.Errorf("failed to marshal players response: %v", err)
    }


    newPlayersEvent := Event{
        Type:    EventNewPlayers, 
        Payload: responsePayload,
    }

    for client := range manager.connections {
        client.egress <- newPlayersEvent
    }

    return nil
}