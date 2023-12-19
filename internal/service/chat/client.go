package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/brianaung/rtm/view"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// client is a middleman between the websocket connection and the hub.
type client struct {
	hub      *hub
	roomID   uuid.UUID
	userID   uuid.UUID
	username string
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	send chan *message
}

func newClient(hub *hub, rid uuid.UUID, uid uuid.UUID, uname string, conn *websocket.Conn) *client {
	return &client{
		hub:      hub,
		roomID:   rid,
		userID:   uid,
		username: uname,
		send:     make(chan *message),
		conn:     conn,
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readPump(r *http.Request, db *pgxpool.Pool) {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, m, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		mid := uuid.Must(uuid.NewV4())

		data := &struct {
			Msg     string            `json:"msg"`
			Headers map[string]string `json:"HEADERS"`
		}{}
		json.Unmarshal(m, data)
		err = addMessageEntry(context.Background(), db, &Message{ID: mid, Msg: data.Msg, Time: time.Now(), RoomID: c.roomID, UserID: c.userID})
		if err != nil {
			log.Printf("error: %v", err)
			break
		}

		m = bytes.TrimSpace(bytes.Replace(m, newline, space, -1))
		c.hub.broadcast <- &message{body: data.Msg, roomID: c.roomID, userID: c.userID, username: c.username, time: time.Now()}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			time := message.time
			formatted := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
				time.Year(), time.Month(), time.Day(),
				time.Hour(), time.Minute(), time.Second())
			view.MessageLog(view.MsgDisplayData{
				RoomID:   message.roomID,
				Username: message.username,
				Msg:      message.body,
				Time:     formatted,
				Mine:     c.userID == message.userID,
			}).Render(context.Background(), w)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				// w.Write(newline)
				// w.Write(<-c.send)
				// t.Execute(w, newline)
				// t.Execute(w, data.Msg)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
