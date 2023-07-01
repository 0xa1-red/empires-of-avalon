package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/logging"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	intnats "github.com/0xa1-red/empires-of-avalon/transport/nats"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

var (
	wsupgrader = websocket.Upgrader{ // nolint:exhaustruct
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// Time allowed to write a message to the peer.
	// writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	// pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	// pingPeriod = (pongWait * 9) / 10
)

var hub = map[uuid.UUID]*Client{}

type Message struct {
	Timestamp time.Time
	Message   string
}

type Client struct {
	Connection  *Connection
	UserID      uuid.UUID
	InventoryID uuid.UUID
	Outgoing    chan Message
}

type Connection struct {
	*websocket.Conn
}

func (conn *Connection) handshake(msg []byte) (*Client, error) {
	handshake := make(map[string]string)
	if err := json.Unmarshal(msg, &handshake); err != nil {
		slog.Error("failed to unmarshal handshake message", err, "message", msg)
		return nil, fmt.Errorf("failed to unmarshal handshake message: %w", err)
	}

	id, err := uuid.Parse(handshake["UserID"])
	if err != nil {
		slog.Error("failed to parse user ID", err, "user_id", handshake["UserID"])
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		slog.Error("failed to set read deadline", err)
	}

	inventoryID := blueprints.GetInventoryID(id)

	return &Client{
		Connection:  conn,
		UserID:      id,
		InventoryID: inventoryID,
		Outgoing:    make(chan Message),
	}, nil
}

func (conn *Connection) writeMessage(msg interface{}) error {
	var err error

	retries := 3

	for {
		err = conn.WriteJSON(msg)
		if err != nil {
			retries--
		}

		if err == nil || retries == 0 {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

var configPath string

func main() {
	flag.StringVar(&configPath, "config-file", "", "path to config file")
	flag.Parse()

	config.Setup(configPath)

	if err := logging.Setup(); err != nil {
		os.Exit(1)
	}

	r := gin.Default()

	nc, err := intnats.GetConnection()
	if err != nil {
		slog.Error("failed to get NATS connection", err)
		os.Exit(1) // nolint
	}

	_, subErr := nc.Subscribe("status.*", func(m *nats.Msg) {
		userID := strings.TrimPrefix(m.Subject, "status.")
		id, err := uuid.Parse(userID)
		if err != nil {
			slog.Error("failed to parse user ID", err, "user_id", userID)
			return
		}

		slog.Debug("received message", "subject", m.Subject)

		i := protobuf.InventoryStatusUpdate{} // nolint:exhaustruct
		if err := proto.Unmarshal(m.Data, &i); err != nil {
			slog.Error("failed to unmarshal protobuf message", err)
			return
		}

		b := bytes.NewBuffer([]byte(""))
		encoder := json.NewEncoder(b)
		if err := encoder.Encode(&i); err != nil {
			slog.Error("failed to encode json message", err)
			return
		}

		if client, ok := hub[id]; ok {
			client.Outgoing <- Message{
				Timestamp: time.Now(),
				Message:   b.String(),
			}
		}
	})

	if subErr != nil {
		slog.Error("failed to subscribe to status subject", err)
		os.Exit(1)
	}

	defer nc.Close()

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})

	fmt.Println("http://localhost:3000")

	if err := r.Run("localhost:3000"); err != nil {
		slog.Error("http server returned error", err)
	}
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed to set websocket upgrade", err)
		return
	}

	conn := &Connection{
		Conn: wsConn,
	}

	if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		slog.Error("failed to set read deadline", err)
		conn.Close()

		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		slog.Error("failed to read message", err)
		conn.Close()

		return
	}

	client, err := conn.handshake(msg)
	if err != nil {
		slog.Error("failed to validate handshake", err)
		conn.Close()

		return
	}

	hub[client.InventoryID] = client
	slog.Info("client connected", "user_id", client.UserID, "inventory_id", client.InventoryID, "connection", conn.RemoteAddr())

	defer func() {
		delete(hub, client.UserID)
	}()

	go watchOutgoing(hub, client)

	for {
		t, _, err := conn.ReadMessage()
		if t != -1 || err == nil {
			continue
		}

		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			slog.Info("client disconnected", "user_id", client.UserID, "connection", conn.RemoteAddr())
			conn.Close()

			return
		}

		slog.Error("failed to read message", err, "user_id", client.UserID, "connection", conn.RemoteAddr())
		conn.Close()

		return
	}
}

func watchOutgoing(hub map[uuid.UUID]*Client, client *Client) {
	for msg := range hub[client.InventoryID].Outgoing {
		if err := client.Connection.writeMessage(msg); err != nil {
			slog.Error("failed to write message", err, "user_id", client.UserID, "connection", client.Connection.RemoteAddr())
		}
	}
}
