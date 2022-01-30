package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

type dbCredentials struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// SetupDatabase sets up the database connection
// if it doesn't already exist
func SetupDatabase() error {

	if db != nil {
		if db.Ping() == nil {
			return nil
		}
	}

	creds := dbCredentials{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
	}
	socketDir := fmt.Sprintf("/cloudsql/%s", creds.Host)
	dbConfig := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", socketDir, creds.User, creds.Password, creds.Database)

	// connect to database
	var err error
	db, err = sql.Open("postgres", dbConfig)
	if err != nil {
		return fmt.Errorf("could not connect to database: " + err.Error())
	}

	// ping database
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not ping database: " + err.Error())
	}

	fmt.Println("Database connected")

	return nil
}

// SetDataForUsername inserts the data into the database
func SetDataForUsername(username string, data string) (int, error) {

	if err := SetupDatabase(); err != nil {
		return 0, err
	}

	// prepare query
	query := "INSERT INTO users (username, data) VALUES ($1, $2)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// execute query
	res, err := stmt.Exec(username, data)
	if err != nil {
		return 0, err
	}

	// get affected rows
	affRows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	fmt.Printf("%d rows inserted\n", affRows)

	return int(affRows), nil
}

// GetDataForUsername selects the data associated
// with the given username
func GetDataForUsername(username string) (string, error) {

	if err := SetupDatabase(); err != nil {
		return "", err
	}

	// prepare query
	query := "SELECT data FROM users WHERE username = $1"
	stmt, err := db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	// execute query
	rows, err := stmt.Query(username)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("no data found for username: " + username)
	}

	// read data
	var data string
	err = rows.Scan(&data)
	if err != nil {
		return "", err
	}

	fmt.Println("Data read from DB: " + data)

	return data, nil
}
