---
title: "Scan Pipeline"
weight: 3
---

# Scan Pipeline

bgscan supports multi-stage scanning where the output of one scan becomes the input for the next. This is configured via the [General Settings](../settings/general.md) under "Pipeline Execution Mode".

## Pipeline Modes

bgscan offers three pipeline modes to balance performance, memory usage, and flexibility:

#### 1. Streaming (Default)

```
Stage 1 (IPs) → [ICMP Scan] → [TCP Scan] → [HTTP Scan] → Results
                   │           │           │
                   ▼           ▼           ▼
              (passed IPs)  (passed IPs)  (passed IPs)
```

- **How it works**: All stages run simultaneously, with IPs flowing between stages via in-memory channels
- **Memory usage**: Highest (holds multiple copies of IP lists in memory)
- **Speed**: Fastest (no waiting for stages to complete)
- **Best for**: High-performance scanning with sufficient memory
- **When to use**: When you have enough memory and want maximum throughput

#### 2. Sequential

```
Stage 1 → [ICMP Scan] → (wait) → [TCP Scan] → (wait) → [HTTP Scan] → Results
```

- **How it works**: Each stage waits for the previous to complete before starting
- **Memory usage**: Lowest (only one stage's data in memory at a time)
- **Speed**: Slowest (total time is sum of all stage times)
- **Best for**: Memory-constrained environments
- **When to use**: When running on systems with limited RAM

#### 3. Batch

```
[Batch 1 of IPs] → [ICMP Scan] → [TCP Scan] → [HTTP Scan] → Results
[Batch 2 of IPs] → [ICMP Scan] → [TCP Scan] → [HTTP Scan] → Results
...
```

- **How it works**: IPs are divided into batches; each batch goes through all stages before the next batch starts
- **Memory usage**: Moderate (holds one batch's data per stage)
- **Speed**: Moderate (better than sequential, worse than streaming)
- **Best for**: Balancing memory and performance
- **When to use**: When you want to limit memory spikes while maintaining decent throughput

## Configuring the Pipeline

The pipeline mode is set in [`general_settings.toml`](../settings/general.md):

```toml
# Pipeline execution mode: "streaming", "sequential", or "batch"
pipeline_mode = "streaming"
```

## Stage Configuration

Each stage in the pipeline uses its own configuration file:

- ICMP Stage: [`icmp_settings.toml`](../settings/icmp.md)
- TCP Stage: [`tcp_settings.toml`](../settings/tcp.md)
- HTTP Stage: [`http_settings.toml`](../settings/http.md)
- DNS Stage: [`dns_settings.toml`](../settings/dns.md)
- Xray Stage: [`xray_settings.toml`](../settings/xray.md)

## Data Flow Between Stages

Only IPs that pass a stage's success criteria proceed to the next stage:

1. **ICMP Stage**: Only IPs that respond to ping (within timeout/tries) proceed
2. **TCP Stage**: Only IPs with the specified port open proceed
3. **HTTP Stage**: Only IPs returning an accepted status code proceed
4. **DNS Stage**: Only IPs that respond with valid DNS answers proceed
5. **Xray Stage**: Only IPs that pass the connectivity test (if pre-scan enabled) proceed to bandwidth test

## Example: ICMP → TCP → HTTP Pipeline

1. **Input**: 10,000 IPs from IP list
2. **ICMP Stage**: 2,000 IPs respond to ping → 2,000 IPs passed to TCP stage
3. **TCP Stage**: 500 IPs have port 80 open → 500 IPs passed to HTTP stage
4. **HTTP Stage**: 300 IPs return HTTP 200 → 300 successful web servers found
5. **Output**:
   - `results/icmp/` contains results for all 10,000 IPs
   - `results/tcp/` contains results for the 2,000 ICMP-responsive IPs
   - `results/http/` contains results for the 500 TCP-responsive IPs
   - Final count: 300 working web servers

## Related Topics

- [Scanner Overview](../scanner.md)
- [Scan Types](./scan-types.md)
- [General Settings](../settings/general.md) - Pipeline mode configuration
