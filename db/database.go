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
	Patients int
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
		err = rows.Scan(&aux.Username, &aux.Name, &aux.Password, &aux.Type, &aux.Subtitle, &aux.Avatar, &aux.Patients)
		user = append(user, aux)
		if err != nil {
			//db.Close()
			return []User{}, err
		}
	}

	rows.Close()
	return user, nil
}

// GetPatients from a doctor (need to improve table schema)
func (db *DbConnection) GetPatients(doctor string) ([]User, error) {

	//rows, err := db.Query("SELECT * FROM users WHERE username IN (SELECT p.username FROM patients AS p JOIN users AS u ON p.id = u.patients WHERE u.name = '" + doctor + "')")
	// temporarily workaround until find a way to automatically update the list of patients
	rows, err := db.Query("SELECT * FROM users WHERE type = 'patient'")
	if err != nil {
		return []User{}, err
	}

	user := []User{}

	for rows.Next() {
		aux := User{}
		err = rows.Scan(&aux.Username, &aux.Name, &aux.Password, &aux.Type, &aux.Subtitle, &aux.Avatar, &aux.Patients)
		user = append(user, aux)
		if err != nil {
			//db.Close()
			return []User{}, err
		}
	}

	rows.Close()
	return user, nil
}

func (db *DbConnection) CreateUser(username, name, password, userType, subtitle, avatar string) error {
	_, err := db.Exec(`INSERT INTO users VALUES(
		'` + username + `',
		'` + name + `',
		'` + password + `',
		'` + userType + `',
		'` + subtitle + `',
		'` + avatar + `',
		'0')`)
	if err != nil {
		return err
	}
	return nil
}
