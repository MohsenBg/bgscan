---
title: "Slipstream"
weight: 3
---

# Slipstream settings

Slipstream runs inside BGScan itself through `libslipstream-client`. This client library loads with `purego`, so there is no need for a separate binary and BGScan does not look for anything in `PATH`.

Each probe starts a tunnel and checks that traffic passes through it using a local SOCKS5 listener on the port assigned to that probe.

## Connection settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `domain` | Required | Valid domain | The domain your Slipstream server serves |
| `resolver_port` | `53` | Greater than `0` | The resolver port |
| `cert_path` | `""` | File path | The CA certificate for validating the server. An empty value means using the system certificate store |
| `dns_resolution` | `"recursive"` | `recursive`, `authoritative` | How the client resolves names |
| `gso` | `false` | `true`, `false` | Enables Generic Segmentation Offload for UDP |
| `congestion_control` | `"cubic"` | `bbr`, `cubic` | The tunnel congestion control algorithm. An empty value means using the library default |

## Proxy and authentication

Slipstream uses the same proxy fields as DNSTT. For details, see the [proxy and authentication](./dnstt/#proxy-and-authentication) section.

To use the SSH proxy, the `known_hosts_file` field must point to a valid `known_hosts` file. This field is not shown in the TUI form, so you must enter its path directly in the TOML file.

## Library location

BGScan checks these directories in order to find the client library for its platform:

1. `<bgscan-root>/assets/slipstream-client/`
2. `<bgscan-root>/assets/slipstream/`
3. `<bgscan-root>/assets/`
4. `<bgscan-root>/`

The file names it looks for depend on the platform:

| Platform | File names |
| --- | --- |
| Linux / Android | `libslipstream-client.so`, `libslipstream-client-*.so` |
| macOS | `libslipstream-client.dylib`, `libslipstream-client-*.dylib` |
| Windows | `slipstream-client.dll`, `libslipstream-client.dll`, `libslipstream-client-*.dll` |

The build scripts download the library and the file is not kept in the repository. When the program starts, BGScan checks that the library loads and reports its version. If the library is not found, the Slipstream scan cannot run and an error is reported.

## Example config

```toml id="k4v9sz"
domain = "example.com"
resolver_port = 53
cert_path = ""
dns_resolution = "recursive"
gso = false
congestion_control = "cubic"
proxy_type = "socks"
proxy_port = 1080
auth_method = "none"
username = ""
password = ""
private_key = ""
known_hosts_file = ""
```

## Related topics

- [DNS tunnel](../): an introduction to the protocols and the scan flow
- [DNSTT](./dnstt/): descriptions of the proxy and authentication fields
