package chat

import (
	"github.com/brianaung/rtm/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	r        *chi.Mux
	db       *pgxpool.Pool
	userauth *auth.Auth
    hub *hub
}

func NewService(r *chi.Mux, db *pgxpool.Pool, userauth *auth.Auth) (s *service) {
    // Note: it can scale by injecting multiple hubs (each handling a set of chat rooms)
	h := newHub()
    go h.run()
    s = &service{r: r, db: db, userauth: userauth, hub: h}
	return
}

func (s *service) Routes() {
	// protected
	s.r.Group(func(r chi.Router) {
		// middlewares
		r.Use(jwtauth.Verifier(s.userauth.GetJA()))
		r.Use(s.userauth.Authenticator())

		r.Get("/dashboard", s.handleDashboard)
		r.Post("/create", s.handleCreateRoom)
		//r.Post("/join", s.handleJoinRoom)
		//r.Get("/room/{rid}", s.handleGotoRoom)
		//r.Get("/delete/{rid}", s.handleDeleteRoom)

		// todo: unregister route?

		// serve ws connection
		//r.Get("/ws/chat/{rid}", s.serveWs)
	})
}
