package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DbConnection struct {
	*sql.DB
}

type User struct {
	Username string
	Name     string
	Password string
	Type     string
	Subtitle string
	Avatar   string
	DbConnection
}

// database connection function
func Connect(database, user, password, address string) (db *sql.DB, err error) {
	dsn := user + ":" + password + "@tcp(" + address + ")/" + database
	db, err = sql.Open("mysql", dsn)

	return db, err
}

// GetPassword retrieves a password from a specific user
func (db *DbConnection) GetAttribute(username, attribute string) (string, error) {
	rows, err := db.Query("SELECT " + attribute + " from users WHERE username = '" + username + "'")
	if err != nil {
		//db.Close()
		return "", err
	}

	rows.Next()
	var attr string
	err = rows.Scan(&attr)
	if err != nil {
		//db.Close()
		return "", err
	}
	rows.Close()
	return attr, nil
}

// GetDoctors
func (db *DbConnection) GetDoctors() ([]User, error) {
	rows, err := db.Query("SELECT * from users WHERE type = 'doctor'")
	if err != nil {
		//db.Close()
		return []User{}, err
	}

	user := []User{}

	for rows.Next() {
		aux := User{}
		err = rows.Scan(&aux.Username, &aux.Name, &aux.Password, &aux.Type, &aux.Subtitle, &aux.Avatar)
		user = append(user, aux)
		if err != nil {
			//db.Close()
			return []User{}, err
		}
	}

	rows.Close()
	return user, nil
}
