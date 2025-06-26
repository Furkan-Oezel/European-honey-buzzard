package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	_ "modernc.org/sqlite"
)

/*
 * Attache the firewall to with what's currently recorded in the filtered_logs database.
 * - Attaches to any new veths
 * - Detaches from any veths that have been removed
 */
func updateAttachments(db *sql.DB, prog *ebpf.Program, attached map[string]link.Link) {
	// fetch the comma-seperated 'veth' column
	rows, err := db.Query(`SELECT veth FROM filtered_logs WHERE action != 'destroy'`)
	if err != nil {
		log.Printf("Error querying filtered_logs: %v", err)
		return
	}
	defer rows.Close()

	// map with desired veth interfaces as keys and empty structs (0 value) as values
	desired := make(map[string]struct{})
	// iterate through the veth column
	for rows.Next() {
		var csv string
		if err := rows.Scan(&csv); err != nil {
			log.Printf("Warning: scan error: %v", err)
			continue
		}
		// split comma-separated veth names ("veth0,veth1") into individual names
		for _, name := range strings.Split(csv, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				desired[name] = struct{}{}
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Warning: rows iteration error: %v", err)
	}

	//iterate through the desired veth names map
	for name := range desired {
		// check if that veth name has already been attached to
		if _, ok := attached[name]; ok {
			continue
		}
		// get numeric interface based on veth interface
		// To attach a TC hook, you must tell the kernel which interface by its numeric index, not its string name.
		iface, err := net.InterfaceByName(name)
		if err != nil {
			log.Printf("Could not find interface %q: %v", name, err)
			continue
		}
		// attach the firewall at the TC ingress hook on this index.
		lnk, err := link.AttachTCX(link.TCXOptions{
			Program:   prog,
			Interface: iface.Index,
			Attach:    ebpf.AttachTCXIngress,
		})
		if err != nil {
			log.Printf("Failed to attach to %q: %v", name, err)
			continue
		}
		// record the attachment handle so it can be detached later if necessary
		attached[name] = lnk
		fmt.Printf(">> attached eBPF TC to %q ingress\n", name)
	}

	// detach from interfaces no longer desired
	for name, lnk := range attached {
		if _, ok := desired[name]; !ok {
			lnk.Close()
			delete(attached, name)
			fmt.Printf("<< detached eBPF TC from %q\n", name)
		}
	}
}

func main() {
	// Allow locking memory for eBPF
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("removing memlock rlimit: %v", err)
	}

	// Load the compiled eBPF ELF and load it into the kernel.
	var objs firewall_containerObjects
	if err := loadFirewall_containerObjects(&objs, nil); err != nil {
		log.Fatalf("loading eBPF objects: %v", err)
	}
	defer objs.Close()

	// open filtered_logs database
	db, err := sql.Open("sqlite", "../manager/data/filtered_logs.db")
	if err != nil {
		log.Fatalf("opening filtered_logs.db: %v", err)
	}
	defer db.Close()

	// create map to check wether the containers (and its veth) are still active
	attached := make(map[string]link.Link)

	updateAttachments(db, objs.TcIngressProgram, attached)

	// set up a ticker to poll the database every 5s
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// catch Ctrl+C to cleanly detach
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	log.Printf("<<<<--------------------------------------------------------->>>>")
	log.Printf("                 successfully loaded firewall_container")
	log.Printf("<<<<--------------------------------------------------------->>>>")

	for {
		select {
		case <-ticker.C:
			// re-sync attachments to database changes
			updateAttachments(db, objs.TcIngressProgram, attached)

		case <-sig:
			// detach all before exit
			fmt.Println("\nShutting down, detaching all programs...")
			for name, lnk := range attached {
				lnk.Close()
				fmt.Printf("<< detached from %q\n", name)
			}
			return
		}
	}
}
