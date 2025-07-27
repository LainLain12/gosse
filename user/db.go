package user

import (
	"database/sql"
)

// CreateUserAccountTable creates the useraccount table if it does not exist
func CreateUserAccountTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS useraccount (
        id TEXT PRIMARY KEY,
        name TEXT,
        profile_pic TEXT,
        email TEXT
    );`
	_, err := db.Exec(query)
	return err
}
