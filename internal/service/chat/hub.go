package chat

import (
	"github.com/gofrs/uuid/v5"
)

type hub struct {
	rooms      map[uuid.UUID]map[*client]bool
	broadcast  chan *message // inbound messsages from the client
	register   chan *client  // register requests from the client
	unregister chan *client  // unregister requests from the client
	quit       chan bool
}

type message struct {
	rid   uuid.UUID
	uname string
	data  []byte
}

func newHub() *hub {
	return &hub{
		rooms:      make(map[uuid.UUID]map[*client]bool),
		broadcast:  make(chan *message),
		register:   make(chan *client),
		unregister: make(chan *client),
		quit:       make(chan bool),
	}
}

func (h *hub) run() {
	for {
		select {
		// register, unregister chan is only for client/conn, not for removing entire user
		// TODO: another chan for removing user
		case c := <-h.register:
			// register client to the hub
			h.rooms[c.rid][c] = true
		case c := <-h.unregister:
			// remove client from the hub, and close its send channel
			if room, ok := h.rooms[c.rid]; ok {
				if _, ok := room[c]; ok {
					delete(h.rooms[c.rid], c)
					close(c.send)
				}
			}
		case m := <-h.broadcast:
			// broadcast messages to every client in the room
			for client := range h.rooms[m.rid] {
				select {
				case client.send <- m:
				default:
					close(client.send)
					delete(h.rooms[m.rid], client)
				}
			}
		case quit := <-h.quit:
			if quit {
				return
			}
		}
	}
}
