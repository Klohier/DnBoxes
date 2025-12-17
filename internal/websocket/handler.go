package websocket

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"sync"
)



var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

type subscribeRequest struct {
	topic string
	conn  *Connection
}

// Manager holds connections and Events possible
type Manager struct {
	connections ConnectionList
	rooms       map[string]ConnectionList
	userConns   map[int]*Connection
	register  chan *Connection
	unregister chan *Connection
	broadcast   chan BroadcastEvent
	eventBus  events.EventBus
	subscribe   chan subscribeRequest
	mu          sync.RWMutex 
}

func NewManager(eventBus events.EventBus) *Manager {
	m := &Manager{
		connections:    make(ConnectionList),
		userConns:   make(map[int]*Connection),
		rooms:          make(map[string]ConnectionList),
		eventBus:    eventBus,
		register: make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:   make(chan BroadcastEvent, 100),
		subscribe:   make(chan subscribeRequest, 100),

	}
	return m
}

func (m *Manager) Run() {
    for {
        select {
        case conn := <-m.register:
            m.connections[conn] = true
			m.userConns[conn.userID] = conn
			slog.Info("Connection registered", "userID", conn.userID, "total", len(m.connections))
			m.eventBus.Publish(context.Background(), "connections", events.Event{
        	Topic:   "connections",
        	Type:    "user_connected",
        	Payload: json.RawMessage(fmt.Sprintf(`{"user_id":%d}`, conn.userID)),
    	})

        case conn := <-m.unregister:
            delete(m.connections, conn)
			delete(m.userConns, conn.userID)
            m.UnsubscribeAll(conn)
			m.eventBus.Publish(context.Background(), "connections", events.Event{
        	Topic:   "connections",
        	Type:    "user_disconnected",
        	Payload: json.RawMessage(fmt.Sprintf(`{"user_id":%d}`, conn.userID)),
    	})
		case req := <-m.subscribe:
			if m.rooms[req.topic] == nil {
				m.rooms[req.topic] = make(ConnectionList)
				slog.Info("Created new topic room", "topic", req.topic)
				go m.subscribeToEventBus(req.topic)
			}
			m.rooms[req.topic][req.conn] = true
			slog.Info("Connection subscribed to topic", 
				"topic", req.topic, 
				"userID", req.conn.userID,
				"roomSize", len(m.rooms[req.topic]))


        case msg := <-m.broadcast:
			slog.Info("Broadcasting message", 
				"topic", msg.Topic, 
				"type", msg.Type,
				"roomExists", m.rooms[msg.Topic] != nil)
            if conns, ok := m.rooms[msg.Topic]; ok {
                for conn := range conns {
                    conn.Send(msg)
                }
            } else {
				slog.Warn("No connections for topic", "topic", msg.Topic)
			}
        }
    }
}


func (m *Manager) subscribeToEventBus(topic string) {
	err := m.eventBus.Subscribe(context.Background(), topic, func(e events.Event) {
		slog.Info("Event received from EventBus", 
			"topic", e.Topic, 
			"type", e.Type)
		
		be := BroadcastEvent{
			Topic:   e.Topic,
			Type:    e.Type,
			Payload: e.Payload,
		}
		
		select {
		case m.broadcast <- be:
			slog.Info("Event queued for broadcast", "topic", e.Topic)
		default:
			slog.Warn("Broadcast channel full, dropping event", 
				"topic", topic, 
				"type", e.Type)
		}
	})
	
	if err != nil {
		slog.Error("Failed to subscribe to EventBus", "topic", topic, "error", err)
	} else {
		slog.Info("Manager subscribed to EventBus topic", "topic", topic)
	}
}


func (m *Manager) Broadcast(topic string, eventType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal broadcast payload", "error", err)
		return
	}

	m.broadcast <- BroadcastEvent{
		Topic: topic,
		Type: eventType,
		Payload: data,
	}
}

func (m *Manager) UserConnection(userID int) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.userConns[userID]
	return conn, ok
}




func (m *Manager) Subscribe(topic string, conn *Connection) {
	m.subscribe <- subscribeRequest{
		topic: topic,
		conn:  conn,
	}
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

func (m *Manager) SubscribeUser(userID int, topic string) {
	conn, ok := m.UserConnection(userID)
	if !ok {
		slog.Debug("User not connected, cannot subscribe", "userID", userID, "topic", topic)
		return
	}
	
	m.Subscribe(topic, conn)
	slog.Info("User subscribed to topic", "userID", userID, "topic", topic)
}

func (m *Manager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]int{
		"total_connections": len(m.connections),
		"total_rooms":       len(m.rooms),
		"total_users":       len(m.userConns),
	}
}