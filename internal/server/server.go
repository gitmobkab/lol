package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/gitmobkab/lol/internal/protocol"
	"github.com/google/uuid"
)


func New(logger *slog.Logger) *Server {
    return &Server{
        clients: make(map[uuid.UUID]*Client),
        logger:  logger.With("package", "lol::server"),
        Events:  make(chan Event, 64),
    }
}

func (server *Server) emit(e Event) {
    select {
    case server.Events <- e:
    default:
    }
}

func (server *Server) register(client *Client) {
    server.mutex.Lock()
    defer server.mutex.Unlock()
    server.clients[client.id] = client
}

func (server *Server) unregister(client *Client) {
    server.mutex.Lock()
    defer server.mutex.Unlock()
    delete(server.clients, client.id)
}

func (server *Server) GetMembers() []protocol.Member {
    server.mutex.RLock()
    defer server.mutex.RUnlock()
    members := make([]protocol.Member, 0, len(server.clients))
    for _, client := range server.clients {
        members = append(members, protocol.Member{Name: client.name, ID: client.id})
    }
    return members
}

func (server *Server) broadcast(ctx context.Context, author *Client, msgType protocol.MessageType, payload any) {
    server.mutex.RLock()
    defer server.mutex.RUnlock()
    for _, client := range server.clients {
        if client == author {
			server.logger.Debug("Skipping author from broadcast")
            continue
        }
        server.send(ctx, client, msgType, payload)
    }
}

func (server *Server) send(
	ctx context.Context, 
	client *Client, msgType protocol.MessageType, payload any) error {
    encoded_data, err := protocol.Encode(msgType, payload)
    if err != nil {
        server.logger.Error("Failed to encode message",
			"Message type", msgType,
			"To", client.id,
			"Err", err,
		)
		return err
	}

    err = client.conn.Write(ctx, websocket.MessageText, encoded_data)
	if err != nil {
		server.logger.Error("failed to send message",
			"Message type", msgType,
			"To", client.id,
			"Err", err,
		)
	}
	return err
}

func (server *Server) sendError(ctx context.Context, client *Client, code int, body any) {
    server.send(ctx, client, protocol.ErrorMessage, 
		protocol.ErrorPayload{
			Code: code,
			Body: body,
		})
}

func (server *Server) canRegister(ctx context.Context, client *Client, rawPayload []byte) bool {
    expected_msg := protocol.RegisterMessage

	msgType, data, err := protocol.DecodeEnvelope(rawPayload)
	if err != nil{
		server.logger.Error("Decoding envelope failed", "Err", err)
		return false
	}
	if msgType != expected_msg {
		server.logger.Error("Unexpected message type, denying connection",
			"Expected", expected_msg,
			"Got", msgType,
		)
		server.sendError(ctx, client, 1, "Expected register message type, connection denied")
		return false
	}
	decoded_payload, err := protocol.DecodePayload[protocol.RegisterPayload](data)
	if err != nil {
		server.logger.Error("Decoding register data failed",
			"Err", err,
		)
        server.sendError(ctx, client, 2, "Invalid register payload, connection denied")
		return false
	}
	client.name = decoded_payload.Name
	server.register(client)
	return true
}

func (server *Server) postRegisterProcess(ctx context.Context, client *Client) {
    server.broadcast(ctx, client, protocol.JoinMessage, protocol.JoinPayload{
        Id:   client.id,
        Name: client.name,
    })

    members := server.GetMembers()
    peers := make([]protocol.Member, 0, len(members)-1)
    for _, member := range members {
        if member.ID != client.id {
            peers = append(peers, member)
        }
    }
    server.send(ctx, client, protocol.SyncMessage, protocol.SyncPayload{
        Self:    client.id,
        Members: peers,
    })
}

func (server *Server) HandleClientLoop(
	ctx context.Context,
	conn *websocket.Conn,
	client *Client,
) {
	    for {
        _, raw, err := conn.Read(ctx)
        if err != nil {
			server.logger.Error("Error reading from connection",
				"Id", client.id,
				"Err", err,
			)
            break
        }

        msgType, data, err := protocol.DecodeEnvelope(raw)
        if err != nil {
            server.logger.Error("Failed to decode message envelope",
                "Id", client.id,
                "Err", err,
            )
            server.sendError(ctx, client, 1, err)
            continue
        }

        switch msgType {

        case protocol.ChatMessage:
            payload, err := protocol.DecodePayload[protocol.ChatPayload](data)
            if err != nil {
                server.logger.Error("Failed to decode chat message",
                    "Id", client.id,
                    "Err", err,
                )
                server.sendError(ctx, client, 2, "invalid msg payload")
                continue
            }
            var broadcast_payload protocol.BroadcastPayload
            broadcast_payload.From = client.id
            broadcast_payload.Body = payload.Body
            server.broadcast(ctx, client, protocol.BroadcastMessage, broadcast_payload)

        case protocol.DMsMessage:
            payload, err := protocol.DecodePayload[protocol.DMsPayload](data)
            if err != nil {
                server.logger.Error("Failed to decode dm message",
                    "Id", client.id,
                    "Err", err,
                )
                server.sendError(ctx, client, 2, "invalid dm payload")
                continue
            }
            server.mutex.RLock()
            target, ok := server.clients[payload.To]
            server.mutex.RUnlock()
            if !ok {
                server.logger.Error("DM target not found",
                    "Author Id", client.id,
                    "Target Id", payload.To,
                )
                server.sendError(ctx, client, 3, "user not found")
                continue
            }
            whisper_payload := protocol.WhisperPayload{
                From: client.id,
                Body: payload.Body,
            }
            server.send(ctx, target, protocol.WhisperMessage, whisper_payload)

        case protocol.PingMessage:
            server.send(ctx, client, protocol.PongMessage, protocol.PongPayload{})

        default:
            server.logger.Error("Received unauthorized message type",
                "Id", client.id,
                "Message type", msgType,
            )
            server.sendError(ctx, client, 4, "unauthorized message type")
        }
    }


}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Accept(w, r, nil)
    if err != nil {
        server.logger.Error("Failed to accept websocket connection", 
			"Err", err,
		)
        return
    }

    client := &Client{
        conn: conn,
        id:   uuid.New(),
    }

	server.logger.Info("New connection detected",
		"remote", r.RemoteAddr,	
	)	

	server.logger.Info("Awaiting register...")


    ctx := r.Context()
	_, rawPayload, err := conn.Read(ctx)
	if err != nil {
        server.logger.Error("Error reading from connection",
            "remote", r.RemoteAddr,
            "Err", err,
        )
		return
	}

    server.logger.Debug("Received register message, starting pre-registration checks...")
    if !server.canRegister(ctx, client, rawPayload) {
        return
    }

    server.logger.Info("Client registered successfully",
        "Id", client.id,
        "Name", client.name,
    )

    server.logger.Debug("Starting post-registration process...")
    server.postRegisterProcess(ctx, client)
    server.emit(Event{Type: ClientConnected, ID: client.id, Name: client.name, Time: time.Now()})

    
    server.HandleClientLoop(ctx, conn, client)
    
    server.logger.Info("Connection closed",
        "Id", client.id,
    )
    server.emit(Event{Type: ClientDisconnected, ID: client.id, Name: client.name, Time: time.Now()})
    server.unregister(client)
    server.broadcast(ctx,client, protocol.LeaveMessage, protocol.LeavePayload{
        From: client.id,
    })
    conn.Close(websocket.StatusNormalClosure, "")
}

func (server *Server) Start(addr string) error {
    return http.ListenAndServe(addr, server)
}