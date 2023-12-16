package chat

type hub struct {
	rooms      map[string]*room // room: map[string]*member
	broadcast  chan *message    // inbound messsages from the client
	register   chan *client     // register requests from the client
	unregister chan *client     // unregister requests from the client
	quit       chan bool
}

type room struct {
	rid   string
	rname string

	members map[string]*member
}

type member struct {
	uid     string
	clients map[*client]bool
}

type message struct {
	rid  string
	data []byte
}

func newHub() *hub {
	return &hub{
		rooms:      make(map[string]*room),
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
		// todo: another chan for removing user
		case c := <-h.register:
			// register client to the member in the hub
			// Note: make sure spaces are allocated before sending to this chan
			h.rooms[c.rid].members[c.uid].clients[c] = true
		case c := <-h.unregister:
			// remove client from clients map, and close the send channel
			if room, ok := h.rooms[c.rid]; ok {
				if _, ok := room.members[c.uid].clients[c]; ok {
					delete(h.rooms[c.rid].members[c.uid].clients, c)
					close(c.send)
				}
			}
		case m := <-h.broadcast:
			// broadcast messages to every client in the room
			members := h.rooms[m.rid].members
			for _, member := range members {
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
