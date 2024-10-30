package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ProfilesRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	r.Post("/", getUserProfile(db))
	r.Patch("/", editUserProfile(db))

	return r
}

func getUserProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "get user profile")
	}
}

func editUserProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "edit user profile")
	}
}
