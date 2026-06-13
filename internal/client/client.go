package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/x/ansi"
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
		Name:   name,
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
	for i := range syncPayload.Members {
		syncPayload.Members[i].Name = ansi.Strip(syncPayload.Members[i].Name)
	}
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

		case protocol.PongMessage:
			client.Events <- Event{Type: msgType, Payload: protocol.PongPayload{}}

		case protocol.FileShareMessage:
			p, err := protocol.DecodePayload[protocol.FileSharePayload](data)
			if err != nil {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(p.Data)
			if err != nil {
				client.logger.Error("failed to decode file data", "err", err)
				continue
			}
			name := filepath.Base(p.Name)
			client.storeFile(name, decoded)
			client.Events <- Event{Type: msgType, Payload: protocol.FileSharePayload{
				From: p.From,
				Name: name,
				Size: int64(len(decoded)),
			}}

		default:
			client.logger.Warn("unexpected message type", "type", msgType)
		}
	}
}

func (client *Client) storeFile(name string, data []byte) {
	client.fileMu.Lock()
	defer client.fileMu.Unlock()
	if client.files == nil {
		client.files = make(map[string][]byte)
	}
	client.files[name] = data
}

func (client *Client) SaveFile(name, dir string) error {
	client.fileMu.Lock()
	data, ok := client.files[name]
	client.fileMu.Unlock()
	if !ok {
		return fmt.Errorf("no buffered file %q — has it been received?", name)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, filepath.Base(name)), data, 0644)
}

const maxFileSize = 10 * 1024 * 1024 // 10 MB

func (client *Client) SendFile(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) > maxFileSize {
		return fmt.Errorf("file too large (max 10 MB)")
	}
	encoded, err := protocol.Encode(protocol.FileMessage, protocol.FilePayload{
		Name: filepath.Base(path),
		Data: base64.StdEncoding.EncodeToString(data),
	})
	if err != nil {
		return err
	}
	return client.conn.Write(ctx, websocket.MessageText, encoded)
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

func (client *Client) SendPing(ctx context.Context) error {
	encoded, err := protocol.Encode(protocol.PingMessage, protocol.PingPayload{})
	if err != nil {
		return err
	}
	return client.conn.Write(ctx, websocket.MessageText, encoded)
}

func (client *Client) BufferedFileNames() []string {
	client.fileMu.Lock()
	defer client.fileMu.Unlock()
	names := make([]string, 0, len(client.files))
	for name := range client.files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (client *Client) Close() {
	client.conn.Close(websocket.StatusNormalClosure, "")
}
