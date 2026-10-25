---
title: "Scan Source"
weight: 1
---

# Scan source

Before the scan starts, you must define which addresses bgscan should check. For this, go to **Main Menu → Run Scan** and open the scan source picker. Here you have two options:

| Key | Option | Description |
|---|---|---|
| `i` | IP List | Pick a file from the IP lists |
| `r` | Result List | Pick a result file from an earlier scan |

---

{{< img "/bgscan-target-source.webp" "bgscan scan source" >}}

## IP List

Choosing **IP List** opens the IP file list. Pick the file you want from the `ips/` directory so bgscan uses the addresses inside it for the scan.

This option is for when you want to check a new set of IPs. For example, ranges from a CDN, addresses from a country, or a list you prepared yourself.

To learn how to import and manage files, see the [IP lists](./ip-files.md) page.

---

## Result List

Choosing **Result List** opens the result file list. Pick one of the files you saved earlier so bgscan scans the addresses inside it again.

This method is useful when you want to do a closer check on the results of an earlier scan or run a different scan type. For example, you can choose the IPs that passed an ICMP pre-scan for an Xray scan.

To learn about the structure of these files, see the [Result files](./result-files.md) page.

---

## Next step

After you choose a file, bgscan takes you to the **Scan Type** selection stage. Here you define which probe runs on the chosen addresses.

Read the description of each probe on the [Scan types](./scan-types.md) page.
