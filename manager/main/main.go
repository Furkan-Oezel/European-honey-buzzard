package main

import (
	"fmt"
	"manager/database"
	"manager/filter"
	"manager/observer"
)

func main() {
	fmt.Println("hi from main")
	filter.Spawn_filter()
	database.Spawn_database()
	observer.Ob()
}
