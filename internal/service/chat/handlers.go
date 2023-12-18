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
	rids, err := getRidsForUser(r.Context(), s.db, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	for _, rid := range rids {
		room, _ := getRoomByID(r.Context(), s.db, rid)
		rsData = append(rsData, view.RoomData{Rid: room.ID, Rname: room.Name})
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
	s.hub.rooms[rid] = make(map[*client]bool)
	err := addRoom(r.Context(), s.db, &Room{ID: rid, Name: rname, CreatorID: user.ID})
	err = addMembership(r.Context(), s.db, &RoomUser{Rid: rid, Uid: user.ID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	// todo: create membership table? (i.e. room and user/members junction table)
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
	if _, err := getRoomByID(r.Context(), s.db, rid); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	if ok, _ := isAMember(r.Context(), s.db, &RoomUser{Rid: rid, Uid: user.ID}); ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You are already in the room"))
		return
	}
	err := addMembership(r.Context(), s.db, &RoomUser{Rid: rid, Uid: user.ID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
	}
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
	if ok, err := isAMember(r.Context(), s.db, &RoomUser{Rid: rid, Uid: user.ID}); err != nil || !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists or user have no access to it"))
		return
	}
	room, err := getRoomByID(r.Context(), s.db, rid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists"))
		return
	}
	view.Chatroom(user, view.RoomData{Rid: room.ID, Rname: room.Name}).Render(r.Context(), w)
}

// serveWs creates a websocket connection/client to use while in the chatroom.
//
// It upgrades the http connection to a websocket protocol. A new client is created
// with this connection which is registered to the hub. Then two goroutines starts
// for reading and writing messages.
func (s *service) serveWs(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
	//	room, ok := s.hub.rooms[rid]
	//	if !ok {
	//		w.WriteHeader(http.StatusBadRequest)
	//		w.Write([]byte("Something went wrong, room no longer exists"))
	//		return
	//	}
	//	member, ok := room.members[user.ID]
	//	if !ok {
	//		w.WriteHeader(http.StatusBadRequest)
	//		w.Write([]byte("Something went wrong, you no longer have access to the room"))
	//		return
	//	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	// client conn data is in-memory only so this can get erased when the server restarts
	if s.hub.rooms == nil {
		s.hub.rooms = make(map[uuid.UUID]map[*client]bool)
		s.hub.rooms[rid] = make(map[*client]bool)
	} else if s.hub.rooms[rid] == nil {
		s.hub.rooms[rid] = make(map[*client]bool)
	}
	c := newClient(s.hub, rid, user.ID, user.Username, conn)
	s.hub.register <- c
	go c.writePump()
	go c.readPump()
}

/*
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
*/
