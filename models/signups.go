package models

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// Signup struct
type Signup struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	SessionID  uint   `json:"session_id"`
	QueueNo    uint   `json:"queue_no"`
	Status     string `json:"status"` // priority / waiting and soft delete means cancel
	Attendence bool   `json:"attendence"`
}

// GetSignupByID returns signup by ID
func (signup *Signup) GetSignupByID() interface{} {
	err := GetDB().First(&signup, signup.ID).Error

	if err == nil {
		return signup
	}

	return nil
}

// GetSignupCounts - returns the number of participants signed up for the trainning session
func (signup *Signup) GetSignupCounts() int {
	var count int
	GetDB().Table("signups").Where("deleted_at is null and session_id = ?", signup.SessionID).Count(&count)
	return count
}

// HasUserSignedupAlready - check if the user has already signed up
func (signup *Signup) HasUserSignedupAlready() bool {
	var count int
	GetDB().Table("signups").Where("deleted_at is null and session_id = ? and user_id = ?", signup.SessionID, signup.UserID).Count(&count)

	if count >= 1 {
		return true
	}

	return false
}

// IsReSigning - check if delete_at is null
func (signup *Signup) IsReSigning() bool {
	var count int
	GetDB().Table("signups").Where("not deleted_at is null and session_id = ? and user_id = ?", signup.SessionID, signup.UserID).Count(&count)

	if count >= 1 {
		return true
	}

	return false

}

// Resignup - set delete_at to null
func (signup *Signup) Resignup() (map[string]interface{}, string, int) {
	err := GetDB().Model(&signup).Update("deleted_at", "null").Error

	if err != nil {
		return nil, err.Error(), http.StatusInternalServerError
	}

	return nil, "Signed up", http.StatusOK
}

// DoSignup - insert the signup request
func (signup *Signup) DoSignup() (map[string]interface{}, string, int) {
	GetDB().Create(signup)

	if signup.ID <= 0 {
		return nil, "Failed to create user.", http.StatusInternalServerError
	}

	response := make(map[string]interface{})
	response["signup"] = signup
	return response, "", http.StatusOK
}

// CheckAttendence - inserting true or false based on signupid
