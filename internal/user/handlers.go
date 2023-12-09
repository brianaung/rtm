package user

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type JwtClaim struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func (s *service) handleSignup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// input validations
	if u, _ := getOneUser(r.Context(), s.db, username); u != nil {
		w.Write([]byte("username already exists"))
		return
	}

	hashedPassword, _ := hashAndSalt(password)
	u := &user{Username: username, Email: email, Password: hashedPassword}
	u, err := addUser(r.Context(), s.db, u)
	res, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(res)
}

func (s *service) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// authenticate user
	u, err := getOneUser(r.Context(), s.db, username)
	if err != nil {
		w.Write([]byte("user not found"))
		return
	}
	err = checkPassword(u.Password, password)
	if err != nil {
		w.Write([]byte("wrong password"))
		return
	}

	_, tokenString, _ := s.jwtAuth.Encode(map[string]interface{}{"username": u.Username})
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		HttpOnly: true, // Helps to mitigate XSS attacks
	}
	http.SetCookie(w, &cookie)

	// response user data
	res, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(res)
}

func (s *service) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "jwt", Value: "", MaxAge: -1, HttpOnly: true})
}
