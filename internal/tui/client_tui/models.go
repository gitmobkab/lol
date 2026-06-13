package client_tui

import (
	"context"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	lolclient "github.com/gitmobkab/lol/internal/client"
	"github.com/gitmobkab/lol/internal/protocol"
	"github.com/gitmobkab/lol/internal/tui/message_bubble"
)

type connClosedMsg struct{}

type ClientModel struct {
	viewport   viewport.Model
	input      textarea.Model
	messages   []message_bubble.Message
	msgLines   []int // content-line offset of each message bubble's top border
	members    []protocol.Member
	client     *lolclient.Client
	ctx        context.Context
	history    []string
	histIdx    int // -1 = editing current draft
	draft      string
	width      int
	height     int
	ready      bool
	scrollMode bool
	keyMap     KeyMap
	filePicker *filepicker.Model // non-nil when the file picker overlay is open
}
