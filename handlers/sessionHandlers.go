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

// @TODO need some refactor move sql query to model

// CreateSession User - with roles admin or trainer can create session
var CreateSession = func(w http.ResponseWriter, r *http.Request) {
	session := &models.Session{}
	payload, message, status := u.GetDefaultResponseData()
	user := u.GetCurrentUser(r).(*models.Token)
	// t, _ := time.Parse("2006-01-02T15:04:05-0200", "2019-10-08T20:46:54-0200")
	err := json.NewDecoder(r.Body).Decode(session)

	if err != nil {
		message = err.Error()
		status = http.StatusBadRequest
	} else if user.Role != "admin" && user.Role != "trainer" {
		message = "Unauthorized."
		status = http.StatusForbidden
	} else {
		session.UserID = user.UserID
		if session.Status == "" {
			session.Status = "next"
		}

		tokens := models.GetTokens(user.UserID)
		u.SendNotification(tokens, "New trainning session open for signups.")

		payload, message, status = session.Create()
	}

	u.Respond(w, r, payload, message, "session", status)
}

// GetNextSessions - returns the upcoming snapback trainning sessions
var GetNextSessions = func(w http.ResponseWriter, r *http.Request) {
	var sessions []*models.Session
	payload, message, status := u.GetDefaultResponseData()
	err := models.GetDB().Table("sessions").Where("status = ?", "next").Find(&sessions).Error

	if err != nil {
		message = "Could not perform the request"
		status = http.StatusInternalServerError
	} else {
		payload = sessions
	}
	u.Respond(w, r, payload, message, "sessions", status)
}

// GetAllSessions to get all the sessions for admin and trainers only
// trainers get their sessions only
var GetAllSessions = func(w http.ResponseWriter, r *http.Request) {
	limit, convErrLimit := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	page, convErrPage := strconv.ParseInt(r.FormValue("page"), 10, 64)

	if convErrLimit != nil {
		limit = -1
	}

	if convErrPage != nil {
		page = 1
	}

	var sessions []*models.Session
	payload, message, status := u.GetDefaultResponseData()
	user := u.GetCurrentUser(r).(*models.Token)
	var err error

	actualOffset := (page - 1) * limit

	if IsCurrentUser(w, r, "admin") {
		fmt.Println(actualOffset)
		err = models.GetDB().Order("id desc").Limit(limit).Table("sessions").Offset(actualOffset).Find(&sessions).Error
		payload = sessions
	} else if IsCurrentUser(w, r, "trainer") {
		err = models.GetDB().Order("id desc").Limit(limit).Table("sessions").Where("user_id = ?", user.UserID).Offset(actualOffset).Find(&sessions).Error
		payload = sessions
	}

	if err != nil {
		message = err.Error()
		status = http.StatusInternalServerError
	}

	u.Respond(w, r, payload, message, "sessions", status)
}

// UpdateSession to update the trainning session
var UpdateSession = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := u.GetCurrentUser(r).(*models.Token)

	var newData map[string]interface{}

	ID, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		u.Respond(w, r, nil, "id conversion error", "", http.StatusInternalServerError)
	}

	if user.Role != "admin" && user.Role != "trainer" {
		u.Respond(w, r, nil, "Not allowed", "", http.StatusForbidden)
	}

	session := models.Session{}
	session.ID = uint(ID)

	errDecode := json.NewDecoder(r.Body).Decode(&newData)

	if errDecode != nil {
		u.Respond(w, r, nil, "Invalid request.", "", http.StatusBadRequest)
	}

	models.UpdateSession(session, newData)

}

// GetSessionByID returns session detail
var GetSessionByID = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := u.GetCurrentUser(r).(*models.Token)

	if user.Role != "admin" && user.Role != "trainer" {
		u.Respond(w, r, nil, "Not allowed", "", http.StatusForbidden)
		return
	}

	payload, message, status := models.GetSessionByID(vars["id"])
	u.Respond(w, r, payload, message, "session", status)
}

// DeleteSession softdelete
var DeleteSession = func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if IsCurrentUser(w, r, "admin") || IsCurrentUser(w, r, "trainer") {
		message, status := models.DeleteSession(vars["id"])
		u.Respond(w, r, nil, message, "", status)
	} else {
		u.Respond(w, r, nil, "Unauthorized", "", http.StatusForbidden)
	}

}

// SetSessionDone update status to done
var SetSessionDone = func(session *models.Session) (bool, error) {
	err := models.GetDB().Model(&session).Update("status", "done").Error

	if err != nil {
		return false, err
	}

	return true, nil
}

// GetSessionCount all sessions count
var GetSessionCount = func(w http.ResponseWriter, r *http.Request) {
	var count int
	err := models.GetDB().Table("sessions").Where("deleted_at ISNULL").Count(&count).Error

	fmt.Println(count)

	if err != nil {
		u.Respond(w, r, nil, err.Error(), "", http.StatusInternalServerError)
	}

	u.Respond(w, r, count, "", "count", http.StatusOK)
}
