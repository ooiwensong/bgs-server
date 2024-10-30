package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/ooiwensong/bgs_server/internal/db"
	"github.com/ooiwensong/bgs_server/internal/handlers"
)

func main() {
	db, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	// r.Mount("/auth", handlers.authRouter(db))
	// r.Mount("/api/sessions", handlers.SessionsRouter(db))
	r.Mount("/api/profiles", handlers.ProfilesRouter(db))
	// r.Mount("/api/library", handlers.LibrayRouter(db))
	// r.Mount("/admin", handler.AdminRouter(db))

	log.Fatal(http.ListenAndServe(":5001", r))
}
