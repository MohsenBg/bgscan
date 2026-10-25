---
title: "Scanner"
weight: 2
bookFlatSection: true
bookCollapseSection: true
---

# Scanner

The scanner runs different probes on a list of IP addresses. Most scans run on their own, but some can have an initial check stage before the main scan.

For DNS tunnels you can add a resolver pre-scan stage. An Xray scan can also check the targets with TCP, ICMP, or HTTP before it starts. This way only the targets that respond reach the next stage, and no time is wasted on slower checks for the rest.

The scan engine also supports longer chains, but the terminal interface (TUI) currently offers only these pre-scan combinations.

- [Scan source](./scan-source/): where the target addresses come from
- [Scan types](./scan-types/): what each probe does
- [Scan stages](./scan-pipeline/): how the scan stages connect
- [IP lists](./ip-files/): the scan input files
- [Result files](./result-files/): the scan output files
- [Xray outbounds](./xray-outbound/): outbound templates for scanning with Xray
