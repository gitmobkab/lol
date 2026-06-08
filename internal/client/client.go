package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/gitmobkab/lol/internal/protocol"
)


func Connect(ctx context.Context, addr, name string, logger *slog.Logger) (*Client, error) {
	conn, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:   conn,
		Events: make(chan Event, 64),
		logger: logger.With("package", "lol::client"),
	}

	encoded, err := protocol.Encode(protocol.RegisterMessage, protocol.RegisterPayload{Name: name})
	if err != nil {
		conn.Close(websocket.StatusInternalError, "")
		return nil, err
	}
	if err := conn.Write(ctx, websocket.MessageText, encoded); err != nil {
		conn.Close(websocket.StatusInternalError, "")
		return nil, err
	}

	
	_, raw, err := conn.Read(ctx)
	if err != nil {
		conn.Close(websocket.StatusInternalError, "")
		return nil, err
	}
	msgType, data, err := protocol.DecodeEnvelope(raw)
	if err != nil {
		conn.Close(websocket.StatusInternalError, "")
		return nil, err
	}
	if msgType != protocol.SyncMessage {
		conn.Close(websocket.StatusPolicyViolation, "expected sync")
		return nil, fmt.Errorf("expected sync, got %s", msgType)
	}
	syncPayload, err := protocol.DecodePayload[protocol.SyncPayload](data)
	if err != nil {
		conn.Close(websocket.StatusInternalError, "")
		return nil, err
	}
	client.Self = syncPayload.Self
	client.members = syncPayload.Members

	return client, nil
}

// Members returns a snapshot of the current member list. Safe to call concurrently.
func (client *Client) Members() []protocol.Member {
	client.mu.RLock()
	defer client.mu.RUnlock()
	out := make([]protocol.Member, len(client.members))
	copy(out, client.members)
	return out
}

// ReadLoop reads incoming messages from the server and pushes them onto Events.
// Closes Events when the connection drops. Run this in a goroutine.
func (client *Client) ReadLoop(ctx context.Context) {
	defer close(client.Events)
	for {
		_, raw, err := client.conn.Read(ctx)
		if err != nil {
			client.logger.Error("connection closed", "err", err)
			return
		}
		msgType, data, err := protocol.DecodeEnvelope(raw)
		if err != nil {
			client.logger.Error("failed to decode envelope", "err", err)
			continue
		}
		switch msgType {
		case protocol.BroadcastMessage:
			p, err := protocol.DecodePayload[protocol.BroadcastPayload](data)
			if err != nil {
				continue
			}
			client.Events <- Event{Type: msgType, Payload: p}

		case protocol.WhisperMessage:
			p, err := protocol.DecodePayload[protocol.WhisperPayload](data)
			if err != nil {
				continue
			}
			client.Events <- Event{Type: msgType, Payload: p}

		case protocol.JoinMessage:
			p, err := protocol.DecodePayload[protocol.JoinPayload](data)
			if err != nil {
				continue
			}
			client.mu.Lock()
			client.members = append(client.members, protocol.Member{Name: p.Name, ID: p.Id})
			client.mu.Unlock()
			client.Events <- Event{Type: msgType, Payload: p}

		case protocol.LeaveMessage:
			p, err := protocol.DecodePayload[protocol.LeavePayload](data)
			if err != nil {
				continue
			}
			client.mu.Lock()
			for i, m := range client.members {
				if m.ID == p.From {
					client.members = append(client.members[:i], client.members[i+1:]...)
					break
				}
			}
			client.mu.Unlock()
			client.Events <- Event{Type: msgType, Payload: p}

		case protocol.ErrorMessage:
			p, err := protocol.DecodePayload[protocol.ErrorPayload](data)
			if err != nil {
				continue
			}
			client.Events <- Event{Type: msgType, Payload: p}

		default:
			client.logger.Warn("unexpected message type", "type", msgType)
		}
	}
}

func (client *Client) SendChat(ctx context.Context, body string) error {
	encoded, err := protocol.Encode(protocol.ChatMessage, protocol.ChatPayload{Body: body})
	if err != nil {
		return err
	}
	return client.conn.Write(ctx, websocket.MessageText, encoded)
}

func (client *Client) SendDM(ctx context.Context, to uuid.UUID, body string) error {
	encoded, err := protocol.Encode(protocol.DMsMessage, protocol.DMsPayload{To: to, Body: body})
	if err != nil {
		return err
	}
	return client.conn.Write(ctx, websocket.MessageText, encoded)
}

func (client *Client) Close() {
	client.conn.Close(websocket.StatusNormalClosure, "")
}
