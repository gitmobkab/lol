package server

import (
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type Client struct {
	conn    *websocket.Conn
	id      uuid.UUID
	name    string
	limiter *rate.Limiter
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
	clients      map[uuid.UUID]*Client
	mutex        sync.RWMutex
	connLimiters sync.Map // map[string]*rate.Limiter, keyed by IP
	logger       *slog.Logger
	Events       chan Event
}
