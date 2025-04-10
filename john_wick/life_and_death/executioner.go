package life_and_death

import (
	"log"
	"os/exec"
	"strings"
)

func Kill() {
	targetName := "lsm_chmod"

	out, err := exec.Command("pgrep", "-f", targetName).Output()
	if err != nil {
		log.Fatalf("❌ Could not find process matching '%s': %v", targetName, err)
	}

	pids := strings.Fields(string(out))
	if len(pids) == 0 {
		log.Println("ℹ️ No matching processes found.")
		return
	}

	for _, pid := range pids {
		log.Printf("🔪 Killing process with PID %s", pid)
		exec.Command("kill", "-INT", pid).Run() // Send SIGINT, like CTRL+C
	}
}
