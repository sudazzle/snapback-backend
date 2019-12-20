package models

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
)

// Session - Trainning session
type Session struct {
	gorm.Model
	Title           string `json:"title"`
	UserID          uint   `json:"user_id"`
	MaxParticipants uint   `json:"max_participants"`
	DateNTime       string `json:"date_n_time"`
	Description     string `json:"description"`
	Status          string `json:"status"` // done, cancel, next
}

// Create new trainning session
func (session *Session) Create() (interface{}, string, int) {
	GetDB().Create(session)

	if session.ID <= 0 {
		return nil, "Failed to create session", http.StatusInternalServerError
	}

	return session, "", http.StatusOK
}

// UpdateSession updates single session
func UpdateSession(session Session, newData interface{}) (string, int) {
	err := GetDB().Model(&session).Updates(newData).Error
	if err != nil {
		return "Failed to update session.", http.StatusInternalServerError
	}

	return "Session updated", http.StatusOK
}

// GetSessionByID returns session
func GetSessionByID(id string) (interface{}, string, int) {
	session := &Session{}
	err := GetDB().Table("sessions").Where("id = ?", id).First(&session).Error
	fmt.Println(err)
	if err != nil {
		return nil, err.Error(), http.StatusInternalServerError
	}

	return session, "", http.StatusOK
}

// DeleteSession is only soft delete
func DeleteSession(id string) (string, int) {
	session := &Session{}
	ID, converterr := strconv.ParseUint(id, 10, 64)

	if converterr != nil {
		return "Id conversion error.", http.StatusInternalServerError
	}

	session.ID = uint(ID)

	tempSession := &Session{}
	errExists := GetDB().Table("sessions").Where("id = ? and status = ?", id, "done").First(tempSession).Error

	if errExists == nil {
		return "Session in done state. Session was successfully finished.", http.StatusForbidden
	}

	err := GetDB().Where("status=?", "next").Delete(session).Error

	if err != nil {
		return "Database error.", http.StatusInternalServerError
	}

	return "Session Deleted.", http.StatusOK
}
