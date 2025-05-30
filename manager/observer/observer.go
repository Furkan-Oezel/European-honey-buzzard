package observer

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	_ "modernc.org/sqlite"
)

func Observe() {
	// open database
	db, err := sql.Open("sqlite", "data/container_logs.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	cli.NegotiateAPIVersion(context.Background())

	// initialize Go channel that subscribes to Docker event stream
	eventCh, errCh := cli.Events(context.Background(), events.ListOptions{})

	// infinite loop
	for {
		select {
		// when a new Docker event arrives (case <-eventCh), Go assigns it to the event variable and executes this block of code
		case event := <-eventCh:
			if event.Type == events.ContainerEventType {
				containerID := event.Actor.ID
				containerName := event.Actor.Attributes["name"]
				image := event.Actor.Attributes["image"]
				action := event.Action
				eventType := event.Type
				eventTime := event.Time
				eventTimeNano := event.TimeNano

				// insert logging data into database
				insertStmt := `
					INSERT INTO container_logs (
						container_id, container_name, image, action, event_type, event_time, event_time_nano
					) VALUES (?, ?, ?, ?, ?, ?, ?);`

				_, err := db.Exec(insertStmt,
					containerID, containerName, image, action, eventType, eventTime, eventTimeNano)
				if err != nil {
					log.Printf("Error inserting event: %v", err)
				} else {
					fmt.Printf("New event: ID=%s, Action=%s, Name=%s, Image=%s\n",
						containerID[:12], action, containerName, image)
				}
			}

		case err := <-errCh:
			if err != nil {
				log.Fatalf("Error from Docker event stream: %v", err)
			}
		}
	}
}
