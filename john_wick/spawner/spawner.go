package spawner

import (
	"fmt"
	"os"
	"os/exec"
)

func Spawn(path string) error {
	cmd := exec.Command(path)

	// Pass through stdin, stdout, stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the program
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %s: %w", path, err)
	}
	return nil
}
