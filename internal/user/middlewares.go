package user

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func Authenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil {
				// redirect to home
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			if token == nil || jwt.Validate(token, ja.ValidateOptions()...) != nil {
				// redirect to home
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
