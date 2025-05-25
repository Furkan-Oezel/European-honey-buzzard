package kernel_spy

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"syscall"

	"github.com/docker/docker/client"
	_ "modernc.org/sqlite"
)

func GetContainerCgroupID() {
	// Open the SQLite database
	db, err := sql.Open("sqlite", "../manager/data/database.db")
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}
	defer db.Close()

	// Fetch the second container ID (change LIMIT and OFFSET to get different entries)
	var containerID string
	query := `SELECT container_id FROM containers ORDER BY id ASC LIMIT 1 OFFSET 0;`

	err = db.QueryRow(query).Scan(&containerID)
	if err != nil {
		log.Fatalf("❌ Failed to fetch container ID: %v", err)
	}
	fmt.Printf("📦 Using container ID from database: %s\n", containerID)

	// Inspect the container using the Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("❌ Creating Docker client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()
	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Fatalf("❌ Inspecting container: %v", err)
	}

	pid := inspect.State.Pid
	if pid == 0 {
		log.Fatalf("❌ Container is not running (PID 0)")
	}

	// Access the container’s cgroup path
	cgroupPath := fmt.Sprintf("/proc/%d/root/sys/fs/cgroup", pid)

	var stat syscall.Stat_t
	if err := syscall.Stat(cgroupPath, &stat); err != nil {
		log.Fatalf("❌ Stat on cgroup path failed: %v", err)
	}

	fmt.Printf("🧠 Cgroup ID (inode): %d\n", stat.Ino)
}
