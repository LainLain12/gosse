package lottosociety

import (
	"database/sql"
)

// LottoSociety represents a lottery entry
type LottoSociety struct {
	Date     string `json:"date"`
	ThaiDate string `json:"thaidate"`
	FNum     string `json:"fnum"`
	SNum     string `json:"snum"`
	ID       string `json:"id"`
	Text     string `json:"text"`
}

// InitLottoSocietyTable creates the lottosociety table if it does not exist
func InitLottoSocietyTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS lottosociety (
        date TEXT,
        thaidate TEXT,
        fnum TEXT,
        snum TEXT,
        id TEXT,
        text TEXT
    );`
	_, err := db.Exec(query)
	return err
}
