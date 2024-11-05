package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func decodeReqBody(r *http.Request) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func generate200Response(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
}

func generateClaims(userId, email, role, avatar string, username *string) Claims {
	return Claims{
		userId,
		email,
		role,
		username,
		avatar,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}
}

func generateOKJsonMsg(w http.ResponseWriter, msg string) {
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
	}{
		"ok",
		msg,
	})
}
