package user

import (
	"github.com/brianaung/rtm/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-chi/jwtauth/v5"
)

type service struct {
	r  *chi.Mux
	db *pgxpool.Pool
    userauth *auth.Auth 
}

func NewService(r *chi.Mux, db *pgxpool.Pool, userauth *auth.Auth) (s *service) {
	s = &service{r: r, db: db, userauth: userauth}
	return
}

func (s *service) Routes() {
	// public
	s.r.Group(func(r chi.Router) {
		r.Get("/", s.handleHome)
		r.Post("/signup", s.handleSignup)
		r.Get("/signup-form", s.handleGetSignupForm)
		r.Post("/login", s.handleLogin)
		r.Get("/login-form", s.handleGetLoginForm)
	})

	// protected
	s.r.Group(func(r chi.Router) {
		// middlewares
		r.Use(jwtauth.Verifier(s.userauth.GetJA()))
		r.Use(s.userauth.Authenticator())

		r.Get("/dashboard", s.handleDashboard)
		r.Get("/logout", s.handleLogout)
	})
}
