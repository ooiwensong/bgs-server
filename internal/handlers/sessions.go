package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/ooiwensong/bgs_server/internal/middlewares"
)

type Sessions struct {
	Uuid        string         `json:"uuid"`
	HostId      string         `json:"host_id"`
	GameTitle   string         `json:"game_title"`
	MaxGuests   int8           `json:"max_guests"`
	NumGuests   int8           `json:"num_guests"`
	Date        string         `json:"date"`
	StartTime   string         `json:"start_time"`
	EndTime     string         `json:"end_time"`
	Address     string         `json:"address"`
	IsFull      bool           `json:"is_full"`
	ExpiresAt   string         `json:"expires_at"`
	CreatedAt   string         `json:"created_at"`
	LastUpdated *string        `json:"last_updated"`
	GameImage   *string        `json:"game_image"`
	Guests      pq.StringArray `json:"guests"`
}

func SessionsRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Auth)

	r.Post("/", getSingleSession(db))
	r.Put("/", createSession(db))
	r.Patch("/{sessionId}", editSession(db))
	r.Delete("/", deleteSession(db))

	r.Post("/my-sessions", getMySessions(db))
	r.Post("/host-sessions", getMyHostSessions(db))
	r.Post("/other-user-sessions", getOtherUserSessions(db))
	r.Post("/join-session", joinSession(db))
	r.Post("/leave-session", leaveSession(db))

	return r
}

func getSingleSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "error retrieving session data", http.StatusBadRequest)
		}
		defer tx.Rollback()

		q := `
		SELECT *
		FROM sessions
		WHERE uuid=$1
		`
		row := tx.QueryRow(q, body["sessionId"])
		type SessionData struct {
			Uuid        string  `json:"uuid"`
			HostId      string  `json:"host_id"`
			GameTitle   string  `json:"game_title"`
			MaxGuests   int8    `json:"max_guests"`
			NumGuests   int8    `json:"num_guests"`
			Date        string  `json:"date"`
			StartTime   string  `json:"start_time"`
			EndTime     string  `json:"end_time"`
			Address     string  `json:"address"`
			IsFull      bool    `json:"is_full"`
			ExpiresAt   string  `json:"expires_at"`
			CreatedAt   string  `json:"created_at"`
			LastUpdated *string `json:"last_updated"`
			GameImage   *string `json:"game_image"`
		}
		var s SessionData
		err = row.Scan(&s.Uuid, &s.HostId, &s.GameTitle, &s.MaxGuests, &s.NumGuests, &s.Date, &s.StartTime, &s.EndTime, &s.Address, &s.IsFull, &s.ExpiresAt, &s.CreatedAt, &s.LastUpdated, &s.GameImage)
		if err != nil {
			http.Error(w, "error retrieving session data", http.StatusBadRequest)
			return
		}

		q = `
		SELECT guest_id, date_joined
		FROM guests
		WHERE guests.session_id=$1
		`
		rows, err := tx.Query(q, body["sessionId"])
		if err != nil {
			http.Error(w, "error retrieving session data", http.StatusBadRequest)
		}
		type SessionGuests struct {
			GuestId    string `json:"guest_id"`
			DateJoined string `json:"date_joined"`
		}
		guests := []SessionGuests{}
		for rows.Next() {
			var sg SessionGuests
			err = rows.Scan(&sg.GuestId, &sg.DateJoined)
			if err != nil {
				http.Error(w, "error retrieving session data", http.StatusBadRequest)
				return
			}
			guests = append(guests, sg)
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "error retrieving session data", http.StatusBadRequest)
			return
		}

		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(struct {
			SessionData   SessionData     `json:"sessionData"`
			SessionGuests []SessionGuests `json:"sessionGuests"`
		}{
			s,
			guests,
		})
		if err != nil {
			http.Error(w, "error retrieving session data", http.StatusInternalServerError)
		}
	}
}

func createSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		userId := r.Context().Value("decoded").(*middlewares.Claims).UserId
		if userId != body["userId"] {
			http.Error(w, "not authorised", http.StatusUnauthorized)
			return
		}

		q := `
		INSERT INTO sessions (host_id, game_title, max_guests, date, start_time, end_time, address, game_image)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = db.Exec(q, body["userId"], body["game_title"], body["max_guests"], body["date"], body["start_time"], body["end_time"], body["address"], body["game_image"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "session created successfully")
	}
}

func editSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, "error updating session", http.StatusBadRequest)
			return
		}

		sessionId := chi.URLParam(r, "sessionId")

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "error updating session", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		q := `
		SELECT host_id
		FROM sessions
		WHERE uuid=$1
		`
		row := tx.QueryRow(q, sessionId)
		var hostId string
		if err = row.Scan(&hostId); err != nil {
			http.Error(w, "error updating session", http.StatusBadRequest)
			return
		}

		userId := r.Context().Value("decoded").(*middlewares.Claims).UserId
		if hostId != userId {
			http.Error(w, "not authorised", http.StatusUnauthorized)
			return
		}

		for attribute, value := range body {
			q := fmt.Sprintf(`
			UPDATE sessions
			SET %s=$1
			WHERE uuid=$2
			`, attribute)
			if _, err = tx.Exec(q, value, sessionId); err != nil {
				http.Error(w, "error updating session", http.StatusBadRequest)
				return
			}
		}
		if err = tx.Commit(); err != nil {
			http.Error(w, "error updating session", http.StatusBadRequest)
			return
		}

		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "session updated successfully")
	}
}

func deleteSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		q := `
		SELECT host_id
		FROM sessions
		WHERE uuid=$1
		`
		row := tx.QueryRow(q, body["sessionId"])
		var hostId string
		err = row.Scan(&hostId)
		if err != nil {
			http.Error(w, "error deleting session", http.StatusBadRequest)
			return
		}

		userId := r.Context().Value("decoded").(*middlewares.Claims).UserId
		if hostId != userId {
			http.Error(w, "not authorised", http.StatusBadRequest)
			return
		}

		q = `
		DELETE FROM sessions
		WHERE uuid=$1
		`
		_, err = tx.Exec(q, body["sessionId"])
		if err != nil {
			http.Error(w, "error deleting session", http.StatusBadRequest)
			return
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "error deleting session", http.StatusInternalServerError)
		}

		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "session deleted successfully")
	}
}

func getMySessions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		SELECT s.uuid, s.host_id, s.game_title, s.max_guests, s.num_guests, s.date, s.start_time, s.end_time, s.address, s.is_full, s.expires_at, s.created_at, s.last_updated, s.game_image, array_agg(distinct guests.guest_id) filter (where guests.guest_id is not null) "guests"
		FROM sessions AS s
		FULL JOIN guests ON s.uuid = guests.session_id
		WHERE host_id=$1 OR guest_id=$2
		GROUP BY s.uuid
		ORDER BY date, start_time
		`
		rows, err := db.Query(q, body["userId"], body["userId"])
		if err != nil {
			http.Error(w, "error retrieving sessions", http.StatusBadRequest)
			return
		}
		sessions := []Sessions{}
		for rows.Next() {
			var s Sessions
			err := rows.Scan(&s.Uuid, &s.HostId, &s.GameTitle, &s.MaxGuests, &s.NumGuests, &s.Date, &s.StartTime, &s.EndTime, &s.Address, &s.IsFull, &s.ExpiresAt, &s.CreatedAt, &s.LastUpdated, &s.GameImage, &s.Guests)
			if err != nil {
				http.Error(w, "error retrieving sessions", http.StatusBadRequest)
				return
			}
			sessions = append(sessions, s)
		}
		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(sessions)
		if err != nil {
			http.Error(w, "error retrieving sessions", http.StatusInternalServerError)
		}
	}
}

