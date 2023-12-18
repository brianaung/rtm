package user

import (
	"net/http"

	"github.com/brianaung/rtm/view"
)

func (s *service) handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	view.Landing().Render(r.Context(), w)
}

func (s *service) handleGetLoginForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	//ui.RenderComponent(w, nil, "login-form")
	view.LoginForm().Render(r.Context(), w)
}

func (s *service) handleGetSignupForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	//ui.RenderComponent(w, nil, "signup-form")
	view.SignupForm().Render(r.Context(), w)
}

/* ================================================ */
/* Deals with user auth, jwt token creations and managing token cookies */
func (s *service) handleSignup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// todo: more input validations
	if u, _ := getOneUser(r.Context(), s.db, username); u != nil {
		w.Write([]byte("username already exists"))
		return
	}

	hashedPassword, _ := s.userauth.HashAndSalt(password)
	u := &User{Username: username, Email: email, Password: hashedPassword}
	u, err := addUser(r.Context(), s.db, u)
	// res, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	claims := map[string]interface{}{"id": u.ID, "username": u.Username, "email": u.Email}
	s.userauth.SetTokenCookie(w, claims)

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
	err = s.userauth.CheckPassword(u.Password, password)
	if err != nil {
		w.Write([]byte("wrong password"))
		return
	}

	claims := map[string]interface{}{"id": u.ID, "username": u.Username, "email": u.Email}
	s.userauth.SetTokenCookie(w, claims)

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (s *service) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "jwt", Value: "", MaxAge: -1, HttpOnly: true})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusFound)
}

/* ================================================ */
