
//go:build ignore

/*
 * add kernel type definitions (e.g. path->dentry->d_parent->d_name.name,
 * which is the name of the directory in the current path)
 * how to get this file: bpftool btf dump file /sys/kernel/btf/vmlinux format c
 * > vmlinux.h
 */
#include "../include_dir/vmlinux.h"
// add bpf helper functions (e.g bpf_map_lookup_elem(), bpf_map_update_elem())
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <linux/errno.h>
// #include <stdint.h>

char _license[] SEC("license") = "GPL";

struct {
  // declare pointer called 'type' that points to a int array of the size
  // 'BPF_MAP_TYPE_ARRAY' (2)
  __uint(type, BPF_MAP_TYPE_ARRAY);
  // declare pointer called 'key' that is of the type '__u32'
  __type(key, __u32);
  // declare pointer called 'value' that is of the type '__u64'
  __type(value, __u64);
  // declare pointer called 'max_entries' that points to a int array of the size
  // 5
  __uint(max_entries, 5);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} map_policy SEC(".maps");

/*
 * available LSM hooks: https://www.kernel.org/doc/html/v5.2/security/LSM.html
 * how to get BPF_PROG function declaration for the path_chmod LSM hook:
 * grep path_chmod oth/X/kernel/linux-6.10/include/linux/lsm_hook_defs.h
 */
SEC("lsm/path_chmod")
/*
 * this program executes whenever a chmod command is performed
 * it reads the name of the directory in which the command is performed
 * and compares it to a given string, "furkan" in this case
 * if it is the same, chmod will not be executed
 * the path argument is the kernel data structure representing the file
 * the mode argument is the desired new mode value
 */

int BPF_PROG(path_chmod, const struct path *path, umode_t mode) {
  int key = 0;
  __u64 *value = bpf_map_lookup_elem(&map_policy, &key);
  if (value)
    bpf_printk("Value read from the map: '%llu'\n", *value);
  else
    bpf_printk("Failed to read value from the map\n");

  ///
  __u64 cgrp_id = bpf_get_current_cgroup_id();
  bpf_printk("lets gooo cgroup_id: %llu\n", cgrp_id);
  if (cgrp_id == *value) {
    bpf_printk("lets gooo cgroup_id: %llu\n", cgrp_id);
    bpf_printk("chmod not allowed");
    return -EPERM;
  }
  ///
  // return -EPERM;
  return 0;
}
