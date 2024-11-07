package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/ooiwensong/bgs_server/internal/db"
	"github.com/ooiwensong/bgs_server/internal/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	db, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Mount("/auth", handlers.AuthRouter(db))
	r.Mount("/api/sessions", handlers.SessionsRouter(db))
	r.Mount("/api/profiles", handlers.ProfilesRouter(db))
	r.Mount("/api/library", handlers.LibrayRouter(db))
	r.Mount("/admin", handlers.AdminRouter(db))

	log.Fatal(http.ListenAndServe(":5001", r))
}
