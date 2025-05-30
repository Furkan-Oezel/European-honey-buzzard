package filtered_logs

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func Spawn_filtered_logs() {
	// open database
	db, err := sql.Open("sqlite", "data/filtered_logs.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS filtered_logs (
		container_id TEXT PRIMARY KEY,
		action TEXT
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successfully initialized filtered_logs database")
	}
}

func Filter() {
	// open container_logs database
	logDB, err := sql.Open("sqlite", "data/container_logs.db")
	if err != nil {
		log.Fatal(err)
	}
	defer logDB.Close()

	// open filtered_logs database
	filteredDB, err := sql.Open("sqlite", "data/filtered_logs.db")
	if err != nil {
		log.Fatal(err)
	}
	defer filteredDB.Close()

	// this query selects the last entry of a given container id
	query := `
		SELECT container_id, action FROM container_logs
		WHERE event_time_nano IN (
			SELECT MAX(event_time_nano)
			FROM container_logs
			GROUP BY container_id
		);
	`

	rows, err := logDB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var containerID, action string
		if err := rows.Scan(&containerID, &action); err != nil {
			log.Printf("Error reading the row: %v", err)
			continue
		}

		if action == "destroy" {
			// delete the entry if the last action is destroy
			deleteStmt := `DELETE FROM filtered_logs WHERE container_id = ?`
			_, err := filteredDB.Exec(deleteStmt, containerID)
			if err != nil {
				log.Printf("Error deleting container %s: %v", containerID, err)
			} else {
				fmt.Printf("Deleted container %s (action=destroy)\n", containerID[:12])
			}

		} else {
			// try to insert the data as a new entry
			// if the entry with that container id already exists, update it instead
			insertOrUpdate := `
				INSERT INTO filtered_logs (container_id, action)
				VALUES (?, ?)
				ON CONFLICT(container_id) DO UPDATE SET action=excluded.action;
			`

			_, err := filteredDB.Exec(insertOrUpdate, containerID, action)
			if err != nil {
				log.Printf("Error inserting/updating for %s: %v", containerID, err)
			} else {
				fmt.Printf("Successfully updated status: %s → %s\n", containerID[:12], action)
			}
		}
	}
}
