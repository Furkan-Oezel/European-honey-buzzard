
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

/*=============================================*/
/*              *map config*                   */
/* type = array                                */
/* number of entries = 5                       */
/* 1. entry =    */
/* 2. entry =    */
/* 3. entry =    */
/* 4. entry =    */
/* 5. entry =    */
/*=============================================*/
struct {
  // declare pointer called 'type' that points to a int array of the size
  // 'BPF_MAP_TYPE_ARRAY' (2)
  __uint(type, BPF_MAP_TYPE_ARRAY);
  // declare pointer called 'key' that is of the type '__u32'
  __type(key, __u32);
  // declare pointer called 'value' that is of the type 'char' (points to first
  // element of a char array)
  __type(value, __u64);
  // declare pointer called 'max_entries' that points to a int array of the size
  // 5
  __uint(max_entries, 5);
  // In order for the ELF loader to automatically pin or re-use a pinned map,
  // the map definition needs to have its pinned flag set.
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} map_policy SEC(".maps");

char _license[] SEC("license") = "GPL";
