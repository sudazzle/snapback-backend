package models

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// Token JWT claims structs
type Token struct {
	UserID             uint
	Role               string
	jwt.StandardClaims // https://godoc.org/github.com/dgrijalva/jwt-go#StandardClaims
}

// Struct representing users table

// What is backtick here?
/*
You can add extra meta information to Go struct in the form of tags.
Here, the json:"somekey" is used by the json package to encode the values
of Somekey into the key some "somekey" in the corresponding json object
*/

/*
gorm.Model is a basic GoLang struct which includes the following fields:
ID, CreatedAt, UpdatedAt, DeletedAt.

It may be embeded into your model or you may build your own model without it.
http://gorm.io/docs/models.html
*/

// User struct
type User struct {
	gorm.Model
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	DeviceToken string `json:"device_token"`
	Token       string `json:"token" sql:"-"` // sql:"-" means ignore this field in db
}

/*
Methods:
In between the keyword func and the name of the function we add a receiver.
The receiver is like a parameter - it has a name and a type - but by creating the function in this way
it allows us to call the fuction us the (.) operator
*/

// Validate incoming user details
func (user *User) Validate() (string, int) {
	if !strings.Contains(user.Email, "@") {
		return "Email address is required", http.StatusBadRequest
	}

	if user.Password == "" {
		return "Password is required!", http.StatusBadRequest
	}

	if len(user.Password) < 6 {
		return "Password length should be greater than 6.", http.StatusBadRequest
	}

	// Email must be unique
	temp := &User{}

	// check for erros and duplicate emails
	err := GetDB().Table("users").Where("email = ?", user.Email).First(temp).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "Connection error.", http.StatusInternalServerError
	}

	if temp.Email != "" {
		return "Email address already in use by another user.", http.StatusBadRequest
	}

	return "", http.StatusOK
}

// Create User
func (user *User) Create() (interface{}, string, int) {
	message, status := user.Validate()

	if message != "" && status != http.StatusOK {
		return nil, message, status
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	if user.Role == "" {
		user.Role = "user"
	}

	GetDB().Create(user)

	if user.ID <= 0 {
		return nil, "Failed to create user, connection error.", http.StatusInternalServerError
	}

	// Create new JWT token for the newly registered user
	tk := &Token{UserID: user.ID, Role: user.Role}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	user.Token = tokenString

	user.Password = ""

	return user, "", http.StatusOK
}

// Login user excepts email and password
func Login(email, password string) (interface{}, string, int) {
	user := &User{}

	// Get first record
	err := GetDB().Table("users").Where("email = ?", email).First(user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "Email address not found", http.StatusNotFound
		}

		return nil, "Connection error. Please retry", http.StatusInternalServerError
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, "Invalid login credintials", http.StatusBadRequest
	}

	user.Password = ""

	// Generate JWT token
	tk := &Token{UserID: user.ID, Role: user.Role}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	user.Token = tokenString

	return user, "Logged in.", http.StatusOK
}

// GetUser returns user by id or email
func GetUser(val string, by string) (interface{}, string, int) {
	user := &User{}
	err := GetDB().Table("users").Where(by+" = ?", val).First(user).Error

	if err != nil {
		return nil, err.Error(), http.StatusInternalServerError
	}

	if user.ID < 0 {
		return nil, "Database error.", http.StatusInternalServerError
	}

	user.Password = ""
	return user, "", http.StatusOK
}

// GetUsers returns list of users
func GetUsers(limit int64, offset int64) ([]User, string, int) {
	var users []User
	// err := GetDB().Select("id, email, name, role").Where("id >= ?", start).Limit(limit).Order("id").Find(&users).Error
	err := GetDB().Table("users").Order("id desc").Offset(offset).Limit(limit).Find(&users).Error

	if err != nil {
		return nil, "Database error", http.StatusInternalServerError
	}

	return users, "", http.StatusOK
}

// GetTokens returns the list of firebase device tokens
func GetTokens(userID uint, sessionID interface{}) []string {
	var tokens []string
	var rows *sql.Rows
	var err error

	fmt.Print(userID)

	if sessionID == nil {
		rows, err = GetDB().Table("users").Select("device_token").Where("device_token IS NOT NULL AND ID <> ?", userID).Rows()
	} else {
		rows, err = GetDB().Raw("SELECT u.device_token FROM users u INNER JOIN signups s ON u.ID = s.user_id WHERE u.device_token IS not NULL and s.session_id = ? and s.deleted_at IS NULL and s.status='waiting' and u.ID <> ?", sessionID, userID).Rows()
	}

	defer rows.Close()
	if err == nil {
		for rows.Next() {
			var token string
			rows.Scan(&token)
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// UpdateUser updates the user info
func UpdateUser(id string, newData map[string]interface{}) (string, int) {
	user := &User{}
	ID, converterr := strconv.ParseUint(id, 10, 64)

	if converterr != nil {
		return "Id conversion error.", http.StatusInternalServerError
	}

	user.ID = uint(ID)
	err := GetDB().Model(user).Updates(newData).Error

	if err != nil {
		return "Database error.", http.StatusInternalServerError
	}

	return "Update Successful.", http.StatusOK
}

// ChangeUserPassword chages password
func ChangeUserPassword(id string, password string) (string, int) {
	fmt.Println(id)
	fmt.Println(password)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &User{}
	ID, converterr := strconv.ParseUint(id, 10, 64)

	if converterr != nil {
		return "Id conversion error.", http.StatusInternalServerError
	}

	user.ID = uint(ID)

	err := GetDB().Model(user).Update("password", string(hashedPassword)).Error

	if err != nil {
		return "Database error.", http.StatusInternalServerError
	}

	return "Password Updated.", http.StatusOK
}

// DeleteUser is only soft delete
func DeleteUser(id string) (string, int) {
	user := &User{}
	ID, converterr := strconv.ParseUint(id, 10, 64)

	if converterr != nil {
		return "Id conversion error.", http.StatusInternalServerError
	}

	user.ID = uint(ID)

	err := GetDB().Delete(user).Error

	if err != nil {
		return "Database error.", http.StatusInternalServerError
	}

	return "User Deleted.", http.StatusOK
}
