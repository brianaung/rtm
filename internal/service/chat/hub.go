package chat

type hub struct {
	rooms      map[string]room // rooms[rid][uid] = *client
	broadcast  chan *message   // inbound messsages from the client
	register   chan *client    // register requests from the client
	unregister chan *client    // unregister requests from the client
	quit       chan bool
}

type room map[string]*member

type member struct {
	uid     string
	clients map[*client]bool
}

type message struct {
	uid  string
	rid  string
	data []byte
}

func newHub() *hub {
	return &hub{
		rooms:      make(map[string]room),
		broadcast:  make(chan *message),
		register:   make(chan *client),
		unregister: make(chan *client),
		quit:       make(chan bool),
	}
}

func (h *hub) run() {
	for {
		select {
		// register, unregister route should only be for client/conn, not for removing user
		// todo: another route for removing user as well(i.e. remove all conn for that user)?
		case c := <-h.register:
			// register client to the hub
			// Note: pls make sure spaces are allocated before sending to this chan (no nil dereference)
			h.rooms[c.rid][c.uid].clients[c] = true
		case c := <-h.unregister:
            // remove client from map, and close the send channel
			if _, ok := h.rooms[c.rid][c.uid].clients[c]; ok {
				delete(h.rooms[c.rid][c.uid].clients, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			room := h.rooms[m.rid]
			for _, member := range room {
				for c := range member.clients {
					select {
					case c.send <- m.data:
					default:
						delete(member.clients, c)
						close(c.send)
					}
				}
			}
		case quit := <-h.quit:
			if quit {
				return
			}
		}
	}
}
