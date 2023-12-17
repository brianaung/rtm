package chat

import (
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/ui"
	"github.com/go-chi/chi/v5"
)

// roomData is used to pass room data into the html templates
type roomData struct {
	Rid   string
	Rname string
}

// handleDashboard serve the dashboard html with relevant information.
//
// Get the rooms that the current authorized user is a part of. The
// html is served using this information. It also includes forms for the
// user to create and join rooms.
func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rsData := make([]roomData, 0)
	for _, room := range s.hub.rooms {
		if _, ok := room.members[user.ID]; ok {
			rsData = append(rsData, roomData{Rid: room.rid, Rname: room.rname})
		}
	}
	data := struct {
		User *auth.UserContext
		Rs   []roomData
	}{User: user, Rs: rsData}
	w.WriteHeader(http.StatusOK)
	ui.RenderPage(w, data, "dashboard")
}

// handleCreateRoom creates a new room with the current user added.
//
// First, it checks whether the room already exists. Normally, this should always
// pass since a unique id will generate for every room creation. If it passes, a map to store
// the room information is allocated. Then a "member" object is created with
// no clients.
func (s *service) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rname := r.FormValue("rname")
	rid := rname + user.ID
	if _, ok := s.hub.rooms[rid]; ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room already exists"))
		return
	}
	// TODO: create uuid or sth similar
	s.hub.rooms[rid] = &room{rid: rid, rname: rname, members: make(map[string]*member)}
	s.hub.rooms[rid].members[user.ID] = &member{uid: user.ID, clients: make(map[*client]bool)}
	w.Header().Set("HX-Redirect", "/room/"+rid)
	w.WriteHeader(http.StatusOK)
}

// handleJoinRoom allows a user to gain access to a room.
//
// If the room exists and the user is not already in the room, a new "member"
// object is created with the current user information. No clients are created
// yet.
func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := r.FormValue("rid")
	room, ok := s.hub.rooms[rid]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	if _, ok := room.members[user.ID]; ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You are already in the room"))
		return
	}
	room.members[user.ID] = &member{uid: user.ID, clients: make(map[*client]bool)}
	w.Header().Set("HX-Redirect", "/room/"+rid)
	w.WriteHeader(http.StatusOK)
}

// handleGotoRoom will serve the html for the chatroom.
//
// Given that the user is authorized, it will render a html for the chatroom
// where a new websocket connection will begin, which is handled by `serveWs`.
func (s *service) handleGotoRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := chi.URLParam(r, "rid")
	room, ok := s.hub.rooms[rid]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	if _, ok := s.hub.rooms[rid].members[user.ID]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User does not have access to the room"))
		return
	}
	ui.RenderPage(w, roomData{Rid: room.rid, Rname: room.rname}, "chatroom")
}

// serveWs creates a websocket connection/client to use while in the chatroom.
//
// It upgrades the http connection to a websocket protocol. A new client is created
// with this connection which is registered to the hub. Then two goroutines starts
// for reading and writing messages.
func (s *service) serveWs(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := chi.URLParam(r, "rid")
	room, ok := s.hub.rooms[rid]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Something went wrong, room no longer exists"))
		return
	}
	member, ok := room.members[user.ID]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Something went wrong, you no longer have access to the room"))
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	c := newClient(s.hub, rid, user.ID, user.Username, conn)
	member.clients[c] = true
	s.hub.register <- c
	go c.writePump()
	go c.readPump()
}

// TODO: only allow admin to delete
func (s *service) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "rid")
	room, ok := s.hub.rooms[rid]
	if !ok || room.members == nil {
		return
	}
	for _, member := range room.members {
		for c := range member.clients {
			s.hub.unregister <- c
			c.conn.Close()
		}
	}
	delete(s.hub.rooms, rid)
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}
