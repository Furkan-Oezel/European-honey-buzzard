package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

func main() {
	// Allow locking memory for eBPF
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("failed to set memlock rlimit: %v", err)
	}

	// Load the compiled eBPF ELF and load it into the kernel.
	var objs firewall_containerObjects
	if err := loadFirewall_containerObjects(&objs, nil); err != nil {
		log.Fatal("Loading eBPF objects:", err)
	}
	defer objs.Close()

	// hardcoded second part of a virtuel ethernet pair of a docker container
	ifaceName := "veth23de0e0"
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("could not find interface %s: %v", ifaceName, err)
	}

	lnk, err := link.AttachTCX(link.TCXOptions{
		Program:   objs.TcIngressProgram,
		Interface: iface.Index,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		log.Fatalf("failed to attach TC program: %v", err)
	}
	defer lnk.Close()

	fmt.Printf("eBPF tc program attached to %s ingress\n", ifaceName)
	fmt.Println("Waiting for packets... press Ctrl+C to exit")

	// Wait for Ctrl+C to exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("\nDetaching program and exiting")
}
