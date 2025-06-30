package main

import (
	"john_wick/kernel_spy"
	"john_wick/spawner"
	"log"
	"time"
)

func main() {
	paths := []string{
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_chmod",
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_rmdir",
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_file_permission",
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/firewall_modules/firewall_container",
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/firewall_modules/firewall_system",
	}

	for _, path := range paths {
		go func(p string) {
			if err := spawner.Spawn(p); err != nil {
				log.Printf("error spawning %s: %v", p, err)
			}
		}(path)
	}

	/*
	 * Temporary solution to fix the race condition between spawning LSM modules (which create the map)
	 * and calling kernel_spy.GetContainerCgroupIDs() (which tries to open it).
	 * Sleep is also necessary because 'set_ip_range' needs to be launched after 'firewall_system'
	 */
	time.Sleep(10 * time.Second)
	go spawner.Spawn(
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/userspace_programs/set_ip_range")
	kernel_spy.GetContainerCgroupIDs()

	// keep Goroutines alive by blocking main
	select {}
}
