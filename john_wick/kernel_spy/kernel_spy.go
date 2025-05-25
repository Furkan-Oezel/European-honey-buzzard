package kernel_spy

import (
	"context"
	"fmt"
	"syscall"

	"github.com/docker/docker/client"
)

func GetContainerCgroupID() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Errorf("creating Docker client: %w", err)
	}
	defer cli.Close()

	ctx := context.Background()
	inspect, err := cli.ContainerInspect(ctx, "1314fa4e37e9")
	if err != nil {
		fmt.Errorf("inspecting container: %w", err)
	}

	pid := inspect.State.Pid
	if pid == 0 {
		fmt.Errorf("container is not running")
	}

	// Point to /proc/<pid>/root/sys/fs/cgroup inside the container's namespace
	cgroupPath := fmt.Sprintf("/proc/%d/root/sys/fs/cgroup", pid)

	var stat syscall.Stat_t
	if err := syscall.Stat(cgroupPath, &stat); err != nil {
		fmt.Errorf("stat on cgroup path failed: %w", err)
	}
	fmt.Printf("Cgroup ID (inode): %d\n", stat.Ino)
}
