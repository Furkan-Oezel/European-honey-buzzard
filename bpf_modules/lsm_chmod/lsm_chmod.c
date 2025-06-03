//go:build ignore

#include "../include_dir/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <linux/errno.h>

char _license[] SEC("license") = "GPL";

struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, __u32);
  __type(value, __u64);
  // max 64 docker containers can be observed
  __uint(max_entries, 64);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} map_container_cgroup_ids SEC(".maps");

SEC("lsm/path_chmod")
int BPF_PROG(path_chmod, const struct path *path, umode_t mode) {
  __u32 key;
  __u64 *value;
  __u64 cgrp_id = bpf_get_current_cgroup_id();
  bpf_printk("current cgroup_id: %llu\n", cgrp_id);

// loop through all cgroup ids
#pragma unroll
  for (int i = 0; i < 64; i++) {
    key = i;
    value = bpf_map_lookup_elem(&map_container_cgroup_ids, &key);
    if (!value)
      continue;

    if (cgrp_id == *value) {
      bpf_printk("matched cgroup_id %llu in map at key %d → blocking chmod\n",
                 cgrp_id, i);
      return -EPERM;
    }
  }

  // else allow chmod
  return 0;
}
