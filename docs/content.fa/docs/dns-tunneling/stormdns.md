---
title: "StormDNS"
weight: 5
---

# تنظیمات StormDNS

StormDNS هم با کتابخانهٔ داخلی خودش داخل برنامه اجرا می‌شود. شبیه MasterDNS است با پیش‌فرض‌های خودش، به‌علاوهٔ انتخاب نوع کوئری DNS که ترافیک تونل را حمل می‌کند. تنظیمات Proxy ندارد.

## تنظیمات اتصال

| فیلد | پیش‌فرض | توضیحات |
| --- | --- | --- |
| `Domain` | (لازم) | Zone سرویس‌دهی‌شده توسط سرور StormDNS شما |
| `DataEncMethod` | `1` | رمزنگاری: `0` بدون رمزنگاری، `1` پیش‌فرض (XOR)، `2` یعنی ChaCha20، `3` یعنی AES-128-GCM، `4` یعنی AES-192-GCM، `5` یعنی AES-256-GCM |
| `EncryptionKey` | (لازم، مگر وقتی روش None است) | ۳۲ کاراکتر hex که باید با سرور یکی باشد |
| `ResolverPort` | `53` | پورت Resolver |
| `DNSQueryType` | `'TXT'` | نوع کوئری حامل تونل: `TXT`، `NS`، `CNAME` یا `ROTATE` (خالی یعنی TXT) |

## تنظیم MTU و Session

| فیلد | پیش‌فرض | توضیحات |
| --- | --- | --- |
| `MTUTestTimeoutSec` | `2.0` | انتظار برای هر بستهٔ پروب MTU، ثانیه |
| `MTUTestRetries` | `3` | تلاش برای هر اندازهٔ MTU قبل از رفتن به اندازهٔ بعد |
| `SessionInitRetryMaxSec` | `60.0` | پنجرهٔ تلاش برای شروع Session قبل از شروع دوباره |
| `MinUploadMTU` / `MaxUploadMTU` | `100` / `200` | بازهٔ MTU آپلود، بایت |
| `MinDownloadMTU` / `MaxDownloadMTU` | `1000` / `4000` | بازهٔ MTU دانلود، بایت |
| `MTUParallelism` | `16` | Workerهای موازی تست MTU |
| `RxTxWorkers` | `4` | استریم‌های هم‌زمان دادهٔ تونل |

## نمونهٔ Config

```toml
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
dns_query_type = 'TXT'
mtu_test_timeout_sec = 2.0
mtu_test_retries = 3
session_init_retry_max_sec = 60.0
min_upload_mtu = 100
max_upload_mtu = 200
min_download_mtu = 1000
max_download_mtu = 4000
mtu_parallelism = 16
rx_tx_workers = 4
```

## موضوعات مرتبط

- [تونل DNS](../) — نمای کلی پروتکل‌ها و نحوهٔ اسکن
- [MasterDNS](../masterdns/) — همان مدل با پیش‌فرض‌های خودش
