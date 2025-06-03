package main

import (
	"john_wick/kernel_spy"
	"john_wick/spawner"
	"log"
)

func main() {
	paths := []string{
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_chmod",
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_rmdir",
		"/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_file_permission",
	}

	for _, path := range paths {
		go func(p string) {
			if err := spawner.Spawn(p); err != nil {
				log.Printf("error spawning %s: %v", p, err)
			}
		}(path)
	}

	kernel_spy.GetContainerCgroupIDs()

	// keep Goroutines alive by blocking main
	select {}
}
