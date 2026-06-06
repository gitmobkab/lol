package server

import (
	"log/slog"
	"sync"
	"github.com/google/uuid"

	"github.com/coder/websocket"
)

type Client struct {
    conn *websocket.Conn
    id uuid.UUID
    name string
}

type Server struct {
    clients map[uuid.UUID]*Client
    mutex sync.RWMutex
	logger *slog.Logger
}