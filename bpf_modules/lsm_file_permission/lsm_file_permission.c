//go:build ignore

#include "../include_dir/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <linux/errno.h>
#include <string.h>

#define EXECUTE 0x1
#define WRITE 0x2

char _license[] SEC("license") = "GPL";

struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, __u32);
  __type(value, __u64);
  __uint(max_entries, 64);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} map_container_cgroup_ids SEC(".maps");

SEC("lsm/file_permission")
int BPF_PROG(file_permission, struct file *file, int mask) {
  char filename[256];
  const char *suffix = "confidential";
  int suffix_len = 12;

  __u32 key;
  __u64 *value;
  __u64 cgrp_id = bpf_get_current_cgroup_id();
  bool cgroupid_exists;

// loop through all cgroup ids
#pragma unroll
  for (int i = 0; i < 64; i++) {
    key = i;
    value = bpf_map_lookup_elem(&map_container_cgroup_ids, &key);
    if (!value)
      continue;

    if (cgrp_id == *value) {
      cgroupid_exists = true;
    }
  }

  if (cgroupid_exists) {
    // read filename that is being accessed
    bpf_probe_read_str(filename, sizeof(filename),
                       file->f_path.dentry->d_name.name);

    int filename_len = 0;

    // iterate through the filename and count the letters
#pragma unroll
    for (int i = 0; i < sizeof(filename); i++) {
      if (filename[i] == '\0')
        break;
      filename_len++;
    }

    // make sure the filename is long enough to possibly end with "confidential"
    // If it's shorter than 12 characters, it can’t match, so skip the check
    if (filename_len >= suffix_len) {
      /*
       * &filename[filename_len - suffix_len] → points to the last N characters
       * of the filename. e.g. banana.confidential = 19 letters. 19 - suffix_len
       * = 19 - 12 = 7. -> pointer = &filename[7] = "confidential" in this case.
       * Lastly both pointers get compared.
       */
      if (memcmp(&filename[filename_len - suffix_len], suffix, suffix_len) ==
          0) {
        bpf_printk("Access to confidential file: %s\n", filename);
        bpf_printk("Requested access mask: %d\n", mask);

        // was this a write or execute access?
        if ((mask & WRITE) | (mask & EXECUTE)) {
          bpf_printk("Write/execute denied for file: %s\n", filename);
          return -EPERM;
        }
      }
    }
  }
  return 0;
}
