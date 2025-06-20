package filtered_logs

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/vishvananda/netlink"
	_ "modernc.org/sqlite"
)

func Spawn_filtered_logs() error {
	// open database
	db, err := sql.Open("sqlite", "data/filtered_logs.db")
	if err != nil {
		return fmt.Errorf("opening filtered_logs.db: %w", err)
	}
	defer db.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS filtered_logs (
		container_id TEXT PRIMARY KEY,
		action       TEXT,
		veth         TEXT
	);`
	if _, err := db.Exec(createTable); err != nil {
		return fmt.Errorf("creating filtered_logs table: %w", err)
	}

	/*
	 * Write-Ahead Logging (WAL)
	 *
	 * why do I need WAL? -> because otherwise "database is locked" error occurs
	 *
	 * DELETE-journal mode is the default updating mode. For each database
	 * page that’s about to change (e.g. the disk block holding your row),
	 * SQLite first copies the original page into the journal file. This
	 * happens before the write to the main database. SQLite writes the new
	 * data directly into the database(overwrites the database, not the
	 * journal). When the transaction succeeds, SQLite can safely delete
	 * the entire journal file. If something goes wrong, then SQLite reads
	 * each saved page in the journal and restores it into the database,
	 * undoing every write that has been done during this transaction.
	 *
	 * A transaction in a database is a sequence of one or more operations
	 * (typically SQL statements) that are treated as a single, indivisible
	 * “unit of work.” The key idea is that either all of the operations in
	 * the transaction succeed (and are made permanent), or none of them do
	 * (and the database is rolled back (e.g with the journal) to its prior
	 * state). Imagine you’re moving money from account A to account B: you
	 * need to debit A and credit B. If the first succeeds and the second
	 * fails, your data is inconsistent. Wrapping both in a single
	 * transaction ensures you never see a half-completed transfer.
	 *
	 * What exactly is WAL? -> A journaling mode focusing concurrency.
	 *
	 * When you modify a row, SQLite figures out which database page that
	 * row lives on (typically a 4 KiB block). It takes the entire contents
	 * of that page after modification (database stays untouched until a
	 * checkpoint) and appends it as a frame to the end of the WAL file.
	 * Along with each frame, it records the page number (i.e. “this is
	 * page 27 of the DB”) and a commit sequence number in the WAL header.
	 * In addition to the WAL file itself (.db-wal), SQLite keeps a tiny
	 * shared-memory index (.db-shm) that, for each page number, tells you
	 * the offset in the WAL of the most recent committed frame for that
	 * page. On startup of any connection (reader or writer), SQLite reads
	 * that index into a small in-memory table. Reading a page merges main
	 * DB + WAL. When you do a SELECT or otherwise read a page: Lookup the
	 * page number in the in-memory WAL index. If there is an entry, fetch
	 * that frame from the WAL file—it’s the newest committed copy of that
	 * page. If there is no entry, fall back to reading the page from the
	 * main database file itself. Readers ignore any frames tagged with a
	 * newer commit number than what they saw at the start of their own
	 * transaction.     Periodically (or on PRAGMA wal_checkpoint), SQLite
	 * copies the latest frames for each page from the WAL back into the
	 * main database file, then truncates the WAL. After checkpointing, the
	 * main file once again contains the freshest data, and the WAL file
	 * goes back to being empty.
	 */
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		log.Printf("Warning: could not enable WAL mode: %v", err)
	}

	// Instruct SQLite: if a database lock is encountered, keep retrying
	// for up to 5000 ms before returning SQLITE_BUSY.
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		log.Printf("Warning: could not set busy_timeout: %v", err)
	}

	fmt.Println("Successfully initialized filtered_logs database")
	return nil
}

// filter for docker id, event and veth
func Filter() {
	// Calls into the github.com/vishvananda/netlink library to get a slice of all network links (interfaces) on the host.
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Failed to list links: %v", err)
	}
	// allocate a map where the keys are the interface names and the values are always true
	existing := make(map[string]bool, len(links))
	for _, l := range links {
		name := l.Attrs().Name
		if l.Type() == "veth" && strings.HasPrefix(name, "veth") {
			existing[name] = true
		}
	}

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

	/*
	 * Attempt to create a new row with the three columns (container_id, action, veth).
	 * On conflict (there’s already a row with the same container_id (the primary key)):
	 * Update the existing row’s action and veth to the new values.
	 */
	upsert := `
		INSERT INTO filtered_logs (container_id, action, veth)
		VALUES (?, ?, ?)
		ON CONFLICT(container_id) DO UPDATE
		  SET action = excluded.action,
		      veth   = excluded.veth;
	`

	// grab rows whose event_time_nano equals the maximum for that container_id
	rows, err := logDB.Query(`
		SELECT container_id, action, veth FROM container_logs
		WHERE event_time_nano IN (
			SELECT MAX(event_time_nano)
			FROM container_logs
			GROUP BY container_id
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var cid, action, vethCSV string
		if err := rows.Scan(&cid, &action, &vethCSV); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		if action == "destroy" {
			// delete the entry if the last action is destroy
			if _, err := filteredDB.Exec(
				`DELETE FROM filtered_logs WHERE container_id = ?`, cid,
			); err != nil {
				log.Printf("Error deleting %s: %v", cid, err)
			} else {
				fmt.Printf("Deleted container %s (action=destroy)\n", cid[:12])
			}
		} else {
			// try to insert the data as a new entry (use the SQL statement 'upsert' from before)
			if _, err := filteredDB.Exec(upsert, cid, action, vethCSV); err != nil {
				log.Printf("Error upserting %s: %v", cid, err)
			} else {
				fmt.Printf("Upserted %s → action=%s, veth=[%s]\n",
					cid[:12], action, vethCSV)
			}
		}
	}
	rows.Close()

	type cleanTask struct {
		cid    string // container_id
		oldCSV string // original comma-seperated veth list
		newCSV string // pruned list (only interfaces that still exist)
	}
	var tasks []cleanTask

	// get rows with container_id and veth
	cleanupRows, err := filteredDB.Query(`SELECT container_id, veth FROM filtered_logs`)
	if err != nil {
		log.Fatalf("Error fetching for cleanup: %v", err)
	}

	// if Next() returns true -> iterate one more time
	for cleanupRows.Next() {
		var cid, vethCSV string
		if err := cleanupRows.Scan(&cid, &vethCSV); err != nil {
			log.Printf("Cleanup scan error: %v", err)
			continue
		}
		// Split the raw CSV string (e.g. "veth0,veth1") into a slice:
		// parts == []string{"veth0", "veth1"}
		parts := strings.Split(vethCSV, ",")
		var kept []string
		for _, v := range parts {
			if existing[v] {
				kept = append(kept, v)
			}
		}
		// Re-join the filtered list back into a single string.
		newCSV := strings.Join(kept, ",")
		// check whether any veths got removed
		if newCSV != vethCSV {
			// buffer for later when updating the database
			tasks = append(tasks, cleanTask{cid, vethCSV, newCSV})
		}
	}
	// Close the rows to release the read lock and free resources.
	// After this point it’s safe to run UPDATEs against filtered_logs.
	cleanupRows.Close()

	for _, t := range tasks {
		if _, err := filteredDB.Exec(
			/*
			 * Run an SQL UPDATE on the filtered_logs table,
			 * setting its veth column to the new, pruned CSV
			 * (t.newCSV) for the matching container_id (t.cid).
			 */
			`UPDATE filtered_logs SET veth = ? WHERE container_id = ?`,
			t.newCSV, t.cid,
		); err != nil {
			log.Printf("Error cleaning veth for %s (was [%s]): %v",
				t.cid[:12], t.oldCSV, err)
		} else {
			fmt.Printf("Cleaned %s: old=[%s], new=[%s]\n",
				t.cid[:12], t.oldCSV, t.newCSV)
		}
	}
}
