package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("./lsm_chmod") // replace with full path if needed

	// Pass through stdin, stdout, stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the program
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run program: %v", err)
	}
}
