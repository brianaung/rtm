package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	ja *jwtauth.JWTAuth
}

type userContext struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func Init() (a *Auth) {
	jwtAuth := jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)
	a = &Auth{ja: jwtAuth}
	return
}

func (a *Auth) GetJA() *jwtauth.JWTAuth {
	return a.ja
}

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
			res := userContext{
				ID:       claims["id"].(string),
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

// helpers
func (a *Auth) HashAndSalt(password string) (string, error) {
	// GenerateFromPassword salt the password for us aside from hashing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashedPassword), err
}

func (a *Auth) CheckPassword(hashed string, p string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(p))
}

func (a *Auth) SetTokenCookie(w http.ResponseWriter, claims map[string]interface{}) {
	_, tokenString, _ := a.ja.Encode(claims)
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		HttpOnly: true, // Helps to mitigate XSS attacks
	}
	http.SetCookie(w, &cookie)
}
