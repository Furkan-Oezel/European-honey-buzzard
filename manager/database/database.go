package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func Spawn_database() {
	// Datenbankverbindung öffnen oder erstellen
	db, err := sql.Open("sqlite", "data/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Tabelle erstellen (falls nicht vorhanden)
	createTable := `CREATE TABLE IF NOT EXISTS containers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		container_id TEXT UNIQUE,
		start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ SQLite-Datenbank und Tabelle initialisiert")

	// Beispiel-Daten einfügen
	containerID := "abc123def456"
	insertStmt := `INSERT OR IGNORE INTO containers (container_id) VALUES (?);`
	_, err = db.Exec(insertStmt, containerID)
	if err != nil {
		log.Fatalf("❌ Fehler beim Einfügen: %v", err)
	}
	fmt.Println("📦 Beispiel-Container-ID eingefügt:", containerID)

	// Daten auslesen
	rows, err := db.Query(`SELECT id, container_id, start_time FROM containers;`)
	if err != nil {
		log.Fatalf("❌ Fehler beim Auslesen der Datenbank: %v", err)
	}
	defer rows.Close()

	fmt.Println("📋 Aktueller Inhalt der Tabelle:")
	for rows.Next() {
		var id int
		var cid string
		var startTime string
		err := rows.Scan(&id, &cid, &startTime)
		if err != nil {
			log.Fatalf("❌ Fehler beim Auslesen einer Zeile: %v", err)
		}
		fmt.Printf("  🆔 %d | 🐳 %s | 🕒 %s\n", id, cid, startTime)
	}
}
