package observer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/vishvananda/netlink"
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
			// only care about container events
			if event.Type != events.ContainerEventType {
				continue
			}

			containerID := event.Actor.ID
			containerName := event.Actor.Attributes["name"]
			image := event.Actor.Attributes["image"]
			action := event.Action
			eventType := event.Type
			eventTime := event.Time
			eventTimeNano := event.TimeNano

			// Calls into the github.com/vishvananda/netlink library to get a slice of all network links (interfaces) on the host.
			links, err := netlink.LinkList()
			if err != nil {
				log.Printf("Error listing links: %v", err)
			}

			/*
			 * We iterate over every link.
			 * l.Type() tells us the kernel’s type of the interface; we only care about those whose type is "veth".
			 * We pull out the interface’s name (l.Attrs().Name) and do an extra sanity check that it actually begins with the string "veth".
			 * Matching names get added to the veths slice.
			 */
			var veths []string
			for _, l := range links {
				if l.Type() == "veth" {
					name := l.Attrs().Name
					if strings.HasPrefix(name, "veth") {
						veths = append(veths, name)
					}
				}
			}
			// turn slice of names []string{"veth0","veth1234", …} into one string like "veth0,veth1234,..." that can be stored in a single database column
			vethField := strings.Join(veths, ",")

			// insert logging data into database
			insertStmt := `
				INSERT INTO container_logs (
					container_id,
					container_name,
					image,
					action,
					event_type,
					event_time,
					event_time_nano,
					veth
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

			_, err = db.Exec(insertStmt,
				containerID,
				containerName,
				image,
				action,
				eventType,
				eventTime,
				eventTimeNano,
				vethField,
			)
			if err != nil {
				log.Printf("Error inserting event: %v", err)
			} else {
				fmt.Printf(
					"New event: ID=%s, Action=%s, Name=%s, Image=%s, VETH=[%s]\n",
					containerID[:12], action, containerName, image, vethField,
				)
			}

		case err := <-errCh:
			if err != nil {
				log.Fatalf("Error from Docker event stream: %v", err)
			}
		}
	}
}
