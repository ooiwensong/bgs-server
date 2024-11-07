package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ooiwensong/bgs_server/internal/middlewares"
)

type User struct {
	Email     *string `json:"email"`
	Username  *string `json:"username"`
	Avatar    *string `json:"avatar"`
	CreatedAt *string `json:"created_at"`
}

func ProfilesRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Auth)

	r.Post("/", getUserProfile(db))
	r.Patch("/", editUserProfile(db))

	return r
}

func getUserProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
			SELECT email, username, avatar, created_at
			FROM users
			WHERE uuid=$1
			`

		row := db.QueryRow(q, body["userId"])

		var (
			email     string
			username  sql.NullString
			avatar    string
			createdAt string
		)
		err = row.Scan(&email, &username, &avatar, &createdAt)
		if err != nil {
			http.Error(w, "error retreiving user profile", http.StatusBadRequest)
			log.Fatal(err)
		}

		u := &User{}
		u.Email = &email
		u.Avatar = &avatar
		u.CreatedAt = &createdAt
		if username.Valid {
			u.Username = &username.String
		} else {
			u.Username = nil
		}

		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(u)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	}
}

func editUserProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userId := r.Context().Value("decode").(*middlewares.Claims).UserId
		if userId != body["userId"] {
			http.Error(w, "not authorised", http.StatusBadRequest)
			return
		}

		// Maybe abstract into a validator middleware
		if body["attribute"] == "created_at" {
			http.Error(w, "error editing user profile", http.StatusBadRequest)
			return
		}

		q := fmt.Sprintf(
			`
			UPDATE users
			SET %s=$1
			WHERE uuid=$2
			`,
			body["attribute"],
		)

		_, err = db.Exec(q, body["value"], body["userId"])
		if err != nil {
			http.Error(w, "error editing user profile", http.StatusBadRequest)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(struct {
			Status string `json:"struct"`
			Msg    string `json:"msg"`
		}{
			"ok",
			"user profile updated successfully",
		})
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	}
}
