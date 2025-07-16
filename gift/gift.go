package gift

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GiftDataHandler handles GET /giftdata and returns all rows as JSON
func GiftDataHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, name, url FROM gift`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var all []Gift
		for rows.Next() {
			var g Gift
			err := rows.Scan(&g.ID, &g.Name, &g.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			all = append(all, g)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
	}
}

// AddImageHandler handles POST /addimage to upload an image to the images folder
func AddImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "Internal server error (panic)", http.StatusInternalServerError)
				fmt.Fprintf(os.Stderr, "PANIC in AddImageHandler: %v\n", rec)
			}
		}()
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Create images folder if not exists
		if err := ensureImagesDir(); err != nil {
			http.Error(w, "Failed to create images dir: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Read all image data from body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read image: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Use extension from Content-Type if possible, else default to .png
		ext := ".png"
		switch r.Header.Get("Content-Type") {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		}

		fname := fmt.Sprintf("%f%s", float64(time.Now().UnixNano())/1e9, ext)
		fpath := "images/" + fname
		out, err := os.Create(fpath)
		if err != nil {
			http.Error(w, "Failed to save image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = out.Write(body)
		if err != nil {
			http.Error(w, "Failed to write image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert or update gift table
		id := r.URL.Query().Get("id")
		if id == "" {
			id = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		name := fname
		relUrl := "/images/" + fname
		scheme := "http"
		host := r.Host
		fullUrl := scheme + "://" + host + relUrl

		// If id exists, get old url and delete old file after update
		var oldUrl string
		var oldFilename string
		err = db.QueryRow("SELECT url FROM gift WHERE id=?", id).Scan(&oldUrl)
		if err == nil && oldUrl != "" {
			// Always extract filename after last slash
			lastSlash := -1
			for i := len(oldUrl) - 1; i >= 0; i-- {
				if oldUrl[i] == '/' {
					lastSlash = i
					break
				}
			}
			if lastSlash != -1 && lastSlash+1 < len(oldUrl) {
				oldFilename = oldUrl[lastSlash+1:]
			}
		}

		res, err := db.Exec("UPDATE gift SET url=? WHERE id=?", fullUrl, id)
		rowsAffected, _ := res.RowsAffected()
		if err != nil || rowsAffected == 0 {
			// Insert if update did not affect any row
			_, err := db.Exec("INSERT OR REPLACE INTO gift (id, name, url) VALUES (?, ?, ?)", id, name, fullUrl)
			if err != nil {
				http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Delete old file if needed
		if oldFilename != "" && oldFilename != fname {
			oldPath := "images/" + oldFilename
			_ = os.Remove(oldPath)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "success",
			"imagename": fname,
			"id":        id,
			"url":       fullUrl,
		})
		// To serve images, add this in main.go:
		// http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	}
}

// ensureImagesDir creates the images directory if it doesn't exist
func ensureImagesDir() error {
	return os.MkdirAll("images", 0755)
}
