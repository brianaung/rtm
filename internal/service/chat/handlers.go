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
// Get the rooms that the current authorized user is apart of. The
// html for dashboard is then served using this information.
func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rooms, err := getRoomsFromUser(r.Context(), s.db, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	roomsData := make([]view.RoomDisplayData, 0)
	for _, r := range rooms {
		roomsData = append(roomsData, view.RoomDisplayData{RoomID: r.ID, RoomName: r.Name})
	}
	w.WriteHeader(http.StatusOK)
	view.Dashboard(user, roomsData).Render(r.Context(), w)
}

// handleCreateRoom creates a new room with the current user added.
//
// It stores the room information and its member details in the database. Then
// to store the client connection informations, in-memory space is allocated.
func (s *service) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rname := r.FormValue("rname")
	rid := uuid.Must(uuid.NewV4())
	// to store client connections in-memory
	s.hub.rooms[rid] = make(map[*client]bool)
	// store room data and its members details in db
	if err := addRoom(r.Context(), s.db, &Room{ID: rid, Name: rname, CreatorID: user.ID}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if err := addUserToRoom(r.Context(), s.db, &RoomUser{RoomID: rid, UserID: user.ID}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("HX-Redirect", "/room/"+rid.String())
	w.WriteHeader(http.StatusOK)
}

// handleJoinRoom allows a user to gain access to a room.
//
// If the room exists and the user is not already in the room,
// the user is added to the room in the database, but the client
// connections are not yet created.
func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(r.FormValue("rid")))
	// check for room
	room, err := getRoomByID(r.Context(), s.db, rid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if room == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Room does not exists."))
		return
	}
	// check for user
	ok, err := isAMember(r.Context(), s.db, &RoomUser{RoomID: rid, UserID: user.ID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You are already in the room."))
		return
	}
	// finally, add user to room
	err = addUserToRoom(r.Context(), s.db, &RoomUser{RoomID: rid, UserID: user.ID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("HX-Redirect", "/room/"+rid.String())
	w.WriteHeader(http.StatusOK)
}

// handleGotoRoom will serve the html for the chatroom.
//
// Given that the user is authorized and have access to the room,
// it will render a html for the chatroom where a new websocket connection
// will begin, which is handled by `serveWs`.
func (s *service) handleGotoRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
	ok, err := isAMember(r.Context(), s.db, &RoomUser{RoomID: rid, UserID: user.ID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You do not have access to the room."))
		return
	}
	room, err := getRoomByID(r.Context(), s.db, rid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if room == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	// get message history
	msgData, err := getMessagesFromRoom(r.Context(), s.db, rid, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	view.Chatroom(user, view.RoomDisplayData{RoomID: room.ID, RoomName: room.Name}, msgData).Render(r.Context(), w)
}

// handleDeleteRoom allows user to delete the entire room.
//
// If the user have the permission to delete the room (i.e. is the creator of the room),
// then the in-memory client connections will be first cleaned. Then, related entries
// in the database will be removed.
func (s *service) handleDeleteRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
	// check for room
	room, err := getRoomByID(r.Context(), s.db, rid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if room == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	// check for permission
	if room.CreatorID != user.ID {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You do not have permission to delete this room."))
		return
	}
	// clean in-memory client connections
	if s.hub.rooms != nil && s.hub.rooms[rid] != nil {
		for c := range s.hub.rooms[rid] {
			s.hub.unregister <- c
			c.conn.Close()
		}
		delete(s.hub.rooms, rid)
	}
	if err := deleteRoom(r.Context(), s.db, rid); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// serveWs creates a websocket connection/client to use while in the chatroom.
//
// It upgrades the http connection to a websocket protocol. A new client is created
// with this connection which is registered to the hub. Then two goroutines starts
// for reading and writing messages.
func (s *service) serveWs(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*auth.UserContext)
	rid := uuid.Must(uuid.FromString(chi.URLParam(r, "rid")))
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
	go c.readPump(r, s.db)
}
