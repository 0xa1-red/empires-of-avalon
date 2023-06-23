package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/logging"
	intnats "github.com/0xa1-red/empires-of-avalon/transport/nats"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"golang.org/x/exp/slog"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var hub = map[uuid.UUID]*Client{}

type Message struct {
	Timestamp time.Time
	Message   string
}

type Client struct {
	Connection *websocket.Conn
	UserID     uuid.UUID
	Outgoing   chan Message
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

		if client, ok := hub[id]; ok {
			client.Outgoing <- Message{
				Timestamp: time.Now(),
				Message:   string(m.Data),
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
	auth := r.Header.Get("Authorization")
	if auth == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(auth)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	hub[id] = &Client{
		Connection: conn,
		UserID:     id,
		Outgoing:   make(chan Message),
	}
	slog.Info("client connected", "user_id", id, "connection", conn.RemoteAddr())
	defer func() {
		delete(hub, id)
	}()

	go func() {
		for msg := range hub[id].Outgoing {
			if err := conn.WriteJSON(msg); err != nil {
				slog.Error("failed to write message", err, "user_id", id, "connection", conn.RemoteAddr())
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
				slog.Info("client connected", "user_id", id, "connection", conn.RemoteAddr())
				conn.Close()
				return
			}
			slog.Error("failed to read message", err, "user_id", id, "connection", conn.RemoteAddr())
			conn.Close()
			return
		}
	}
}
