package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func AuthRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", login(db))
	r.Put("/register", register(db))
	r.Post("/refresh", refresh())

	return r
}

type Claims struct {
	UserId   string  `json:"userId"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	Username *string `json:"username"`
	Avatar   string  `json:"avatar"`
	jwt.RegisteredClaims
}

func login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		SELECT *
		FROM users
		WHERE email=$1
		`
		row := db.QueryRow(q, body["email"])
		var (
			userId    string
			email     string
			hash      string
			username  *string
			avatar    string
			role      string
			createdAt string
		)
		err = row.Scan(&userId, &email, &hash, &username, &avatar, &role, &createdAt)
		if err != nil {
			http.Error(w, "email or password is incorrect", http.StatusForbidden)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(body["password"].(string)))
		if err != nil {
			http.Error(w, "email or password is incorrect", http.StatusForbidden)
			return
		}

		claims := generateClaims(userId, email, role, avatar, username)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		accessToken, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
		if err != nil {
			// TODO: log server error
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		refreshToken, err := token.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
		if err != nil {
			// TODO: log server error
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		}{
			accessToken,
			refreshToken,
		})
		if err != nil {
			http.Error(w, "login unsuccessful", http.StatusInternalServerError)
		}
	}
}

func register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		SELECT *
		FROM users
		WHERE email=$1
		`
		rows, err := db.Query(q, body["email"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if rows.Next() {
			http.Error(w, "email already exists", http.StatusBadRequest)
			return
		} else if err = rows.Err(); err != nil {
			// error encountered while preparing the next result row
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(body["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		q = `
		INSERT INTO users (email, hash, username)
		VALUES ($1, $2, $3)
		`
		_, err = db.Exec(q, body["email"], hash, body["username"])
		if err != nil {
			http.Error(w, "error creating user", http.StatusBadRequest)
			return
		}

		generate200Response(w, http.StatusCreated)
		json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
			Msg    string `json:"msg"`
		}{
			"ok",
			"user created successfully",
		})
	}
}

func refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		token, err := jwt.ParseWithClaims(body["refresh"].(string), &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("REFRESH_SECRET")), nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok {
			claims := generateClaims(claims.UserId, claims.Email, claims.Role, claims.Avatar, claims.Username)
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			accessToken, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
			if err != nil {
				http.Error(w, "something went wrong", http.StatusInternalServerError)
				return
			}
			generate200Response(w, http.StatusOK)
			json.NewEncoder(w).Encode(struct {
				AccessToken string `json:"accessToken"`
			}{
				accessToken,
			})
		} else {
			// TODO: log server error
			http.Error(w, "something went wrong", http.StatusBadRequest)
			return
		}
	}
}
