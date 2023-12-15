package chat

import (
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/ui"
	"github.com/go-chi/chi/v5"
)

func (s *service) serveWs(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "rid")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	// does hub exists
	_, ok := s.hubs[rid]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	// if so, register client
	client := newClient(s.hubs[rid].h, conn)
	client.hub.register <- client
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	// append room if user is inside
	rids := make([]string, 0)
	for k := range s.hubs {
		if _, ok := s.hubs[k].uids[user.ID]; ok {
			rids = append(rids, k)
		}
	}
	// data for html
	data := struct {
		User *auth.UserContext
		Rids []string
	}{
		User: user,
		Rids: rids,
	}
	w.WriteHeader(http.StatusOK)
	ui.RenderPage(w, data, "dashboard")
}

func (s *service) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := r.FormValue("rid")
	// room name should be unique (todo: should i use uuid?)
	if _, ok := s.hubs[rid]; ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room already exists"))
		return
	}
	// setup the hub (room)
	h := newHub()
	s.hubs[rid] = &hubdata{h: h, uids: map[string]bool{user.ID: true}}
	go h.run()
	// goto room by setting htmx redirect header
	w.Header().Set("HX-Redirect", "/room/"+rid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := r.FormValue("rid")
	h, ok := s.hubs[rid]
	// check if room actually exists
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	// add user to uid set
	h.uids[user.ID] = true
	// goto room
	w.Header().Set("HX-Redirect", "/room/"+rid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleGotoRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := chi.URLParam(r, "rid")
	// todo: s.hubs[rid] might be nil
	if _, ok := s.hubs[rid].uids[user.ID]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User does not have access to the room"))
		return
	}
	// otherwise go to chatroom
	ui.RenderPage(w, struct{ RoomId string }{RoomId: rid}, "chatroom")
}

// todo: everyone in the room can delete rooms right now, which is bad
func (s *service) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "rid")
	cs := s.hubs[rid].h.clients

	for c := range cs {
		c.hub.unregister <- c
		c.conn.Close()
	}

    // todo: i should not need this i only i had not created extra uid set =))
	delete(s.hubs, rid)

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}
