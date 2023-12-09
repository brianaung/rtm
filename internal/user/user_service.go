package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	router *chi.Mux
	db     *pgxpool.Pool
}

func NewService(r *chi.Mux, db *pgxpool.Pool) (s *service) {
	s = &service{router: r, db: db}
	return
}

func (s *service) Routes() {
	s.router.Get("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from User service!"))
	})
	s.router.Post("/signup", myMiddleware(s.handleSignup))
	// s.r.Post("/signup", myMiddleware(s.handleSignup))
	// s.r.Post("/signup", myMiddleware(s.handleSignup))
	// s.r.Post("/signup", myMiddleware(s.handleSignup))
}
