package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	cases := []struct {
		name    string
		msgType MessageType
		payload any
	}{
		{"chat", ChatMessage, ChatPayload{Body: "hello"}},
		{"broadcast", BroadcastMessage, BroadcastPayload{From: id, Body: "hi"}},
		{"dm", DMsMessage, DMsPayload{To: id, Body: "secret"}},
		{"whisper", WhisperMessage, WhisperPayload{From: id, Body: "psst"}},
		{"register", RegisterMessage, RegisterPayload{Name: "alice"}},
		{"join", JoinMessage, JoinPayload{Id: id, Name: "bob"}},
		{"leave", LeaveMessage, LeavePayload{From: id}},
		{"error", ErrorMessage, ErrorPayload{Code: 42, Body: "oops"}},
		{"ping", PingMessage, PingPayload{}},
		{"pong", PongMessage, PongPayload{}},
		{"file", FileMessage, FilePayload{Name: "photo.png", Data: "base64data"}},
		{"file_share", FileShareMessage, FileSharePayload{From: id, Name: "photo.png", Size: 1234, Data: "base64data"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := Encode(tc.msgType, tc.payload)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			gotType, data, err := DecodeEnvelope(raw)
			if err != nil {
				t.Fatalf("DecodeEnvelope: %v", err)
			}
			if gotType != tc.msgType {
				t.Errorf("type: got %q, want %q", gotType, tc.msgType)
			}
			if len(data) == 0 {
				t.Error("data is empty")
			}
		})
	}
}

func TestDecodePayloadChat(t *testing.T) {
	want := ChatPayload{Body: "hello world"}
	raw, _ := Encode(ChatMessage, want)
	_, data, _ := DecodeEnvelope(raw)
	got, err := DecodePayload[ChatPayload](data)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if got.Body != want.Body {
		t.Errorf("Body: got %q, want %q", got.Body, want.Body)
	}
}

func TestDecodePayloadBroadcast(t *testing.T) {
	id := uuid.New()
	want := BroadcastPayload{From: id, Body: "hey everyone"}
	raw, _ := Encode(BroadcastMessage, want)
	_, data, _ := DecodeEnvelope(raw)
	got, err := DecodePayload[BroadcastPayload](data)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if got.From != want.From || got.Body != want.Body {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestDecodePayloadFileShare(t *testing.T) {
	id := uuid.New()
	want := FileSharePayload{From: id, Name: "test.txt", Size: 300, Data: "abc123=="}
	raw, _ := Encode(FileShareMessage, want)
	_, data, _ := DecodeEnvelope(raw)
	got, err := DecodePayload[FileSharePayload](data)
	if err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if got.From != want.From || got.Name != want.Name || got.Size != want.Size || got.Data != want.Data {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestDecodeEnvelopeVersionMismatch(t *testing.T) {
	env := Envelope{
		Type:    ChatMessage,
		Version: ProtocolVersion + 1,
		Data:    json.RawMessage(`{}`),
	}
	raw, _ := json.Marshal(env)
	_, _, err := DecodeEnvelope(raw)
	if err == nil {
		t.Error("expected error for version mismatch, got nil")
	}
}

func TestDecodeEnvelopeInvalidJSON(t *testing.T) {
	_, _, err := DecodeEnvelope([]byte("not json {{{"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDecodePayloadInvalidJSON(t *testing.T) {
	_, err := DecodePayload[ChatPayload](json.RawMessage(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestEnvelopeCarriesCorrectVersion(t *testing.T) {
	raw, err := Encode(PingMessage, PingPayload{})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if env.Version != ProtocolVersion {
		t.Errorf("version: got %d, want %d", env.Version, ProtocolVersion)
	}
}
