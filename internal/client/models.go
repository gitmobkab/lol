package client

import (
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	
	"github.com/gitmobkab/lol/internal/protocol"
)


type Event struct {
	Type    protocol.MessageType
	Payload any 
}

type Client struct {
	conn    *websocket.Conn
	Self    uuid.UUID
	Name    string
	mu      sync.RWMutex
	members []protocol.Member
	Events  chan Event
	logger  *slog.Logger
	fileMu  sync.Mutex
	files   map[string][]byte // keyed by sanitized basename
}