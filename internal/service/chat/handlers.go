package chat

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *service) handleServeWs(w http.ResponseWriter, r *http.Request) {
	roomid := chi.URLParam(r, "id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// todo: add more client info? like the curr user info

	if _, ok := s.hubs[roomid]; !ok {
		return
	}

	client := &client{hub: s.hubs[roomid], conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
