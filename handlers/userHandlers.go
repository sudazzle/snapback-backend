package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"snapback/models"
	u "snapback/utils"

	"github.com/gorilla/mux"

	// "time"
	"context" // TODO read
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// ResetPassword sends email to the email address
var ResetPassword = func(w http.ResponseWriter, r *http.Request) {
	var requestPayload map[string]string

	err := json.NewDecoder(r.Body).Decode(&requestPayload)

	if err != nil {
		u.Respond(w, r, nil, "Invalid request.", "", http.StatusBadRequest)
		return
	}

	res, message, status := models.GetUser(requestPayload["email"], "email")
	if res == nil {
		u.Respond(w, r, nil, message, "", status)
		return
	}

	user := res.(*models.User)

	tk := &models.Token{UserID: user.ID, Role: user.Role}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("reset_password_token")))

	// variables to make PlainAuth compile, without adding unnecessary noise
	var (
		from       = "shrestha.sudaman@gmail.com"
		msg        = []byte("Click this link to change your password gmail.com?token=" + tokenString)
		recipients = []string{user.Email}
	)

	// hostname is used by PlainAuth to validate the TLS certificate
	hostname := "smtp.gmail.com"
	auth := smtp.PlainAuth("", "shrestha.sudaman@gmail.com", "Sud@zzle020219gmail", hostname)
	errEmailSend := smtp.SendMail(hostname+":587", auth, from, recipients, msg)
	if errEmailSend != nil {
		fmt.Println(err)
	}
}

// DeleteUserByAdmin disables user only soft delete
var DeleteUserByAdmin = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if IsCurrentUser(w, r, "admin") {
		message, status := models.DeleteUser(vars["id"])
		u.Respond(w, r, nil, message, "", status)
	}
}

// ChangePasswordByToken changes password on reset
var ChangePasswordByToken = func(w http.ResponseWriter, r *http.Request) {
	var requestPayload map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		u.Respond(w, r, nil, err.Error(), "", http.StatusBadRequest)
		return
	}

	currentUser := u.GetCurrentUser(r).(*models.Token)

	if requestPayload["password"] == "" {
		u.Respond(w, r, nil, "Invalid request. Password required.", "", http.StatusBadRequest)
		return
	}

	message, status := models.ChangeUserPassword(fmt.Sprint(currentUser.UserID), requestPayload["password"])
	u.Respond(w, r, nil, message, "", status)
}

// ChangeUserPasswordByAdmin changes password of the user. Admin and User roles are allowed
var ChangeUserPasswordByAdmin = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if IsCurrentUser(w, r, "admin") {
		var user map[string]string
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			u.Respond(w, r, nil, "Invalid request.", "", http.StatusBadRequest)
			return
		}

		if user["password"] == "" {
			u.Respond(w, r, nil, "Invalid request. Password required.", "", http.StatusBadRequest)
			return
		}

		if len(user["password"]) < 6 {
			u.Respond(w, r, nil, "Invalid request. Password length must be 6 or greater.", "", http.StatusBadRequest)
			return
		}

		message, status := models.ChangeUserPassword(vars["id"], user["password"])
		u.Respond(w, r, nil, message, "", status)
		return
	}
}

// ChangePasswordByUser changes password for current user
var ChangePasswordByUser = func(w http.ResponseWriter, r *http.Request) {
	currentUser := u.GetCurrentUser(r).(*models.Token)
	var user map[string]string
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		u.Respond(w, r, nil, "Invalid request.", "", http.StatusBadRequest)
		return
	}

	if user["password"] == "" {
		u.Respond(w, r, nil, "Invalid request. Password required.", "", http.StatusBadRequest)
		return
	}

	if len(user["password"]) < 6 {
		u.Respond(w, r, nil, "Invalid request. Password length must be 6 or greater.", "", http.StatusBadRequest)
		return
	}

	message, status := models.ChangeUserPassword(fmt.Sprint(currentUser.UserID), user["password"])
	u.Respond(w, r, nil, message, "", status)
}

// UpdateUser by id and by current user
var UpdateUser = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Println()
	var user map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&user)

	if user["password"] != nil {
		u.Respond(w, r, nil, "Invalid request. Password not allowed in json.", "", http.StatusBadRequest)
		return
	}

	if err != nil {
		u.Respond(w, r, nil, "Invalid request.", "", http.StatusBadRequest)
		return
	}

	if r.URL.Path == "/api/update-my-info" {
		currentUser := u.GetCurrentUser(r).(*models.Token)
		message, status := models.UpdateUser(fmt.Sprint(currentUser.UserID), user)
		u.Respond(w, r, nil, message, "", status)
	}

	if IsCurrentUser(w, r, "admin") && r.URL.Path == "/api/users/"+vars["id"] {
		message, status := models.UpdateUser(vars["id"], user)
		u.Respond(w, r, nil, message, "", status)
	}

}

