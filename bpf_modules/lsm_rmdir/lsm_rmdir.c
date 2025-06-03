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
  __uint(max_entries, 64);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} map_container_cgroup_ids SEC(".maps");

SEC("lsm/path_rmdir")
int BPF_PROG(path_rmdir, const struct path *path, struct dentry *dentry) {
  __u32 key;
  __u64 *value;
  __u64 cgrp_id = bpf_get_current_cgroup_id();

// loop through all cgroup ids
#pragma unroll
  for (int i = 0; i < 64; i++) {
    key = i;
    value = bpf_map_lookup_elem(&map_container_cgroup_ids, &key);
    if (!value)
      continue;

    if (cgrp_id == *value) {
      bpf_printk("matched cgroup_id %llu in map at key %d → blocking rmdir\n",
                 cgrp_id, i);
      return -EPERM;
    }
  }

  // else allow rmdit
  return 0;
}

////go:build ignore
//
///*
// * add kernel type definitions (e.g. path->dentry->d_parent->d_name.name,
// * which is the name of the directory in the current path)
// * how to get this file: bpftool btf dump file /sys/kernel/btf/vmlinux format
// c
// * > vmlinux.h
// */
// #include "../include_dir/vmlinux.h"
//// add bpf helper functions (e.g bpf_map_lookup_elem(), bpf_map_update_elem())
// #include <bpf/bpf_helpers.h>
// #include <bpf/bpf_tracing.h>
// #include <linux/errno.h>
// #include <string.h>
//
///*
// * available LSM hooks: https://www.kernel.org/doc/html/v5.2/security/LSM.html
// * how to get BPF_PROG function declaration for the path_rmdir LSM hook:
// * grep path_rmdir oth/X/kernel/linux-6.10/include/linux/lsm_hook_defs.h
// */
// SEC("lsm/path_rmdir")
///*
// * this program executes whenever a rmdir command is performed
// * it forbids the deletion of the directory furkan
// */
// int BPF_PROG(path_rmdir, const struct path *path, struct dentry *dentry) {
//  char buf[32];
//  // read the directory name that is to be removed
//  bpf_probe_read_str(buf, sizeof(buf), dentry->d_name.name);
//  if (strncmp(buf, "furkan", 6) == 0) {
//    bpf_printk("rmdir attempted on the directory: '%s'\n", buf);
//    return -EPERM;
//  }
//  return 0;
//}
//
// char _license[] SEC("license") = "GPL";
