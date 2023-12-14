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
        if contains(s.hubs[k].uids, userData.ID) {
		    roomids = append(roomids, k)
        }
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
	userData := r.Context().Value("user").(*auth.UserContext)

	roomid := r.FormValue("roomid")
	if _, ok := s.hubs[roomid]; ok {
		http.Error(w, "Room already exists", http.StatusInternalServerError)
		return
	}

	h := newHub()
	s.hubs[roomid] = &hubdata{h: h, uids: []string{userData.ID}}
	go h.run()

	w.Header().Set("HX-Redirect", "/dashboard/room/"+roomid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value("user").(*auth.UserContext)
	roomid := r.FormValue("roomid")
	hs, ok := s.hubs[roomid]
	if !ok {
		http.Error(w, "Room does not exists", http.StatusInternalServerError)
		return
	}

    // append if not already in the room
    if !contains(hs.uids, userData.ID) {
	    hs.uids = append(hs.uids, userData.ID)
    }

	w.Header().Set("HX-Redirect", "/dashboard/room/"+roomid)
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleGotoRoom(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value("user").(*auth.UserContext)
	roomid := chi.URLParam(r, "roomid")
	if hinfo, ok := s.hubs[roomid]; !ok || !contains(hinfo.uids, userData.ID) {
		http.Error(w, "User does not have access to room", http.StatusInternalServerError)
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

	client := &client{hub: s.hubs[roomid].h, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// helpers
func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
