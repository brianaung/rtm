package chat

type hub struct {
	// clients    map[string]*client // registered clients
	rooms      map[string]map[string]*client // rooms[rid][uid] = *client
	broadcast  chan *message                 // inbound messsages from the client
	register   chan *sub                     // register requests from the client
	unregister chan *sub                     // unregister requests from the client
	quit       chan bool
}

type sub struct {
	rid    string
	uid    string
	client *client
}

type message struct {
	rid  string
	data []byte
}

func newHub() *hub {
	return &hub{
		rooms:      make(map[string]map[string]*client),
		broadcast:  make(chan *message),
		register:   make(chan *sub),
		unregister: make(chan *sub),
		quit:       make(chan bool),
	}
}

func (h *hub) run() {
	for {
		select {
		case sub := <-h.register:
			// register client to the hub
			h.rooms[sub.rid][sub.uid] = sub.client
		case sub := <-h.unregister:
			if _, ok := h.rooms[sub.rid][sub.uid]; ok {
				delete(h.rooms[sub.rid], sub.uid)
				close(sub.client.send)
			}
		case m := <-h.broadcast:
			room := h.rooms[m.rid]
			for uid, client := range room {
				select {
				case client.send <- m.data:
				default:
					delete(room, uid)
					close(client.send)
				}
			}
		case quit := <-h.quit:
			if quit {
				return
			}
		}
	}
}
