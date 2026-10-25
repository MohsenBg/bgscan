---
title: "What is bgscan"
weight: 1
bookComments: false
---

# bgscan

bgscan is a scanner for many protocols, written in Go, and it runs from the terminal. It checks IPv4 and IPv6 addresses with ICMP, TCP, DNS, HTTP (including HTTP/3), Xray, and DNS tunnels such as DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed.

The interface is fully terminal based and you drive it with the keyboard.

With bgscan you can check whether a connection works, compare different endpoints, or test Xray outbounds. Results appear as the scan runs, and you can save them as CSV so a later scan can use them again.

<div align="center">
  <video
    src="https://github.com/user-attachments/assets/4a4c216e-0179-427c-8fe1-89a8fa197730"
    autoplay
    muted
    loop
    playsinline
    controls
    width="100%"
    style="max-width: 900px; border-radius: 12px; border: 1px solid #30363d;"
  ></video>
</div>