// GetUser by id
var GetUser = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// payload, message, status := u.GetDefaultResponseData()

	if r.URL.Path == "/api/get-my-info" {
		currentUser := u.GetCurrentUser(r).(*models.Token)
		payload, message, status := models.GetUser(fmt.Sprint(currentUser.UserID), "id")
		u.Respond(w, r, payload, message, "user", status)
		return
	}

	if IsCurrentUser(w, r, "admin") && r.URL.Path == "/api/users/"+vars["id"] {
		fmt.Println("Got here")
		payload, message, status := models.GetUser(vars["id"], "id")
		u.Respond(w, r, payload, message, "user", status)
		return
	}

}

// GetUsers for admin user
var GetUsers = func(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	_, message, status := u.GetDefaultResponseData()

	if IsCurrentUser(w, r, "admin") {
		users, message, status = models.GetUsers()
		u.Respond(w, r, users, message, "users", status)
	}

}

// CreateUser - handler to register or create new user
var CreateUser = func(w http.ResponseWriter, r *http.Request) {
	// decode the request body into struct and fail if any error occur
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)

	payload, message, status := u.GetDefaultResponseData()

	if err != nil {
		message = "Invalid request."
		status = http.StatusBadRequest
	} else {
		payload, message, status = user.Create()
	}

	u.Respond(w, r, payload, message, "user", status)
}

// IsCurrentUser checks of the user role is role passed
var IsCurrentUser = func(w http.ResponseWriter, r *http.Request, role string) bool {
	user := u.GetCurrentUser(r).(*models.Token)
	if user.Role != role {
		return false
	}

	return true
}

// Authenticate user login
var Authenticate = func(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}

	err := json.NewDecoder(r.Body).Decode(user)
	payload, message, status := u.GetDefaultResponseData()

	if err != nil {
		message = "Invalid request"
		status = http.StatusBadRequest
	} else {
		payload, message, status = models.Login(user.Email, user.Password)
	}

	// if token != "" {
	// 	expire := time.Now().AddDate(0, 0, 1)
	// 	cookie := http.Cookie{
	// 		Name: "access_token",
	// 		Value: token,
	// 		Path: "/",
	// 		Expires: expire,
	// 	}

	// 	http.SetCookie(w, &cookie)
	// }
	u.Respond(w, r, payload, message, "user", status)
}

// JwtAuthentication handler is used to authenticate all the users with jwt token
var JwtAuthentication = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// List of endpoints that doesn't requre auth
		noAuth := []string{
			"/api/users/new",
			"/api/users/login",
			// "/api/getCsrf", // this is when backend and frontend are in different servers
			"/api/info",
			"/api/resetpassword",
		}

		// assetsTypes := []string {
		// 	".css",
		// 	".js",
		// 	".png",
		// 	".jpg",
		// 	".jpeg",
		// 	".svg",
		// 	".ico",
		// }

		requestPath := r.URL.Path // current request path

		// check if request is for static files
		// for _, value := range assetsTypes {
		// 	if strings.HasSuffix(requestPath, value) {
		// 		next.ServeHTTP(w, r)
		// 		return
		// 	}
		// }

		//check if request does not need authentication, serve the request if it doesn't need it
		for _, value := range noAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		tokenHeader := r.Header.Get("Authorization")

		// Handle missing token request and return error code 403 Unauthorized
		if tokenHeader == "" {
			u.Respond(w, r, nil, "Missing auth token.", "", http.StatusForbidden)
			return
		}

		// The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			u.Respond(w, r, nil, "Invalid/Malformed auth token", "", http.StatusForbidden)
			return
		}

		tokenPart := splitted[1] // Grab the token
		tk := &models.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			if requestPath == "/api/changepassword" {
				return []byte(os.Getenv("reset_password_token")), nil
			}
			return []byte(os.Getenv("token_password")), nil
		})

		if err != nil {
			u.Respond(w, r, nil, err.Error(), "", http.StatusForbidden)
			return
		}

		// Token is invalid, maybe not signed on this server
		if !token.Valid {
			u.Respond(w, r, nil, "Token is not valid.", "", http.StatusForbidden)
			return
		}

		const userKey u.UserKey = "user"

		// Everything all well, proceed with the request and set the caller to the user retrieved from the parsed token
		// fmt.Sprintf("User %v", tk.UserID) // Uuseful for monitoring
		ctx := context.WithValue(r.Context(), userKey, tk)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) // proceed to the middleware chain
	})
}
