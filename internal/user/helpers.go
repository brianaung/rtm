package user

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(password string) (string, error) {
	// GenerateFromPassword salt the password for us aside from hashing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashedPassword), err
}

func checkPassword(hashed string, p string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(p))
}

func setTokenCookie(w http.ResponseWriter, ja *jwtauth.JWTAuth, claims map[string]interface{}) {
	_, tokenString, _ := ja.Encode(claims)
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		HttpOnly: true, // Helps to mitigate XSS attacks
	}
	http.SetCookie(w, &cookie)
}
