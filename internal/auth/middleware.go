package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

/* This is a modification of go-chi/jwtauth Authenticator middleware to handle redirection upon successful authentication.
 * See here: https://github.com/go-chi/jwtauth/blob/master/jwtauth.go#L171
 */
func (a *Auth) Authenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			// validate jwt token
			token, claims, err := jwtauth.FromContext(r.Context())
			if err != nil || token == nil || jwt.Validate(token, a.ja.ValidateOptions()...) != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// set context with logged in user data so other handlers have access to it
			res := UserContext{
				ID:       uuid.Must(uuid.FromString(claims["id"].(string))),
				Username: claims["username"].(string),
				Email:    claims["email"].(string),
			}
			ctx := context.WithValue(r.Context(), "user", &res)

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}
