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
	hubs     map[string]*hubdata // { "hubName" : { hub, uids set }
}

type hubdata struct {
	h    *hub
	uids map[string]bool // uids set
}

func NewService(r *chi.Mux, db *pgxpool.Pool, userauth *auth.Auth) (s *service) {
	s = &service{r: r, db: db, userauth: userauth, hubs: make(map[string]*hubdata)}
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
		r.Post("/join", s.handleJoinRoom)
		r.Get("/room/{roomid}", s.handleGotoRoom)

		// todo: unregister route?

		// serve ws connection
		r.Get("/ws/chat/{roomid}", s.serveWs)
	})
}
