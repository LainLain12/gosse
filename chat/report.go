package chat

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Report represents a report entry
type Report struct {
	UserID      string `json:"userid"`
	ReportID    string `json:"reportid"`
	ReportCount int    `json:"reportcount"`
}

// InitReportTable creates the report table if it does not exist
func InitReportTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS report (
        userid TEXT,
        reportid TEXT,
        reportcount INTEGER PRIMARY KEY AUTOINCREMENT
    );`
	_, err := db.Exec(query)
	return err
}

// ReportHandler handles POST /report to add or update a report
func ReportHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			UserID   string `json:"userid"`
			ReportID string `json:"reportid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.UserID == "" || req.ReportID == "" {
			http.Error(w, "Missing userid or reportid", http.StatusBadRequest)
			return
		}

		// Check if this userid and reportid already exists
		var count int
		err := db.QueryRow("SELECT reportcount FROM report WHERE userid=? AND reportid=?", req.UserID, req.ReportID).Scan(&count)
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "already report",
				"userid":   req.UserID,
				"reportid": req.ReportID,
			})
			return
		} else if err != sql.ErrNoRows {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if reportid exists in any row
		err = db.QueryRow("SELECT reportcount FROM report WHERE reportid=? ORDER BY reportcount DESC LIMIT 1", req.ReportID).Scan(&count)
		if err == nil {
			// Exists, insert new row with incremented reportcount
			_, err = db.Exec("INSERT INTO report (userid, reportid, reportcount) VALUES (?, ?, ?)", req.UserID, req.ReportID, count+1)
			if err != nil {
				http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err == sql.ErrNoRows {
			// Not found, insert new row with reportcount = 1
			_, err = db.Exec("INSERT INTO report (userid, reportid, reportcount) VALUES (?, ?, 1)", req.UserID, req.ReportID)
			if err != nil {
				http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "reported",
		})
	}
}
