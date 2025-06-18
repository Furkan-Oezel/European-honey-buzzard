package container_logs

import (
	// import generic database interface package for SQL databases
	"database/sql"
	"fmt"
	"log"
	/*
	 * Import this package only for its side effects.
	 * It's not used directly by me, I don't use something like sqlite.func(),
	 * instead its init() function is called by Go.
	 * This package registers itself as a driver with Go's database/sql package.
	 */
	_ "modernc.org/sqlite"
)

func Spawn_container_logs() {
	// open database
	db, err := sql.Open("sqlite", "data/container_logs.db")
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
	 * veth: (virtual Ethernet) is a virtual network interface pair used to connect a Docker container to the host network or to another container.
	 * TIMESTAMP DEFAULT CURRENT_TIMESTAMP: If no value is provided by observer.go, insert the current time by default.
	 */
	createTable := `CREATE TABLE IF NOT EXISTS container_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	container_id TEXT,
	container_name TEXT,
	image TEXT,
	action TEXT,
	event_type TEXT,
	event_time INTEGER,
	event_time_nano INTEGER,
	veth TEXT,
	start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successfully initialized container_logs database")
	}
}
