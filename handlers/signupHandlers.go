package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"snapback/models"
	u "snapback/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// DoSignup - Signup to the snapback trainning session
var DoSignup = func(w http.ResponseWriter, r *http.Request) {
	signup := &models.Signup{}

	user := u.GetCurrentUser(r).(*models.Token)
	signup.UserID = user.UserID

	err := json.NewDecoder(r.Body).Decode(signup)

	payload, message, status := u.GetDefaultResponseData()

	curSession, _, _ := models.GetSessionByID(fmt.Sprint(signup.SessionID))

	// @TODO get max participants from the table
	if signup.GetSignupCounts() > 5 {
		signup.Status = "waiting"
	} else {
		signup.Status = "priority"
	}

	// count the number of signups for the particular session ()
	if err != nil {
		message = "Invalid request"
		status = http.StatusBadRequest
	} else if curSession.(*models.Session).UserID == user.UserID {
		message = "Not allowed to signup to your own session."
		status = http.StatusForbidden
	} else if signup.HasUserSignedupAlready() {
		if signup.IsReSigning() {
			payload, message, status = signup.Resignup()
		} else {
			message = "Participant already signed up."
			status = http.StatusBadRequest
		}
	} else {
		payload, message, status = signup.DoSignup()
	}

	u.Respond(w, r, payload, message, "signup", status)
}

// CancelSignup to cancel the signup
var CancelSignup = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID, converr := strconv.ParseUint(vars["id"], 10, 64)

	if converr != nil {
		u.Respond(w, r, nil, converr.Error(), "", http.StatusInternalServerError)
		return
	}

	signup := &models.Signup{}
	signup.ID = uint(ID)

	user := u.GetCurrentUser(r).(*models.Token)

	delerr := models.GetDB().Where("user_id = ?", user.UserID).Delete(signup).Error

	if delerr != nil {
		u.Respond(w, r, nil, delerr.Error(), "", http.StatusInternalServerError)
		return
	}

	u.Respond(w, r, nil, "Signup cancelled", "", http.StatusOK)
}

// GetNextSignups returns the users signups for next trainnings
var GetNextSignups = func(w http.ResponseWriter, r *http.Request) {
	user := u.GetCurrentUser(r).(*models.Token)
	// Scan
	type Result struct {
		ID              uint   `json:"signup_id"`
		SessionID       uint   `json:"session_id"`
		Title           string `json:"title"`
		Description     string `json:"description"`
		DateNTime       string `json:"date_n_time"`
		MaxParticipants int    `json:"max_participants"`
		Status          string `json:"status"`
	}

	rows, err := models.GetDB().Raw("SELECT signups.id, signups.session_id, sessions.title, sessions.description, sessions.date_n_time, sessions.max_participants, signups.status FROM sessions INNER JOIN signups ON sessions.id = signups.session_id WHERE signups.user_id = ? and sessions.status = 'next' and signups.deleted_at is null", user.UserID).Rows()
	defer rows.Close()
	if err != nil {
		u.Respond(w, r, nil, err.Error(), "", http.StatusInternalServerError)
	}

	var payload []Result
	for rows.Next() {
		var result Result
		models.GetDB().ScanRows(rows, &result)
		payload = append(payload, result)
	}
	u.Respond(w, r, payload, "", "signups", http.StatusOK)
}

// GetSignupsBySessionID returns users signed up for session
var GetSignupsBySessionID = func(w http.ResponseWriter, r *http.Request) {
	user := u.GetCurrentUser(r).(*models.Token)
	if user.Role == "admin" || user.Role == "trainer" {
		vars := mux.Vars(r)
		session, message, status := models.GetSessionByID(vars["id"])

		if status != 200 {
			u.Respond(w, r, nil, message, "", status)
			return
		}

		if user.Role == "trainer" && session.(*models.Session).UserID != user.UserID {
			u.Respond(w, r, nil, "Unauthorized", "", http.StatusForbidden)
			return
		}

		type Result struct {
			ID       uint   `json:"user_id"`
			Name     string `json:"user_name"`
			SignupID uint   `json:"signup_id"`
		}

		rows, err := models.GetDB().Raw("SELECT users.id, users.name, signups.id as signup_id FROM users INNER JOIN signups ON users.id = signups.user_id WHERE signups.session_id = ? and signups.status = 'priority'", vars["id"]).Rows()
		defer rows.Close()

		if err != nil {
			u.Respond(w, r, nil, err.Error(), "", http.StatusInternalServerError)
		}

		var payload []Result
		for rows.Next() {
			var result Result
			models.GetDB().ScanRows(rows, &result)
			payload = append(payload, result)
		}

		u.Respond(w, r, payload, "", "signups", http.StatusOK)

	} else {
		u.Respond(w, r, nil, "Unauthorized", "", http.StatusForbidden)
	}
}

// DoAttendence - inserting into attendence field based on signup id
var DoAttendence = func(w http.ResponseWriter, r *http.Request) {
	user := u.GetCurrentUser(r).(*models.Token)
	if user.Role == "admin" || user.Role == "trainer" {
		vars := mux.Vars(r)
		session, message, status := models.GetSessionByID(vars["id"])

		if status != 200 {
			u.Respond(w, r, nil, message, "", status)
			return
		}

		if user.Role == "trainer" && session.(*models.Session).UserID != user.UserID {
			u.Respond(w, r, nil, "Unauthorized", "", http.StatusForbidden)
			return
		}

		var attendees []uint
		decodeerr := json.NewDecoder(r.Body).Decode(&attendees)

		if decodeerr != nil {
			u.Respond(w, r, nil, decodeerr.Error(), "", http.StatusBadRequest)
			return
		}

		updateerr := models.GetDB().Table("signups").Where("id IN (?)", attendees).Update("attendence", true).Error
		_, sessionUpdateErr := SetSessionDone(session.(*models.Session))

		if sessionUpdateErr != nil {
			u.Respond(w, r, nil, sessionUpdateErr.Error(), "", http.StatusInternalServerError)
			return
		}

		if updateerr != nil {
			u.Respond(w, r, nil, updateerr.Error(), "", http.StatusInternalServerError)
			return
		}

		u.Respond(w, r, nil, "Attendence Done!", "", http.StatusOK)

	} else {
		u.Respond(w, r, nil, "Unauthorized", "", http.StatusForbidden)
	}

}
