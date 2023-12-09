package user

import (
	"fmt"
	"net/http"
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
	s.r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(s.jwtAuth))
		r.Use(jwtauth.Authenticator(s.jwtAuth))

		r.Get("/testing", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["username"])))
		})
		r.Get("/logout", s.handleLogout)
	})

	s.r.Group(func(r chi.Router) {
		r.Post("/signup", s.handleSignup)
		r.Post("/login", s.handleLogin)
	})

}
