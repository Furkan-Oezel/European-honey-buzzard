//go:build ignore

#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>

SEC("tc")
int tc_ingress_program(struct __sk_buff *skb) {
  bpf_printk("Packet received on veth23de0e0\n");
  return TC_ACT_OK;
}

char LICENSE[] SEC("license") = "GPL";
