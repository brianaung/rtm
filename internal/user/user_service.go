package user

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	r       *chi.Mux
	db      *pgxpool.Pool
	jwtAuth *jwtauth.JWTAuth
}

func NewService(r *chi.Mux, db *pgxpool.Pool) (s *service) {
	jwtAuth := jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)
	s = &service{r: r, db: db, jwtAuth: jwtAuth}
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
		r.Get("/logout", s.handleLogout)
	})

	// protected
	s.r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(s.jwtAuth))
		r.Use(Authenticator(s.jwtAuth))

		r.Get("/dashboard", s.handleDashboard)
	})
}
