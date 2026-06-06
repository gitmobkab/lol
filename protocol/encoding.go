package protocol

import (
	"encoding/json"
    "fmt"
)

func Encode(msgType MessageType, payload any) ([]byte, error) {
    data, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    return json.Marshal(Envelope{Type: msgType, version: ProtocolVersion, Data: data})
}

func DecodeEnvelope(raw []byte) (MessageType, json.RawMessage, error) {
    var env Envelope
    if err := json.Unmarshal(raw, &env); err != nil {
        return "", nil, err
    }
    if env.version != ProtocolVersion {
        return "", nil, fmt.Errorf("incompatible protocol version, expected %d, got %d", ProtocolVersion, env.version)
    }
    return env.Type, env.Data, nil
}

func DecodePayload[T Payload](data json.RawMessage) (T, error) {
    var payload T
    err := json.Unmarshal(data, &payload)
    return payload, err
}