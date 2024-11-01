package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type User struct {
	Email     *string `json:"email"`
	Username  *string `json:"username"`
	Avatar    *string `json:"avatar"`
	CreatedAt *string `json:"created_at"`
}

// if rows.Next() {
// 	user := User{}
// 	if err := rows.Scan(&user.uuid, &user.email, &user.hash, &user.username, &user.avatar, &user.role, &user.created_at); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("%+v", user.username)
// }

func ProfilesRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

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

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(u)
		if err != nil {
			log.Fatal(err)
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
		json.NewEncoder(w).Encode(struct {
			Status string
			Msg    string
		}{
			"ok",
			"user profile updated successfully",
		})
	}
}
