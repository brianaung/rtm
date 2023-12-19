package auth

import (
	"net/http"
	"os"

	"github.com/go-chi/jwtauth/v5"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	ja *jwtauth.JWTAuth
}

type UserContext struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

func Init() (a *Auth) {
	jwtAuth := jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)
	a = &Auth{ja: jwtAuth}
	return
}

func (a *Auth) GetJA() *jwtauth.JWTAuth {
	return a.ja
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
