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

// Gift struct with category support

// GiftDataHandler handles GET /giftdata and returns rows as JSON, filtered by id and/or category if provided
func GiftDataHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		category := r.URL.Query().Get("category")
		var rows *sql.Rows
		var err error
		if id != "" && category != "" {
			rows, err = db.Query(`SELECT id, category, name, url FROM gift WHERE id=? AND category=?`, id, category)
		} else if id != "" {
			rows, err = db.Query(`SELECT id, category, name, url FROM gift WHERE id=?`, id)
		} else if category != "" {
			rows, err = db.Query(`SELECT id, category, name, url FROM gift WHERE category=?`, category)
		} else {
			rows, err = db.Query(`SELECT id, category, name, url FROM gift`)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var all []Gift
		for rows.Next() {
			var g Gift
			err := rows.Scan(&g.ID, &g.Category, &g.Name, &g.URL)
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
func AddGiftHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "Internal server error (panic)", http.StatusInternalServerError)
				fmt.Fprintf(os.Stderr, "PANIC in AddGiftHandler: %v\n", rec)
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
		fpath := "gift/images/" + fname
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

		// Insert or update gift table with id and category
		id := r.URL.Query().Get("id")
		if id == "" {
			id = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		category := r.URL.Query().Get("category")
		if category == "" {
			category = "default"
		}
		name := fname
		relUrl := "/gift/images/" + fname
		scheme := "http"
		host := r.Host
		fullUrl := scheme + "://" + host + relUrl

		// If id+category exists, get old url and delete old file after update
		var oldUrl string
		var oldFilename string
		err = db.QueryRow("SELECT url FROM gift WHERE id=? AND category=?", id, category).Scan(&oldUrl)
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

		res, err := db.Exec("UPDATE gift SET url=?, name=? WHERE id=? AND category=?", fullUrl, name, id, category)
		rowsAffected, _ := res.RowsAffected()
		if err != nil || rowsAffected == 0 {
			// Insert if update did not affect any row
			_, err := db.Exec("INSERT OR REPLACE INTO gift (id, category, name, url) VALUES (?, ?, ?, ?)", id, category, name, fullUrl)
			if err != nil {
				http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Delete old file if needed
		if oldFilename != "" && oldFilename != fname {
			oldPath := "gift/images/" + oldFilename
			_ = os.Remove(oldPath)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "success",
			"imagename": fname,
			"id":        id,
			"category":  category,
			"url":       fullUrl,
		})
	}
}

// ensureImagesDir creates the images directory if it doesn't exist
func ensureImagesDir() error {
	return os.MkdirAll("gift/images", 0755)
}
