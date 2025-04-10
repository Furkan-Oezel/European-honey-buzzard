package updater

import (
	"fmt"
	"log"

	"github.com/cilium/ebpf"
)

const (
	map_key uint32 = 4
)

func Update() {
	// Path where the eBPF map is pinned
	mapPath := "/sys/fs/bpf/kernel_function/map_policy"

	// Open the pinned map
	pinnedMap, err := ebpf.LoadPinnedMap(mapPath, nil)
	if err != nil {
		log.Fatalf("Error opening pinned map: %v", err)
	}

	valueStr := "Hello eBPF!"
	var mapValue [64]byte
	copy(mapValue[:], valueStr)

	// Update the map
	if err := pinnedMap.Update(map_key, mapValue, ebpf.UpdateAny); err != nil {
		log.Fatalf("Error updating map: %v", err)
	}
	log.Printf("Successfully wrote value %s to map key ", mapValue)

	fmt.Println("Map updated successfully")
}
