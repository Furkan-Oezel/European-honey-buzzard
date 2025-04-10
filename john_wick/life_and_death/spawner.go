package life_and_death

import (
	"log"
	"os"
	"os/exec"
)

func Spawn() {
	cmd := exec.Command("/home/furkan/oth/XI/European-honey-buzzard/john_wick/arsenal/maps/map_policy") // replace with full path if needed

	// Pass through stdin, stdout, stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the program
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run program: %v", err)
	}
}
