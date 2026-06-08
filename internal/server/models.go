package server

import (
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Client struct {
	conn *websocket.Conn
	id   uuid.UUID
	name string
}

type EventType string

const (
	ClientConnected    EventType = "connected"
	ClientDisconnected EventType = "disconnected"
)

type Event struct {
	Type EventType
	ID   uuid.UUID
	Name string
	Time time.Time
}

type Server struct {
	clients map[uuid.UUID]*Client
	mutex   sync.RWMutex
	logger  *slog.Logger
	Events  chan Event
}
