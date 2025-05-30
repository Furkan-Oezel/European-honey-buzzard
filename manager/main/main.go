package main

import (
	"manager/container_logs"
	"manager/filtered_logs"
	"manager/observer"
	"time"
)

func main() {
	container_logs.Spawn_container_logs()
	filtered_logs.Spawn_filtered_logs()

	go observer.Observe()

	go func() {
		for {
			filtered_logs.Filter()
			time.Sleep(5 * time.Second)
		}
	}()

	// prevent main() from an immediate stop
	select {}
}
