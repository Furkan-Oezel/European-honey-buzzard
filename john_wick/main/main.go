package main

import (
	"fmt"
	//"john_wick/life_and_death"
	"john_wick/kernel_spy"
	//"john_wick/updater"
)

func main() {
	fmt.Println("hi from main")
	//life_and_death.Spawn()
	//life_and_death.Kill()
	//updater.Update()
	kernel_spy.GetContainerCgroupID()
}
