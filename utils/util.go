package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/NaySoftware/go-fcm"
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

// SendNotification to the mobile devices through firebase
func SendNotification(tokens []string, message string) {
	var serverKey = os.Getenv("firebase_server_key")

	data := map[string]string{
		"msg": message,
		"sum": "Snap-Back",
	}

	c := fcm.NewFcmClient(serverKey)
	c.NewFcmRegIdsMsg(tokens, data)
	// c.AppendDevices(xds)

	status, err := c.Send()

	if err == nil {
		status.PrintResults()
	} else {
		fmt.Println(err)
	}

}
