package chat

type hub struct {
	clients    map[string]*client // registered clients
	broadcast  chan []byte        // inbound messsages from the client
	register   chan *client       // register requests from the client
	unregister chan *client       // unregister requests from the client
	close      chan bool
}

func newHub() *hub {
	return &hub{
		clients:    make(map[string]*client),
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		close:      make(chan bool),
	}
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			// register client to the hub
			h.clients[client.uid] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.uid]; ok {
				delete(h.clients, client.uid)
				close(client.send)
			}
		case message := <-h.broadcast:
			for uid, client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, uid)
					close(client.send)
				}
			}
		case flag := <-h.close:
			if flag {
				break
			}
		}
	}
}
