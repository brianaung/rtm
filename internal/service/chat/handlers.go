package chat

import (
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/view"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

// handleDashboard serve the dashboard html with relevant information.
//
// Get the rooms that the current authorized user is a part of. The
// html is served using this information. It also includes forms for the
// user to create and join rooms.
func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rsData := make([]view.RoomData, 0)
	for _, room := range s.hub.rooms {
		if _, ok := room.members[user.ID]; ok {
			rsData = append(rsData, view.RoomData{Rid: room.rid, Rname: room.rname})
		}
	}
	w.WriteHeader(http.StatusOK)
	view.Dashboard(user, rsData).Render(r.Context(), w)
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
	rid := uuid.Must(uuid.NewV4())
	s.hub.rooms[rid] = &room{rid: rid, rname: rname, members: make(map[uuid.UUID]*member)}
	s.hub.rooms[rid].members[user.ID] = &member{uid: user.ID, clients: make(map[*client]bool)}
	//addRoom(r.Context(), s.db, &Room{ID: rid, Name: rname, CreatorID: user.ID})
	w.Header().Set("HX-Redirect", "/room/"+rid.String())
	w.WriteHeader(http.StatusOK)
}

// handleJoinRoom allows a user to gain access to a room.
//
// If the room exists and the user is not already in the room, a new "member"
// object is created with the current user information. No clients are created
// yet.
func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(r.FormValue("rid")))
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
	w.Header().Set("HX-Redirect", "/room/"+rid.String())
	w.WriteHeader(http.StatusOK)
}

// handleGotoRoom will serve the html for the chatroom.
//
// Given that the user is authorized, it will render a html for the chatroom
// where a new websocket connection will begin, which is handled by `serveWs`.
func (s *service) handleGotoRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
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
	view.Chatroom(user, view.RoomData{Rid: room.rid, Rname: room.rname}).Render(r.Context(), w)
}

// serveWs creates a websocket connection/client to use while in the chatroom.
//
// It upgrades the http connection to a websocket protocol. A new client is created
// with this connection which is registered to the hub. Then two goroutines starts
// for reading and writing messages.
func (s *service) serveWs(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
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
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
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
