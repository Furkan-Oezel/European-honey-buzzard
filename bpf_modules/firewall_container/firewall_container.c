//go:build ignore

#include <linux/bpf.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/tcp.h>
#include <netinet/in.h>
#include <stdbool.h>

#define ALLOWED_SRC_PORT 1234
#define ALLOWED_DST_PORT 80

SEC("tc")
// *skb is a pointer to a bpf context struct that the kernel hands this bpf
// program when it's invoked at the tc hook.
int tc_ingress_program(struct __sk_buff *skb) {
  // set up pointers to the start/end of packet data
  void *data = (void *)(long)skb->data;
  void *data_end = (void *)(long)skb->data_end;

  // parse Ethernet header
  struct ethhdr *eth = data;
  if ((void *)eth + sizeof(*eth) > data_end)
    return TC_ACT_SHOT; // if malformed, drop

  // check ethernet protocol
  if (eth->h_proto != bpf_htons(ETH_P_IP))
    return TC_ACT_OK; // if packet is not IPv4, let it through

  // parse IP header
  struct iphdr *ip = data + sizeof(*eth);
  if ((void *)ip + sizeof(*ip) > data_end)
    return TC_ACT_SHOT; // if malformed, drop

  // check for TCP
  if (ip->protocol != IPPROTO_TCP)
    return TC_ACT_OK; // if packet is not TCP, let it through

  // compute IP header length in bytes
  __u32 ip_hdr_len = ip->ihl * 4;

  // parse TCP header
  struct tcphdr *tcp = (void *)ip + ip_hdr_len;
  if ((void *)tcp + sizeof(*tcp) > data_end)
    return TC_ACT_SHOT; // if malformed, drop

  // read ports (in network order → convert to host order)
  __u16 src_port = bpf_ntohs(tcp->source);
  __u16 dst_port = bpf_ntohs(tcp->dest);

  // enforce TCP port policy (both directions)
  bool forward = (src_port == 1234 && dst_port == 80);
  bool back = (src_port == 80 && dst_port == 1234);
  if (!(forward || back)) {
    // red light
    return TC_ACT_SHOT;
  }

  // green light
  return TC_ACT_OK;
}

char LICENSE[] SEC("license") = "GPL";
