package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// rows, err := db.Query("SELECT * FROM users")
// if err != nil {
// 	log.Fatal(err)
// }

// type User struct {
// 	uuid       string
// 	email      string
// 	hash       string
// 	username   sql.NullString
// 	avatar     string
// 	role       string
// 	created_at string
// }

// if rows.Next() {
// 	user := User{}
// 	if err := rows.Scan(&user.uuid, &user.email, &user.hash, &user.username, &user.avatar, &user.role, &user.created_at); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("%+v", user.username)
// }

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
