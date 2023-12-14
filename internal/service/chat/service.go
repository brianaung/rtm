package chat

import (
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/ui"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	r        *chi.Mux
	db       *pgxpool.Pool
	userauth *auth.Auth
	hubs     map[string]*hub
}

func NewService(r *chi.Mux, db *pgxpool.Pool, userauth *auth.Auth) (s *service) {
	s = &service{r: r, db: db, userauth: userauth, hubs: make(map[string]*hub)}
	return
}

func (s *service) Routes() {
	// protected
	s.r.Group(func(r chi.Router) {
		// middlewares
		r.Use(jwtauth.Verifier(s.userauth.GetJA()))
		r.Use(s.userauth.Authenticator())

		// todo: another auth middleware: user needs to be in the room?

        // todo: route for serving create room form page

		r.Post("/dashboard/create", func(w http.ResponseWriter, r *http.Request) {
			roomName := r.FormValue("roomName")
            if _, ok := s.hubs[roomName]; ok {
                w.Write([]byte("Room already exists"))
                return
            }

			h := newHub()
			s.hubs[roomName] = h
			go h.run()
			http.Redirect(w, r, "/dashboard/room/"+roomName, http.StatusSeeOther)
		})

		r.Get("/dashboard/room/{roomid}", func(w http.ResponseWriter, r *http.Request) {
			roomid := chi.URLParam(r, "roomid")
			if _, ok := s.hubs[roomid]; !ok {
				// todo: error handle
				return
			}
			ui.RenderPage(w, struct{ RoomId string }{RoomId: roomid}, "chatroom")
		})

        // todo: unregister route?

		// serve ws connection
		r.Get("/ws/chat/{roomid}", s.handleServeWs)
	})
}
