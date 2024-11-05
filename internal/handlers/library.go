package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func LibrayRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Post("/", getLibraryEntries(db))
	r.Put("/", createNewEntry(db))

	return r
}

func getLibraryEntries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := `
		SELECT *
		FROM library
		`
		rows, err := db.Query(q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		type Row struct {
			Id            string `json:"id"`
			Title         string `json:"title"`
			YearPublished int16  `json:"year_published"`
			ImageURL      string `json:"image_url"`
		}
		titles := []Row{}
		for rows.Next() {
			var r Row
			err := rows.Scan(&r.Id, &r.Title, &r.YearPublished, &r.ImageURL)
			if err != nil {
				http.Error(w, "error retrieving library entries", http.StatusInternalServerError)
				return
			}
			titles = append(titles, r)
		}
		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(titles)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
	}
}

func createNewEntry(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		SELECT id
		FROM library
		WHERE id=$1
		`
		row := db.QueryRow(q, body["id"])
		if err = row.Scan(); err == nil {
			http.Error(w, "title already exists", http.StatusBadRequest)
			return
		}

		q = `
		INSERT INTO library (id, title, year_published, image_url)
		VALUES ($1, $2, $3, $4)
		`
		_, err = db.Exec(q, body["id"], body["title"], body["year_published"], body["image_url"])
		if err != nil {
			http.Error(w, "errro creating new entry", http.StatusInternalServerError)
			return
		}

		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "entry created successfully")
	}
}
