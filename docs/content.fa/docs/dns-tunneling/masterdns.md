---
title: "MasterDNS"
weight: 4
---

# تنظیمات MasterDNS

MasterDNS با کتابخانهٔ داخلی خودش داخل برنامه اجرا می‌شود. اندازهٔ MTU به‌طور خودکار مذاکره می‌شود و ترافیک با روش رمزنگاری قابل تنظیم محافظت می‌گردد. تنظیمات Proxy ندارد و پروب مستقیم به Resolver وصل می‌شود.

## تنظیمات اتصال

| فیلد | پیش‌فرض | توضیحات |
| --- | --- | --- |
| `Domain` | (لازم) | Zone سرویس‌دهی‌شده توسط سرور MasterDNS شما |
| `DataEncMethod` | `1` | رمزنگاری: `0` بدون رمزنگاری، `1` پیش‌فرض (XOR)، `2` یعنی ChaCha20، `3` یعنی AES-128-GCM، `4` یعنی AES-192-GCM، `5` یعنی AES-256-GCM |
| `EncryptionKey` | (لازم، مگر وقتی روش None است) | ۳۲ کاراکتر hex که باید با سرور یکی باشد |
| `ResolverPort` | `53` | پورت Resolver |

## تنظیم MTU و Session

| فیلد | پیش‌فرض | توضیحات |
| --- | --- | --- |
| `MTUTestTimeoutSec` | `3.0` | انتظار برای هر بستهٔ پروب MTU، ثانیه |
| `MTUTestRetries` | `2` | تلاش برای هر اندازهٔ MTU قبل از رفتن به اندازهٔ بعد |
| `SessionInitRetryMaxSec` | `2.0` | پنجرهٔ تلاش برای شروع Session قبل از شروع دوباره |
| `SessionInitRacingCount` | `1` | تلاش‌های موازی برای شروع Session |
| `MinUploadMTU` / `MaxUploadMTU` | `38` / `150` | بازهٔ MTU آپلود، بایت |
| `MinDownloadMTU` / `MaxDownloadMTU` | `100` / `500` | بازهٔ MTU دانلود، بایت |
| `MTUParallelism` | `1` | Workerهای موازی تست MTU |
| `RxTxWorkers` | `4` | استریم‌های هم‌زمان دادهٔ تونل |

## نمونهٔ Config

```toml
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
mtu_test_timeout_sec = 2.0
mtu_test_retries = 2
session_init_retry_max_sec = 60.0
session_init_racing_count = 3
min_upload_mtu = 38
max_upload_mtu = 150
min_download_mtu = 100
max_download_mtu = 500
mtu_parallelism = 3
rx_tx_workers = 2
```

## موضوعات مرتبط

- [تونل DNS](../) — نمای کلی پروتکل‌ها و نحوهٔ اسکن
- [StormDNS](../stormdns/) — همان مدل با پیش‌فرض‌های خودش و انتخاب نوع کوئری
