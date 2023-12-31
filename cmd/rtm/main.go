package main

import (
	"log"
	"net/http"

	"github.com/brianaung/rtm/internal/auth"
	"github.com/brianaung/rtm/internal/db"
	"github.com/brianaung/rtm/internal/service/chat"
	"github.com/brianaung/rtm/internal/service/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading env")
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	fs := http.FileServer(http.Dir("dist"))
	r.Handle("/dist/*", http.StripPrefix("/dist/", fs))

	// setup db
	dbpool, err := db.Init()
	if err != nil {
		log.Fatal("Error initialising db")
	}
	defer dbpool.Close()

	// setup auth
	userauth := auth.Init()

	// inject dependencies to services
	userService := user.NewService(r, dbpool.Get(), userauth)
	chatService := chat.NewService(r, dbpool.Get(), userauth)

	// start services
	userService.Routes()
	chatService.Routes()

	log.Fatal(http.ListenAndServe(":3000", r))
}
