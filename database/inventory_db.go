package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB *sql.DB

func InitDatabase() {
	var err error
	DB, err = sql.Open("sqlite3", "./game_data.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	createInventoryTable := `
	CREATE TABLE IF NOT EXISTS inventory (
		slot INTEGER PRIMARY KEY,
		type INTEGER,
		name TEXT
	);`

	_, err = DB.Exec(createInventoryTable)
	if err != nil {
		log.Fatal("Failed to create inventory table:", err)
	}
}
