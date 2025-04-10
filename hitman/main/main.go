package main

import (
	"fmt"
	"hitman/life_and_death"
	"hitman/updater"
)

func main() {
	fmt.Println("hi from main")
	life_and_death.Spawn()
	updater.Update()
}
