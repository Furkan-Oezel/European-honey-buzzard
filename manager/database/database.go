package database

import (
	// import generic database interface package for SQL databases
	"database/sql"
	"log"
	/*
	 * Import this package only for its side effects.
	 * It's not used directly by me, I don't use something like sqlite.func(),
	 * instead its init() function is called by Go.
	 * This package registers itself as a driver with Go's database/sql package.
	 */
	_ "modernc.org/sqlite"
)

func Spawn_database() {
	// open database
	db, err := sql.Open("sqlite", "data/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	/*
	 * This table has 10 columns and grows automatically(new rows) when a new entry is saved.
	 * id: The column name.
         * INTEGER: The data type. Stores whole numbers.
         * PRIMARY KEY: Uniquely identifies each row in the table.
         * AUTOINCREMENT: Automatically increases the id with each new row.
	 * TIMESTAMP DEFAULT CURRENT_TIMESTAMP: If no value is provided, insert the current time by default.
	 */
	createTable := `CREATE TABLE IF NOT EXISTS containers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	container_id TEXT,
	container_name TEXT,
	image TEXT,
	action TEXT,
	event_type TEXT,
	cgroup_id TEXT,
	event_time INTEGER,
	event_time_nano INTEGER,
	start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}
