package observer

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	_ "modernc.org/sqlite" // SQLite driver
)

func Ob() {
	// open database
	db, err := sql.Open("sqlite", "data/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err) // Exit if client creation fails
	}
	cli.NegotiateAPIVersion(context.Background()) // Ensure API compatibility

	// Subscribe to Docker events with an empty ListOptions
	eventCh, errCh := cli.Events(context.Background(), events.ListOptions{})

	// Infinite loop to process events
	for {
		select {
		case event := <-eventCh:
			// Check if it's a container start event
			if event.Type == events.ContainerEventType && event.Action == "start" {
				containerID := event.Actor.ID[:12]
				fmt.Printf("New container started: ID=%s\n", containerID)

				// Insert container ID into database
				insertStmt := `INSERT OR IGNORE INTO containers (container_id) VALUES (?);`
				_, err = db.Exec(insertStmt, containerID)
				if err != nil {
					log.Printf("error inserting container with id: %v", err)
				} else {
					fmt.Println("saved container id")
				}
			}
		case err := <-errCh:
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	}
}
