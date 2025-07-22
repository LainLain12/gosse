package gift

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Gift represents a row in the gift table
type Gift struct {
	ID       string
	Name     string
	URL      string
	Category string // Optional, if you want to categorize gifts
}

// InitGiftDB creates the 'gift' table if it does not exist
func InitGiftDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS gift (
        id TEXT PRIMARY KEY,
        name TEXT,
        url TEXT,
		category TEXT
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("failed to create gift table: %v", err)
	}
	return db
}
