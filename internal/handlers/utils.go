package handlers

import (
	"encoding/json"
	"net/http"
)

func decodeReqBody(r *http.Request) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
