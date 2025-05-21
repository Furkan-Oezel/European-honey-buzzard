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

func Ob() {
	db, err := sql.Open("sqlite", "data/database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	cli.NegotiateAPIVersion(context.Background())

	eventCh, errCh := cli.Events(context.Background(), events.ListOptions{})

	for {
		select {
		case event := <-eventCh:
			if event.Type == events.ContainerEventType {
				containerID := event.Actor.ID
				containerName := event.Actor.Attributes["name"]
				image := event.Actor.Attributes["image"]
				action := event.Action
				eventType := event.Type
				eventTime := event.Time
				eventTimeNano := event.TimeNano

				fmt.Printf("New event: ID=%s, Action=%s, Name=%s, Image=%s\n",
					containerID[:12], action, containerName, image)

				insertStmt := `
					INSERT INTO containers (
						container_id, container_name, image, action, event_type, event_time, event_time_nano
					) VALUES (?, ?, ?, ?, ?, ?, ?);`
				_, err := db.Exec(insertStmt,
					containerID, containerName, image, action, eventType, eventTime, eventTimeNano)
				if err != nil {
					log.Printf("❌ error inserting event: %v", err)
				} else {
					fmt.Println("✅ Docker event saved")
				}
			}

		case err := <-errCh:
			if err != nil {
				log.Fatalf("❌ Error from Docker event stream: %v", err)
			}
		}
	}
}
