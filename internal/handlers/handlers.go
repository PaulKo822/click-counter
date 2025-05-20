// handlers/handlers.go
package handlers

import (
	"click-counter/internal/db"
	"click-counter/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func IncrementCounter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bannerIDStr := vars["bannerID"]

	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Invalid banner ID", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	minuteStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var count int
	err = tx.QueryRow(`
        UPDATE clicks SET count = count + 1 
        WHERE banner_id = $1 AND timestamp = $2 
        RETURNING count`,
		bannerID, minuteStart).Scan(&count)

	if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO clicks (timestamp, banner_id) VALUES ($1, $2)", minuteStart, bannerID)
		if err != nil {
			tx.Rollback()
			log.Printf("Insert error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		tx.Rollback()
		log.Printf("Update error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Commit error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

type StatsRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bannerIDStr := vars["bannerID"]

	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Invalid banner ID", http.StatusBadRequest)
		return
	}

	var req StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	from, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		http.Error(w, "Invalid from timestamp", http.StatusBadRequest)
		return
	}

	to, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		http.Error(w, "Invalid to timestamp", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query(`
        SELECT timestamp, count FROM clicks
        WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
        ORDER BY timestamp ASC`, bannerID, from, to)

	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	stats := []models.ClickStat{}

	for rows.Next() {
		var ts time.Time
		var count int
		if err := rows.Scan(&ts, &count); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		stats = append(stats, models.ClickStat{
			Timestamp: ts,
			Count:     count,
		})
	}

	response := struct {
		Stats []models.ClickStat `json:"stats"`
	}{
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
