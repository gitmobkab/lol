package client_tui

import (
	"context"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	lolclient "github.com/gitmobkab/lol/internal/client"
	"github.com/gitmobkab/lol/internal/protocol"
)

type msgKind int

const (
	kindBroadcast msgKind = iota
	kindSystem
	kindDM
)

type connClosedMsg struct{}

type message struct {
	sender string
	body   string
	ts     string
	own    bool
	kind   msgKind
}

type ClientModel struct {
	viewport viewport.Model
	input    textinput.Model
	messages []message
	members  []protocol.Member
	client   *lolclient.Client
	ctx      context.Context
	width    int
	height   int
	ready    bool
}
