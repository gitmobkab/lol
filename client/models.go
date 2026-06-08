package client

import (
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	
	"github.com/gitmobkab/lol/protocol"
)


type Event struct {
	Type    protocol.MessageType
	Payload any 
}

type Client struct {
	conn    *websocket.Conn
	Self    uuid.UUID
	mu      sync.RWMutex
	members []protocol.Member
	Events  chan Event
	logger  *slog.Logger
}