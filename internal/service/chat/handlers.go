package chat

import (
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/ui"
	"github.com/go-chi/chi/v5"
)

func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	// append room if user is inside
	rids := make([]string, 0)
	for rid, room := range s.hub.rooms {
		if _, ok := room[user.ID]; ok {
			rids = append(rids, rid)
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
	if _, ok := s.hub.rooms[rid]; ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room already exists"))
		return
	}
	// allocate space and add creator to the room
	s.hub.rooms[rid] = make(map[string]*member)
	s.hub.rooms[rid][user.ID] = &member{uid: user.ID, clients: make(map[*client]bool)}
	// goto room
	w.Header().Set("HX-Redirect", "/room/"+rid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := r.FormValue("rid")
	// check if room exists
	if _, ok := s.hub.rooms[rid]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	// add user to room
	s.hub.rooms[rid][user.ID] = &member{uid: user.ID, clients: make(map[*client]bool)}
	// goto room
	w.Header().Set("HX-Redirect", "/room/"+rid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleGotoRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := chi.URLParam(r, "rid")
	// check if room exists
	if _, ok := s.hub.rooms[rid]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	// user is not in the room
	if _, ok := s.hub.rooms[rid][user.ID]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User does not have access to the room"))
		return
	}
	// otherwise go to chatroom (which will establish a ws connection)
	ui.RenderPage(w, struct{ RoomId string }{RoomId: rid}, "chatroom")
}

// ws connection
func (s *service) serveWs(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := chi.URLParam(r, "rid")
	// check if room exists
	room, ok := s.hub.rooms[rid]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}
	// check if member is in room
	member, ok := room[user.ID]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User does not have access to the room"))
		return
	}
	// create a new connection for this session
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	c := newClient(s.hub, rid, user.ID, conn)
	member.clients[c] = true
	s.hub.register <- c
	go c.writePump()
	go c.readPump()
}

// todo: everyone in the room can delete rooms right now, which is bad
func (s *service) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "rid")
	room, ok := s.hub.rooms[rid]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}
	for uid, member := range room {
		for c := range member.clients {
			c.hub.unregister <- c
			c.conn.Close()
		}
		delete(room, uid)
	}
	delete(s.hub.rooms, rid)

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}
