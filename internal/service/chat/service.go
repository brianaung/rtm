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

		r.Get("/dashboard/create/{roomid}", func(w http.ResponseWriter, r *http.Request) {
			// todo: use form
			roomid := chi.URLParam(r, "roomid")
			s.hubs[roomid] = newHub()
			http.Redirect(w, r, "/dashboard/room/"+roomid, http.StatusSeeOther)
		})

		r.Get("/dashboard/room/{roomid}", func(w http.ResponseWriter, r *http.Request) {
			roomid := chi.URLParam(r, "roomid")
			if h, ok := s.hubs[roomid]; !ok {
				return
			} else {
				go h.run()
				ui.Render(w, struct{ RoomId string }{RoomId: roomid}, "chatroom")
			}
		})

		// ws connections
		r.Get("/ws/{roomid}", s.handleServeWs)
	})
}
