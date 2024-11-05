package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func AdminRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Get("/users", getAllUsers(db))
	r.Patch("/users", updateUserRole(db))
	r.Delete("/users", deleteUser(db))

	return r
}

func getAllUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := `
		SELECT *
		FROM users
		ORDER BY created_at
		`
		rows, err := db.Query(q)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		type Row struct {
			Uuid      string  `json:"uuid"`
			Email     string  `json:"email"`
			Hash      string  `json:"hash"`
			Username  *string `json:"username"`
			Avatar    string  `json:"avatar"`
			Role      string  `json:"role"`
			CreatedAt string  `json:"created_at"`
		}
		users := []Row{}
		for rows.Next() {
			var r Row
			err = rows.Scan(&r.Uuid, &r.Email, &r.Hash, &r.Username, &r.Avatar, &r.Role, &r.CreatedAt)
			if err != nil {
				http.Error(w, "error retrieving users", http.StatusInternalServerError)
			}
			users = append(users, r)
		}
		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(users)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	}
}

func updateUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		UPDATE users
		SET role=$1
		WHERE uuid=$2
		`
		_, err = db.Exec(q, body["role"], body["userId"])
		if err != nil {
			http.Error(w, "error updating user role", http.StatusBadRequest)
			return
		}
		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "user role updated successfully")
	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		DELETE FROM users
		WHERE uuid=$1
		`
		_, err = db.Exec(q, body["userId"])
		if err != nil {
			http.Error(w, "error deleting user", http.StatusBadRequest)
			return
		}
		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "user deleted successfully")
	}
}
