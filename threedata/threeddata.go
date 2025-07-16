package threedata

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// ThreedData represents a row in the threeddata table
type ThreedData struct {
	Date   string
	Result string
}

// InitThreedDB creates the 'threeddata' table if it does not exist
func InitThreedDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS threeddata (
        date TEXT,
        result TEXT
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("failed to create threeddata table: %v", err)
	}
	return db
}
