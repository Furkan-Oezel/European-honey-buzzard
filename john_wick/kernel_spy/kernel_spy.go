package kernel_spy

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/cilium/ebpf"
	"github.com/docker/docker/client"
	_ "modernc.org/sqlite"
)

const mapPath = "/sys/fs/bpf/maps/map_policy"

func get_cgroupDIR_inode_number(containerIDs []string) map[string]uint64 {
	// establish docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Docker client error: %v", err)
	}
	defer cli.Close()

	// make a Go map to store inode numbers
	cgroup_directory_inodes := make(map[string]uint64)

	// get background context, necessary for functions like cli.ContainerInspect()
	ctx := context.Background()

	// create slice for container ids
	for _, containerID_slice := range containerIDs {

		// get status information of containers by inspecting them based on their container id
		inspect, err := cli.ContainerInspect(ctx, containerID_slice)
		if err != nil {
			log.Printf("Failed to inspect container %s: %v", containerID_slice[:12], err)
			continue
		}

		// print debug message if the container is not running (process id == 0)
		if inspect.State.Pid == 0 {
			log.Printf("Container %s not running", containerID_slice[:12])
			continue
		}

		// get cgroup path of the root directory of the container based on its process id
		cgroupPath := fmt.Sprintf("/proc/%d/root/sys/fs/cgroup", inspect.State.Pid)

		var stat syscall.Stat_t
		// call a syscall on the cgroup path
		// -> stores metadata like Ino (inode number) into the variable stat
		if err := syscall.Stat(cgroupPath, &stat); err != nil {
			log.Printf("⚠️ Stat failed for %s: %v", cgroupPath, err)
			continue
		}

		// store inode into map
		cgroup_directory_inodes[containerID_slice] = stat.Ino
	}

	/*
	 * return inode number of the cgroup directory seen from inside the container's root filesystem
	 * this is needed because bpf programs also return inodes when calling bpf_get_current_cgroup_id()
	 * unambiguous identifier
	 * as long as the container is not removed (docker rm), its cgroup id exist and so does this inode number
	 */
	return cgroup_directory_inodes
}

func readFilteredLogsDB() ([]string, error) {
	// open database
	db, err := sql.Open("sqlite", "../manager/data/filtered_logs.db")
	if err != nil {
		return nil, fmt.Errorf("Failed to open filtered_logs.db: %w", err)
	}
	defer db.Close()

	// query the column container_id (return rows with only container_id as field)
	rows, err := db.Query(`SELECT container_id FROM filtered_logs`)
	if err != nil {
		return nil, fmt.Errorf("Failed to query filtered_logs: %w", err)
	}
	defer rows.Close()

	// iterate the rows struct and save each container id into a slice
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	// return container ids
	return ids, nil
}

func GetContainerCgroupIDs() {
	pinnedMap, err := ebpf.LoadPinnedMap(mapPath, &ebpf.LoadPinOptions{})
	if err != nil {
		log.Fatalf("Failed to open pinned eBPF map: %v", err)
	}

	var current_bpf_map_entries = make(map[string]uint32)
	var index uint32 = 0

	// infinite loop
	for {
		// get docker ids from filtered_logs database
		containerIDs, err := readFilteredLogsDB()
		if err != nil {
			log.Printf("Could not read filtered_logs: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		/*
		 * the keys in presentIDs are container ids of the type string
		 * the values of this Go map are all empty structs
		 * -> it only matters IF this map has an entry at that specified key, not what its value is at that key
		 */
		presentIDs := make(map[string]struct{})
		for _, id := range containerIDs {
			presentIDs[id] = struct{}{}
		}

		// get cgroup ids of running containers
		cgroupMap := get_cgroupDIR_inode_number(containerIDs)

		// nth_containerID is the current container being processed in this iteration, it's a single value of type string
		// the underscore (_) discards the index
		for _, nth_containerID := range containerIDs {
			// check if a cgroup inode (cgroup ID) is available for this container id
			cgroupID, ok := cgroupMap[nth_containerID]
			// ok is a bool that returns true if the string nth_containerID led to a cgroup id
			if !ok {
				continue
				/*
				 * if no cgroup id was found:
				 * container id exists in the filtered_logs database
				 * yet there is no cgroup id
				 * -> container stopped running (exit) but was not removed (with docker rm)
				 * -> container id is still saved in filtered_logs database
				 * but since it is not running, the container has no process id and hence no cgroup id
				 */
			}

			/*
			 * current_bpf_map_entries looks like this:
			 * current_bpf_map_entries = {
			 *   "abc123": 0,
			 *   "def456": 1,
			 * }
			 * it's a map that holds keys
			 * container ids are the respective indices
			 * if a lookup in this Go map with the nth_containerID returns nothing (exists == false)
			 * then create a key
			 * increment index for the next loop iteration
			 */
			key, exists := current_bpf_map_entries[nth_containerID]
			if !exists {
				key = index
				index++
			}

			// update bpf map
			if err := pinnedMap.Update(key, cgroupID, ebpf.UpdateAny); err != nil {
				log.Printf("Failed to update eBPF map for container %s: %v", nth_containerID[:12], err)
				continue
			}

			// save the key if it was just created
			// this line does nothing if the key already existed
			current_bpf_map_entries[nth_containerID] = key

			log.Printf("Updated eBPF map: [%d] -> cgroup inode: %d | container id: %s", key, cgroupID, nth_containerID[:12])
		}

		// remove entries that are no longer in filtered_logs database
		for nth_containerID, key := range current_bpf_map_entries {
			// -> does the lookup return something? if yes, then stillPresent == true. if not, then stillPresent == false
			// also discard the value of the lookup (_), since it is only an empty struct
			if _, stillPresent := presentIDs[nth_containerID]; !stillPresent {
				if err := pinnedMap.Delete(key); err != nil {
					log.Printf("⚠️ Failed to delete key %d for container %s: %v", key, nth_containerID[:12], err)
				} else {
					log.Printf("🗑️ Removed key %d (container %s) from eBPF map", key, nth_containerID[:12])
				}
				// also delete the container id from the Go map
				delete(current_bpf_map_entries, nth_containerID)
			}
		}

		// pause for 3 seconds before restarting the loop
		// prevents constant polling and gives Docker time to change state
		time.Sleep(3 * time.Second)
	}
}
