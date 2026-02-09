![](images/european_honey_buzzard.jpg)

# European honey buzzard 

European honey buzzard is a security system.

## How to get European honey buzzard

Clone this repo and cd into it:

```bash
git clone https://github.com/Furkan-Oezel/European-honey-buzzard.git
cd European-honey-buzzard
```

## How to build bpf modules

cd into a subdirectory (e.g. lsm_chmod):

```bash
cd bpf_modules/lsm_chmod
```

Declare a go module:

```bash
go mod init lsm_chmod
go mod tidy
```

Add a dependency on bpf2go:

```bash
go get github.com/cilium/ebpf/cmd/bpf2go
```

Compile bpf C code and build the project:

```bash
go generate
go build -o lsm_chmod
```

Alternatively build for Raspberry Pi (arm64 architecture):

```bash
CGO_ENABLED=0 GOARCH=arm64 go build -o lsm_chmod_arm
```

## How to build John Wick

Declare a go module:

```bash
go mod init john_wick
go mod tidy
```

Compile and build:

```bash
go build -o john_wick main/main.go
```

## How to build the manager

Declare a go module:

```bash
go mod init manager
go mod tidy
```

Compile and build:

```bash
go build -o manager main/main.go
```

## Docker commands

Look for containers:

```bash
docker ps -a
```

Get cgroup id of a given container:

```bash
docker exec <docker id> sh -c 'stat -c %i /sys/fs/cgroup'
```

Stop and remove all containers:

```bash
docker rm -f $(docker ps -aq)
```

Run an alpine container:

```bash
docker run -it alpine
```

Build docker image:

```bash
docker build -t vergil-docker .
```

Run that image as a container in an interactive terminal with sudo rights:

```bash
docker run -it --privileged vergil-docker
```

Look for veth interface on host (second part of the veth pair, the first part of this pair is eth0 inside the container)

```bash
ip link show type veth
```

## Debugging eBPF Programs

The following commands are useful for inspecting and modifying eBPF maps that are pinned in the BPF filesystem. 

### 1. Printing bpf_printk() messages

```bash
sudo cat /sys/kernel/debug/tracing/trace_pipe
```

**Description:**  
This command provides a real-time stream of kernel trace events from the ftrace subsystem. Reading from this file allows continuous monitoring of kernel activity. This file is available when debugfs is mounted and tracing is enabled.

---

### 2. Dumping an eBPF Map

```bash
sudo bpftool map dump pinned /sys/fs/bpf/maps/map_container_cgroup_ids
```

**Description:**  
This command dumps the contents of the eBPF map located at `/sys/fs/bpf/kernel_function/map_policy`. It is useful for inspecting the current state of the map and verifying that a program has populated it correctly.

---

### 3. Listing BPF Filesystem Contents

```bash
sudo ls -lha /sys/fs/bpf/maps
```

**Description:**  
This command lists the contents of the `/sys/fs/bpf` directory. It allows to view all pinned maps. 

---

### 4. Updating an eBPF Map Entry

```bash
sudo bpftool map update pinned /sys/fs/bpf/kernel_function/map_policy key hex 04 00 00 00 value hex 61 70 70 6c 65 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
```

**Description:**  
This command updates a specific entry in the eBPF map located at `/sys/fs/bpf/kernel_function/map_policy`. It writes a new value (`"apple"`, encoded in hexadecimal) to the key (`04 00 00 00`).

---

## SQLite CLI

How to view container_logs database:

```bash
sqlite3 manager/data/container_logs.db
```

Then run inside the prompt:

```bash
SELECT * FROM container_logs;
```

How to view filtered_logs database:

```bash
sqlite3 manager/data/filtered_logs.db
```

Then run inside the prompt:

```bash
SELECT * FROM filtered_logs;
```

To quit the SQLite prompt:

```bash
.quit
```

## Verify Docker veth TCP-port firewall

Launch a HTTP server container:

```bash
docker run -d --name srv --network bridge nginx:alpine
```

Grab its bridge-network IP:

```bash
SRV_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' srv)
echo "Server IP: $SRV_IP"
```

ALLOWED test  —  source port 1234 → destination port 80:

```bash
printf 'GET / HTTP/1.0\r\nHost: %s\r\n\r\n' "$SRV_IP" \
  | nc -p 1234 -w 2 "$SRV_IP" 80 \
  | head -n1 && echo "allowed"

```

BLOCKED test  —  source port 5555 → destination port 80:

```bash
printf 'GET / HTTP/1.0\r\nHost: %s\r\n\r\n' "$SRV_IP" \
  | nc -p 5555 -w 2 "$SRV_IP" 80 \
  && echo "→ ???" || echo "blocked"

```
