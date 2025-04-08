package main

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

func main() {
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
		// Handle new Docker events
		case event := <-eventCh:
			// Check if it's a container start event
			if event.Type == events.ContainerEventType && event.Action == "start" {
				fmt.Printf("New container started: ID=%s\n", event.Actor.ID[:12])
			}
		// Handle errors
		case err := <-errCh:
			if err != nil {
				log.Fatalf("Error: %v", err) // Stop program on error
			}
		}
	}
}