func getMyHostSessions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		SELECT s.uuid, s.host_id, s.game_title, s.max_guests, s.num_guests, s.date, s.start_time, s.end_time, s.address, s.is_full, s.expires_at, s.created_at, s.last_updated, s.game_image, array_agg(distinct guests.guest_id) filter (where guests.guest_id is not null) "guests"
		FROM sessions AS s
		FULL JOIN guests ON s.uuid = guests.session_id
		WHERE host_id=$1
		GROUP BY s.uuid
		ORDER BY date, start_time
		`
		rows, err := db.Query(q, body["userId"])
		if err != nil {
			http.Error(w, "error retrieving sessions", http.StatusBadRequest)
			return
		}
		sessions := []Sessions{}
		for rows.Next() {
			var s Sessions
			err := rows.Scan(&s.Uuid, &s.HostId, &s.GameTitle, &s.MaxGuests, &s.NumGuests, &s.Date, &s.StartTime, &s.EndTime, &s.Address, &s.IsFull, &s.ExpiresAt, &s.CreatedAt, &s.LastUpdated, &s.GameImage, &s.Guests)
			if err != nil {
				http.Error(w, "error retrieving sessions", http.StatusBadRequest)
				return
			}
			sessions = append(sessions, s)
		}
		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(sessions)
		if err != nil {
			http.Error(w, "error retrieving sessions", http.StatusInternalServerError)
		}
	}
}

func getOtherUserSessions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		q := `
		SELECT s.uuid, s.host_id, s.game_title, s.max_guests, s.num_guests, s.date, s.start_time, s.end_time, s.address, s.is_full, s.expires_at, s.created_at, s.last_updated, s.game_image, array_agg(distinct guests.guest_id) filter (where guests.guest_id is not null) "guests"
		FROM sessions AS s
		FULL JOIN guests ON s.uuid = guests.session_id
		WHERE host_id != $1
		GROUP BY s.uuid
		ORDER BY date, start_time
		`
		rows, err := db.Query(q, body["userId"])
		if err != nil {
			http.Error(w, "error retrieving sessions", http.StatusBadRequest)
			return
		}
		sessions := []Sessions{}
		for rows.Next() {
			var s Sessions
			err := rows.Scan(&s.Uuid, &s.HostId, &s.GameTitle, &s.MaxGuests, &s.NumGuests, &s.Date, &s.StartTime, &s.EndTime, &s.Address, &s.IsFull, &s.ExpiresAt, &s.CreatedAt, &s.LastUpdated, &s.GameImage, &s.Guests)
			if err != nil {
				http.Error(w, "error retrieving sessions", http.StatusBadRequest)
				return
			}
			sessions = append(sessions, s)
		}
		generate200Response(w, http.StatusOK)
		err = json.NewEncoder(w).Encode(sessions)
		if err != nil {
			http.Error(w, "error retrieving sessions", http.StatusInternalServerError)
		}
	}
}

func leaveSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userId := r.Context().Value("decoded").(*middlewares.Claims).UserId
		if userId != body["userId"] {
			http.Error(w, "not authorised", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "error leaving session", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		q := `
		DELETE FROM guests
		WHERE session_id=$1
		AND guest_id=$2
		`
		_, err = tx.Exec(q, body["sessionId"], body["userId"])
		if err != nil {
			http.Error(w, "error leaving session", http.StatusBadRequest)
			return
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "error leaving session", http.StatusBadRequest)
			return
		}
		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "left session successfully")
	}
}

func joinSession(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decodeReqBody(r)
		if err != nil {
			http.Error(w, "error joining session", http.StatusBadRequest)
			return
		}
		userId := r.Context().Value("decoded").(*middlewares.Claims).UserId
		if userId != body["userId"] {
			http.Error(w, "not authorised", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "error joining session", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		q := `
		SELECT host_id, is_full
		FROM sessions
		WHERE uuid=$1
		`
		row := tx.QueryRow(q, body["sessionId"])
		var hostId string
		var isFull bool
		if err = row.Scan(&hostId, &isFull); err != nil {
			http.Error(w, "error joining session", http.StatusBadRequest)
			return
		}
		if hostId == body["userId"] {
			http.Error(w, "cannot join a session you are hosting", http.StatusBadRequest)
			return
		}
		if isFull {
			http.Error(w, "cannot join a session as it is full", http.StatusBadRequest)
			return
		}

		q = `
		INSERT INTO guests (session_id, guest_id)
		VALUES ($1, $2)
		`
		if _, err = tx.Exec(q, body["sessionId"], body["userId"]); err != nil {
			http.Error(w, "error joining session", http.StatusBadRequest)
			return
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "error joining session", http.StatusInternalServerError)
			return
		}

		generate200Response(w, http.StatusOK)
		generateOKJsonMsg(w, "session joined successfully")
	}
}
