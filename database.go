package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq"
)

var db *sql.DB

type dbCredentials struct {
	Host string `json:"host"`
	// Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func SetupDatabase() error {

	if db != nil {
		if db.Ping() == nil {
			return nil
		}
	}

	// read database credentials from db.env file
	dat, err := ioutil.ReadFile("db.env")
	if err != nil {
		return fmt.Errorf("could not find db.env file: " + err.Error())
	}

	// parse database credentials
	var creds dbCredentials
	err = json.Unmarshal(dat, &creds)
	if err != nil {
		return fmt.Errorf("could not parse db.env file: " + err.Error())
	}

	socketDir := fmt.Sprintf("/cloudsql/%s/.s.PGSQL.5432.", creds.Host)

	dbConfig := fmt.Sprintf("host=%s/%s user=%s password=%s dbname=%s sslmode=disable", socketDir, creds.Host, creds.User, creds.Password, creds.Database)

	// connect to database
	db, err = sql.Open("pgx", dbConfig)
	if err != nil {
		return fmt.Errorf("could not connect to database: " + err.Error())
	}

	// ping database
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not ping database: " + err.Error())
	}

	return nil
}

func SetDataForUsername(username string, data string) (int, error) {
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

	return int(affRows), nil
}

func GetDataForUsername(username string) (string, error) {
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

	// read data
	var data string
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	return data, nil
}
