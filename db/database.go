package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DbConnection struct {
	*sql.DB
}

// database connection function
func Connect(database, user, password, address string) (db *sql.DB, err error) {
	dsn := user + ":" + password + "@tcp(" + address + ")/" + database
	db, err = sql.Open("mysql", dsn)

	return db, err
}

// GetPassword retrieves a password from a specific user
func (db *DbConnection) GetPassword(username string) (string, error) {
	rows, err := db.Query("SELECT password from users WHERE name = '" + username + "'")
	if err != nil {
		db.Close()
		return "", err
	}

	rows.Next()
	var password string
	err = rows.Scan(&password)
	if err != nil {
		db.Close()
		return "", err
	}
	rows.Close()
	return password, nil
}
