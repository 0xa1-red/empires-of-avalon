package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xa1-red/empires-of-avalon/common"
	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/logging"
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
	wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var hub = map[uuid.UUID]*Client{}

type Message struct {
	Timestamp time.Time
	Message   string
}

type Client struct {
	Connection  *websocket.Conn
	UserID      uuid.UUID
	InventoryID uuid.UUID
	Outgoing    chan Message
}

var configPath string

func main() {
	flag.StringVar(&configPath, "config-file", "", "path to config file")
	flag.Parse()

	config.Setup(configPath)
	logging.Setup()
	r := gin.Default()

	nc := intnats.GetConnection()
	nc.Subscribe("status.*", func(m *nats.Msg) {
		userID := strings.TrimPrefix(m.Subject, "status.")
		id, err := uuid.Parse(userID)
		if err != nil {
			slog.Error("failed to parse user ID", err, "user_id", userID)
			return
		}

		slog.Debug("received message", "subject", m.Subject)

		i := protobuf.InventoryStatusUpdate{}
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
	defer nc.Close()

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})

	fmt.Println("http://localhost:3000")
	r.Run("localhost:3000")
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed to set websocket upgrade", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	handshake := make(map[string]string)
	if err := json.Unmarshal(msg, &handshake); err != nil {
		slog.Error("failed to unmarshal handshake message", err, "message", msg)
		conn.Close()
	}

	id, err := uuid.Parse(handshake["UserID"])
	if err != nil {
		slog.Error("failed to parse user ID", err, "user_id", handshake["UserID"])
		conn.Close()
	}

	conn.SetReadDeadline(time.Time{})

	inventoryID := common.GetInventoryID(id)

	hub[inventoryID] = &Client{
		Connection:  conn,
		UserID:      id,
		InventoryID: inventoryID,
		Outgoing:    make(chan Message),
	}
	slog.Info("client connected", "user_id", id, "inventory_id", inventoryID, "connection", conn.RemoteAddr())
	defer func() {
		delete(hub, id)
	}()

	go func() {
		for msg := range hub[inventoryID].Outgoing {
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
				slog.Error("failed to write message", err, "user_id", id, "connection", conn.RemoteAddr())
				conn.Close()
				continue
			}
		}
	}()

	for {
		t, _, err := conn.ReadMessage()
		if t != -1 {
			continue
		}
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				slog.Info("client disconnected", "user_id", id, "connection", conn.RemoteAddr())
				conn.Close()
				return
			}
			slog.Error("failed to read message", err, "user_id", id, "connection", conn.RemoteAddr())
			conn.Close()
			return
		}
	}
}
