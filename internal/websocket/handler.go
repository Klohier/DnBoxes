package websocket

import (
	"context"
	"dango/internal/session"
	"dango/internal/user"
	"encoding/json"
	"errors"
	"log/slog"

	// "sync"

	"github.com/redis/go-redis/v9"
)



var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

// Manager holds connections and Events possible
type Manager struct {
	connections ConnectionList
	rooms       map[string]ConnectionList
	register  chan *Connection
	unregister chan *Connection
	broadcast   chan BroadcastEvent
	userMap        map[int]*Connection  // maps userID → connection
	userService    *user.UserService
	sessionService *session.SessionService
	handlers       map[string]EventHandler
	redisClient    *redis.Client
	deps           *HandlerDeps
}

func NewManager(UserService *user.UserService, SessionService *session.SessionService, RedisClient *redis.Client, deps *HandlerDeps) *Manager {
	m := &Manager{
		connections:    make(ConnectionList),
		userMap:     make(map[int]*Connection),
		rooms:          make(map[string]ConnectionList),
		handlers:       make(map[string]EventHandler),
		register: make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:   make(chan BroadcastEvent),
		deps:           deps,
		userService:    UserService,
		sessionService: SessionService,
		redisClient:    RedisClient,
	}
	m.setupEventHandlers(deps)
	go m.listenRedis() // start Redis subscription
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
	// m.handlers[EventJoinPage] = HandlePageJoin
	// m.handlers[EventLeavePage] = HandlePageLeave
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


func (m *Manager) Run() {
    for {
        select {
        case conn := <-m.register:
            m.connections[conn] = true
			m.userMap[conn.userID] = conn
        case conn := <-m.unregister:
            delete(m.connections, conn)
			delete(m.userMap, conn.userID)
            m.UnsubscribeAll(conn)
        case msg := <-m.broadcast:
            if conns, ok := m.rooms[msg.Topic]; ok {
                for conn := range conns {
                    conn.Send(msg.Event)
                }
            }
        }
    }
}

func (m *Manager) GetConnectionByUserID(userID int) *Connection {
	return m.userMap[userID]
}

func (m *Manager) Broadcast(topic string, eventType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal broadcast payload", "error", err)
		return
	}

	m.broadcast <- BroadcastEvent{
		Topic: topic,
		Event: Event{
			Type:    eventType,
			Payload: data,
		},
	}
}



func (m *Manager) Subscribe(topic string, conn *Connection) {
	if m.rooms[topic] == nil {
		m.rooms[topic] = make(ConnectionList)
	}
	m.rooms[topic][conn] = true
	slog.Info("Subscribed", "topic", topic, "userID", conn.userID)
}

func (m *Manager) Unsubscribe(topic string, conn *Connection) {
	slog.Info("Connection unsubscribed from topic", "topic", topic, "userID", conn.userID)
	if conns, ok := m.rooms[topic]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
				delete(m.rooms, topic)
			}
		}
	}


func (m *Manager) UnsubscribeAll(conn *Connection) {
	for topic := range m.rooms {
		m.Unsubscribe(topic, conn)
	}
}

func (m *Manager) listenRedis() {
	ctx := context.Background()
	pubsub := m.redisClient.PSubscribe(ctx, "game:*") // subscribe to all game channels
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			pubsub.Close()
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			// Decode Redis message into Event
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				slog.Error("failed to unmarshal redis event", "error", err)
				continue
			}
slog.Info("Redis message received", "channel", msg.Channel, "payload", msg.Payload)

			// Broadcast to all local connections subscribed to this topic
			m.broadcast <- BroadcastEvent{
				Topic: msg.Channel,
				Event: event,
			}
		}
	}
}