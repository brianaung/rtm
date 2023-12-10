package user

import (
	"net/http"

	"github.com/brianaung/rtm/ui"
)

func (s *service) handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	ui.Render(w, nil, "index")
}

func (s *service) handleGetLoginForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	ui.Render(w, nil, "loginForm")
}

func (s *service) handleGetSignupForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	ui.Render(w, nil, "signupForm")
}

func (s *service) handleSignup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// todo: more input validations
	if u, _ := getOneUser(r.Context(), s.db, username); u != nil {
		w.Write([]byte("username already exists"))
		return
	}

	hashedPassword, _ := hashAndSalt(password)
	u := &user{Username: username, Email: email, Password: hashedPassword}
	u, err := addUser(r.Context(), s.db, u)
	// res, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	claims := map[string]interface{}{"id": u.ID, "username": u.Username, "email": u.Email}
	setTokenCookie(w, s.jwtAuth, claims)

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

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

	claims := map[string]interface{}{"id": u.ID, "username": u.Username, "email": u.Email}
	setTokenCookie(w, s.jwtAuth, claims)

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "jwt", Value: "", MaxAge: -1, HttpOnly: true})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusFound)
}

func (s *service) handleDashboard(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value("user")
	w.WriteHeader(http.StatusFound)
	ui.Render(w, userData, "dashboard")
}
