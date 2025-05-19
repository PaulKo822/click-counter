// db/db.go
package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "user=postgres dbname=click_counter sslmode=disable password=1234 host=localhost port=5434"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database")

	createTables()
}

func createTables() {
	query := `
    CREATE TABLE IF NOT EXISTS banners (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS clicks (
        timestamp TIMESTAMP NOT NULL,
        banner_id INTEGER NOT NULL REFERENCES banners(id),
        count INTEGER NOT NULL DEFAULT 1,
        PRIMARY KEY (timestamp, banner_id)
    );

    CREATE INDEX IF NOT EXISTS idx_banner_id ON clicks(banner_id);
    CREATE INDEX IF NOT EXISTS idx_timestamp ON clicks(timestamp);
    `

	_, err := DB.Exec(query)
	if err != nil {
		log.Printf("Error creating tables: %v", err)
		log.Fatal(err)
	}
}
