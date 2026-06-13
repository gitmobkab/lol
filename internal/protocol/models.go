package protocol

import (
    "encoding/json"
    "github.com/google/uuid"
)

type MessageType string

const ProtocolVersion int = 1

const (
    RegisterMessage  MessageType = "register"
	JoinMessage      MessageType = "join"
    LeaveMessage     MessageType = "leave"
    SyncMessage      MessageType = "sync"
    ChatMessage      MessageType = "chat"
    BroadcastMessage MessageType = "broadcast"
    DMsMessage       MessageType = "dm"
    WhisperMessage   MessageType = "whisper"
	ErrorMessage     MessageType = "error"
	PingMessage      MessageType = "ping"
	PongMessage      MessageType = "pong"
	FileMessage      MessageType = "file"
	FileShareMessage MessageType = "file_share"
)

type Envelope struct {
    Type MessageType     `json:"type"`
    Version int           `json:"version"`
    Data json.RawMessage `json:"data"`
}

type Member struct {
    Name string    `json:"name"`
    ID   uuid.UUID `json:"id"`
}

type RegisterPayload struct{
	Name string `json:"name"`
}

type JoinPayload struct {
	Id uuid.UUID `json:"id"`
    Name string `json:"name"`
}

type LeavePayload struct {
    From uuid.UUID `json:"from"`
}

type SyncPayload struct {
    Self    uuid.UUID `json:"self"`
    Members []Member  `json:"members"`
}

type ChatPayload struct {
    Body string    `json:"body"`
}

type BroadcastPayload struct {
    From uuid.UUID `json:"from"`
    Body string    `json:"body"`
}

type DMsPayload struct {
    To uuid.UUID `json:"to"`
    Body string    `json:"body"`
}

type WhisperPayload struct {
    From uuid.UUID `json:"from"`
    Body string    `json:"body"`
}

type ErrorPayload struct{
	Code int `json:"code"`
	Body any `json:"body"`
}

type PingPayload struct{}
type PongPayload struct{}

// FilePayload is sent client→server when uploading a file.
type FilePayload struct {
	Name string `json:"name"` // original filename (basename only)
	Data string `json:"data"` // base64-encoded file contents
}

// FileSharePayload is broadcast server→clients when a file is shared.
type FileSharePayload struct {
	From uuid.UUID `json:"from"`
	Name string    `json:"name"`
	Size int64     `json:"size"` // decoded byte count (approx from base64 length)
	Data string    `json:"data"` // base64-encoded file contents
}

type Payload interface{
	RegisterPayload |
	JoinPayload |
	LeavePayload |
	SyncPayload |
    ChatPayload |
	BroadcastPayload |
	DMsPayload |
	WhisperPayload |
	ErrorPayload |
	PingPayload |
	PongPayload |
	FilePayload |
	FileSharePayload
}