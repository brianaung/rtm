package user

import (
	"encoding/json"
	"net/http"
)

func (s *service) handleSignup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// input validation

	hashedPassword, _ := hashAndSalt(password)
	u := &user{Username: username, Email: email, Password: hashedPassword}
	u, err := s.addUser(r.Context(), u)
	res, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(res)
}
