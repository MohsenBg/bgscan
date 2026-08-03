---
title: "Scanner"
weight: 2
bookFlatSection: true
bookCollapseSection: true
---

# Scanner

The scanner runs probes against a list of IP addresses. Each probe type (ICMP, TCP, HTTP, DNS, Xray) can run on its own, or several can be chained so that each stage filters the targets for the next.

- [Scan Source](./scan-source/) — where the target IPs come from
- [Scan Types](./scan-types/) — what each probe does
- [Scan Pipeline](./scan-pipeline/) — how stages are chained
- [IP Lists](./ip-files/) — input files
- [Result Files](./result-files/) — output files
- [Xray Outbounds](./xray-outbound/) — outbound templates for Xray scans
