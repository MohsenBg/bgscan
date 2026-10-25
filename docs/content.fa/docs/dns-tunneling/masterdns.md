---
title: "MasterDNS"
weight: 4
---

# تنظیمات MasterDNS

MasterDNS داخل خود BGScan و با استفاده از یک کتابخانه کلاینت داخلی اجرا می‌شه. تونل اندازه MTU رو به‌صورت خودکار مذاکره می‌کنه و ترافیک رو با روش رمزنگاری انتخاب‌شده محافظت می‌کنه.

این پروتکل تنظیمات پروکسی نداره و پروب مستقیماً به ریزالور متصل می‌شه.

## تنظیمات اتصال

| فیلد | پیش‌فرض | توضیح |
| --- | --- | --- |
| `domain` | الزامی | دامنه‌ای که سرور MasterDNS شما سرویس می‌ده |
| `data_enc_method` | `1` | روش رمزنگاری: `0` = بدون رمزنگاری، `1` = XOR که پیش‌فرض است، `2` = ChaCha20، `3` = AES-128-GCM، `4` = AES-192-GCM، `5` = AES-256-GCM |
| `encryption_key` | الزامی، مگر در حالت بدون رمزنگاری | کلید ۳۲ کاراکتری Hex که باید با کلید سرور مطابقت داشته باشه |
| `resolver_port` | `53` | پورت ریزالور |

## تنظیمات MTU و Session

| فیلد | پیش‌فرض | توضیح |
| --- | --- | --- |
| `mtu_test_timeout_sec` | `3.0` | مدت انتظار برای هر بسته آزمایش MTU، برحسب ثانیه |
| `mtu_test_retries` | `2` | تعداد تلاش برای هر اندازه MTU پیش از رفتن به اندازه بعدی |
| `session_init_retry_max_sec` | `2.0` | مدت تلاش مجدد برای راه‌اندازی Session، پیش از شروع دوباره |
| `session_init_racing_count` | `1` | تعداد تلاش‌های موازی برای راه‌اندازی Session |
| `min_upload_mtu` / `max_upload_mtu` | `38` / `150` | حداقل و حداکثر اندازه MTU آپلود، برحسب بایت |
| `min_download_mtu` / `max_download_mtu` | `100` / `500` | حداقل و حداکثر اندازه MTU دانلود، برحسب بایت |
| `mtu_parallelism` | `1` | تعداد Workerهای هم‌زمان برای آزمایش MTU |
| `rx_tx_workers` | `4` | تعداد جریان‌های داده هم‌زمان در تونل |

## نمونه Config

```toml id="j3p9xv"
domain = 'example.com'
data_enc_method = 1
encryption_key = 'cd6d78e954f48f62cb74cdcf8a2459d3'
resolver_port = 53
mtu_test_timeout_sec = 3.0
mtu_test_retries = 2
session_init_retry_max_sec = 2.0
session_init_racing_count = 1
min_upload_mtu = 38
max_upload_mtu = 150
min_download_mtu = 100
max_download_mtu = 500
mtu_parallelism = 1
rx_tx_workers = 4
```

## موضوعات مرتبط

- [تونل DNS](../): آشنایی با پروتکل‌ها و روند اسکن
- [StormDNS](./stormdns/): پروتکلی با ساختار مشابه و تنظیمات پیش‌فرض و گزینه‌های انتخاب نوع کوئری مخصوص خودش

