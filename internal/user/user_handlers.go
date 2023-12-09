package user

import "net/http"

func (s *service) handleFoo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from user service!"))
}
