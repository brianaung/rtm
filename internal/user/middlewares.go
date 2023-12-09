package user

import (
	"context"
	"net/http"
	"time"
)

func myMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
		defer cancel()

		next(w, r.WithContext(ctx))
	}
}
