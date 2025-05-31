package main

import (
	"john_wick/spawner"
	"log"
	//"john_wick/kernel_spy"
	//"john_wick/updater"
)

func main() {
	err := spawner.Spawn("/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/lsm_modules/lsm_chmod")
	if err != nil {
		log.Fatal(err)
	}

	//updater.Update()

	//kernel_spy.GetContainerCgroupID()
}
