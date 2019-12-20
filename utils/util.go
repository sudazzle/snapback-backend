package utils

import (
	"encoding/json"
	"net/http"
)

// UserKey type for context
type UserKey string

// GetDefaultResponseData returns the default response data
func GetDefaultResponseData() (interface{}, string, int) {
	return nil, "", http.StatusOK
}

// GetCurrentUser from the context
func GetCurrentUser(r *http.Request) interface{} {
	const userKey UserKey = "user"
	user := r.Context().Value(userKey)
	return user
}

// Respond - utility function to respond all the http requests
func Respond(w http.ResponseWriter, r *http.Request, data interface{}, message string, key string, status int) {
	payload := make(map[string]interface{})

	if message != "" {
		payload["message"] = message
	}

	if data != nil && key != "" {
		payload[key] = data
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
