package server_tui

import (
	"time"

	"charm.land/bubbles/v2/viewport"
	
	lolserver "github.com/gitmobkab/lol/internal/server"
)

type serverTickMsg struct{}

type ServerModel struct {
	viewport  viewport.Model
	logs      []string
	events    <-chan lolserver.Event
	clients   int
	startTime time.Time
	width     int
	ready     bool
}