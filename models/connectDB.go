package models

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres db driver for gorm
	"github.com/joho/godotenv"
)

var db *gorm.DB //database

func init() {
	e := godotenv.Load() // Load .env file

	if e != nil {
		fmt.Print(e)
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_password")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")

	// Build connection string
	dbURI := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", dbHost, dbPort, username, dbName, password)
	fmt.Println(dbURI)

	conn, err := gorm.Open("postgres", dbURI)
	if err != nil {
		fmt.Print(err)
		return
	}

	db = conn
	db.Debug().AutoMigrate(&User{}, &Session{}, &Signup{}) // Database migration
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
