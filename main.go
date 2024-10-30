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

}
