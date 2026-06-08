package protocol

import (
    "encoding/json"
    "github.com/google/uuid"
)

type MessageType string

const ProtocolVersion int = 1

const (
    RegisterMesage MessageType = "register"
	JoinMessage MessageType = "join"
    LeaveMessage MessageType = "leave"
    SyncMessage MessageType = "sync"
    BroadcastMessage MessageType = "msg"
    DMsMessage MessageType = "dm"
    WhisperMessage MessageType = "whisper"
	ErrorMessage MessageType = "error"
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

type MsgPayload struct {
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

type Payload interface{
	RegisterPayload |
	JoinPayload |
	LeavePayload |
	SyncPayload |
	MsgPayload |
	DMsPayload |
	WhisperPayload | 
	ErrorPayload 
}