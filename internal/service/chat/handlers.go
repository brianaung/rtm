package chat

import (
	"log"
	"net/http"
)

func (s *service) handleServeWs(hub *hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// todo: add more client info? like the curr user info

		client := &client{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		// Allow collection of memory referenced by the caller by doing all work in
		// new goroutines.
		go client.writePump()
		go client.readPump()
	}
}

func (s *service) createRoom(w http.ResponseWriter, r *http.Request) {

}

func (s *service) joinRoom(w http.ResponseWriter, r *http.Request) {

}
