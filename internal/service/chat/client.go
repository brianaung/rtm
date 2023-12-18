package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/brianaung/rtm/view"
	"github.com/gorilla/websocket"
)

type msgData struct {
	Msg     string            `json:"msg"`
	Headers map[string]string `json:"HEADERS"`
}

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
	hub   *hub
	rid   string
	uid   string
	uname string
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	send chan *message
}

func newClient(hub *hub, rid string, uid string, uname string, conn *websocket.Conn) *client {
	return &client{
		hub:   hub,
		rid:   rid,
		uid:   uid,
		uname: uname,
		send:  make(chan *message),
		conn:  conn,
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readPump() {
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
		m = bytes.TrimSpace(bytes.Replace(m, newline, space, -1))
		c.hub.broadcast <- &message{data: m, rid: c.rid, uname: c.uname}
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

			data := &msgData{}
			json.Unmarshal(message.data, data)

			time := time.Now()
			formatted := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
				time.Year(), time.Month(), time.Day(),
				time.Hour(), time.Minute(), time.Second())

			view.MessageLog(view.MsgData{
				Rid:   c.rid,
				Uname: message.uname,
				Msg:   data.Msg,
				Time:  formatted,
				Mine:  c.uname == message.uname,
			}).Render(context.Background(), w)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				//w.Write(<-c.send)
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
