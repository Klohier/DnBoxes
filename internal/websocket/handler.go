package websocket

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"errors"
	"log/slog"
	// "sync"
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
	eventBus  events.EventBus
}

func NewManager(eventBus events.EventBus) *Manager {
	m := &Manager{
		connections:    make(ConnectionList),
		rooms:          make(map[string]ConnectionList),
		eventBus:    eventBus,
		register: make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:   make(chan BroadcastEvent),

	}
	return m
}

func (m *Manager) Run() {
    for {
        select {
        case conn := <-m.register:
            m.connections[conn] = true
        case conn := <-m.unregister:
            delete(m.connections, conn)
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

// func (m *Manager) GetConnectionByUserID(userID int) *Connection {
// 	return m.userMap[userID]
// }

func (m *Manager) Broadcast(topic string, eventType string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal broadcast payload", "error", err)
		return
	}

	m.broadcast <- BroadcastEvent{
		Topic: topic,
		Event: events.Event{
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

func (m *Manager) ListenEventBus(topic string) {
    // subscribe to EventBus using the handler
    err := m.eventBus.Subscribe(context.Background(), topic, func(e events.Event) {
        // Forward the event to all connections subscribed to this topic
        if conns, ok := m.rooms[topic]; ok {
            for conn := range conns {
                conn.Send(e)
            }
        }
    })

    if err != nil {
        slog.Error("failed to subscribe to EventBus", "topic", topic, "error", err)
    } else {
        slog.Info("Manager subscribed to EventBus topic", "topic", topic)
    }
}
