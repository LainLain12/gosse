package twoddata

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB creates the 'twoddata' table if it does not exist
func InitDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS twoddata (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        mset TEXT,
        mvalue TEXT,
        mresult TEXT,
        eset TEXT,
        evalue TEXT,
        eresult TEXT,
        tmodern TEXT,
        tinernet TEXT,
        nmodern TEXT,
        ninternet TEXT,
        updatetime TEXT,
        date DATE,
        status BOOLEAN
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}
	return db
}
