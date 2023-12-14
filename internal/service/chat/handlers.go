package chat

import (
	"log"
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/ui"
	"github.com/go-chi/chi/v5"
)

func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value("user").(*auth.UserContext)
	w.WriteHeader(http.StatusFound)

	roomids := make([]string, 0)
	for k := range s.hubs {
		roomids = append(roomids, k)
	}
	pagedata := struct {
		User    *auth.UserContext
		Roomids []string
	}{
		User:    userData,
		Roomids: roomids,
	}

	ui.RenderPage(w, pagedata, "dashboard")
}

func (s *service) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	roomid := r.FormValue("roomid")
	if _, ok := s.hubs[roomid]; ok {
		w.Write([]byte("Room already exists"))
		return
	}

	h := newHub()
	s.hubs[roomid] = h
	go h.run()

	w.Header().Set("HX-Redirect", "/dashboard/room/"+roomid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	roomid := chi.URLParam(r, "roomid")
	if _, ok := s.hubs[roomid]; !ok {
		// todo: error handle
		return
	}
	ui.RenderPage(w, struct{ RoomId string }{RoomId: roomid}, "chatroom")
}

func (s *service) handleServeWs(w http.ResponseWriter, r *http.Request) {
	roomid := chi.URLParam(r, "roomid")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	_, ok := s.hubs[roomid]
	if !ok {
		return
	}

	client := &client{hub: s.hubs[roomid], conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

