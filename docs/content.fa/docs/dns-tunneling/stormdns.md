---
title: "StormDNS"
weight: 5
---

# تنظیمات StormDNS

StormDNS داخل خود BGScan و از طریق یک کتابخانه کلاینت داخلی اجرا می‌شه. ساختارش شبیه MasterDNS هست، اما پیش‌فرض‌های خودش رو داره و می‌تونید نوع کوئری DNS مورد استفاده برای انتقال ترافیک تونل رو هم انتخاب کنید.

این پروتکل تنظیمات پروکسی نداره.

## تنظیمات اتصال

| فیلد | پیش‌فرض | توضیح |
| --- | --- | --- |
| `domain` | الزامی | دامنه‌ای که سرور StormDNS شما سرویس می‌ده |
| `data_enc_method` | `1` | روش رمزنگاری: `0` = بدون رمزنگاری، `1` = XOR که پیش‌فرض است، `2` = ChaCha20، `3` = AES-128-GCM، `4` = AES-192-GCM، `5` = AES-256-GCM |
| `encryption_key` | الزامی، مگر در حالت بدون رمزنگاری | کلید ۳۲ کاراکتری Hex که باید با کلید سرور مطابقت داشته باشه |
| `resolver_port` | `53` | پورت ریزالور |
| `dns_query_type` | `TXT` | نوع کوئری مورد استفاده برای انتقال ترافیک تونل: `TXT`، `NS`، `CNAME` یا `ROTATE`، مقدار خالی معادل `TXT` هست |

## تنظیمات MTU و Session

| فیلد | پیش‌فرض | توضیح |
| --- | --- | --- |
| `mtu_test_timeout_sec` | `2.0` | مدت انتظار برای هر بسته آزمایش MTU، برحسب ثانیه |
| `mtu_test_retries` | `2` | تعداد تلاش برای هر اندازه MTU پیش از رفتن به اندازه بعدی |
| `session_init_retry_max_sec` | `2.0` | مدت تلاش مجدد برای راه‌اندازی Session، پیش از شروع دوباره |
| `min_upload_mtu` / `max_upload_mtu` | `38` / `150` | حداقل و حداکثر اندازه MTU آپلود، برحسب بایت |
| `min_download_mtu` / `max_download_mtu` | `100` / `500` | حداقل و حداکثر اندازه MTU دانلود، برحسب بایت |
| `mtu_parallelism` | `1` | تعداد Workerهای هم‌زمان برای آزمایش MTU |
| `rx_tx_workers` | `4` | تعداد جریان‌های داده هم‌زمان در تونل |

## نمونه Config

```toml id="b7m2qx"
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
dns_query_type = 'TXT'
mtu_test_timeout_sec = 2.0
mtu_test_retries = 2
session_init_retry_max_sec = 2.0
min_upload_mtu = 38
max_upload_mtu = 150
min_download_mtu = 100
max_download_mtu = 500
mtu_parallelism = 1
rx_tx_workers = 4
```

## موضوعات مرتبط

- [تونل DNS](../): آشنایی با پروتکل‌ها و روند اسکن
- [MasterDNS](./masterdns/): پروتکلی با ساختار مشابه و تنظیمات پیش‌فرض مخصوص خودش

