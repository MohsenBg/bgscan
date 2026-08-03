---
title: "نصب و راه‌اندازی"
weight: 3
bookFlatSection: false
---

# نصب و راه‌اندازی

bgscan روی Linux، macOS، Windows و Android (Termux) اجرا می‌شود. روشی را انتخاب کنید که برای سیستم شما مناسب‌تر است.

## نصب سریع

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**Android (Termux)**

```bash
pkg update -y && pkg install bash curl unzip -y && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

نصب‌کننده آخرین نسخه را دانلود می‌کند، آن را داخل پوشهٔ `bgscan/` باز می‌کند و برنامه را آمادهٔ اجرا می‌کند. اگر نسخه‌ای از قبل نصب شده باشد، از شما می‌پرسد که آن را جایگزین کند یا با نام `bgscan_old` از آن پشتیبان بگیرد.

## نصب دستی

1. فایل ZIP مناسب سیستم‌عامل خود را از [صفحهٔ Releases](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.
2. فایل ZIP را باز کنید.
3. **برنامه را اجرا کنید:**
   - **Linux، macOS و Termux:** ترمینال را باز کنید، وارد پوشهٔ برنامه شوید و `./bgscan` را اجرا کنید.
   - **Windows:** روی `bgscan.exe` دو بار کلیک کنید یا در PowerShell دستور `./bgscan.exe` را اجرا کنید.

در اولین اجرا، پوشهٔ `settings/` با فایل‌های تنظیمات پیش‌فرض و پوشهٔ `ips/` با فهرست‌های IP آماده ساخته می‌شود.

## ساخت از سورس

> **نکته:** bgscan را نمی‌توان با `go install` نصب کرد. برای ساخت برنامه باید از ابزار `bgscan-builder` که داخل مخزن قرار دارد استفاده کنید.

### پیش‌نیازها

- Go نسخهٔ 1.26.3 یا جدیدتر
- Git

### دریافت سورس و ساخت برنامه

```bash
# دریافت سورس
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan

# نصب ابزار builder
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# ساخت برای سیستم فعلی
./bg-builder

# یا ساخت برای یک سیستم مشخص
./bg-builder --os linux --arch amd64
./bg-builder --os windows --arch amd64
./bg-builder --os darwin --arch arm64
./bg-builder --os android --arch arm64
```

## ارتقا

برای ارتقا، دوباره اسکریپت [نصب سریع](#نصب-سریع) را اجرا کنید. اسکریپت نسخهٔ قبلی را پیدا می‌کند و به شما اجازه می‌دهد آن را جایگزین کنید یا از آن پشتیبان بگیرید.

اگر تنظیمات خودتان را تغییر داده‌اید:

- فایل‌های `settings/*.toml` را به نصب جدید منتقل کنید.
- فهرست‌های IP شخصی خود را از پوشهٔ `ips/` کپی کنید.
- قبل از جایگزین‌کردن فایل‌ها، مطمئن شوید bgscan در حال اجرا نیست.

## نیازمندی‌ها

- **سیستم‌عامل:** Linux، macOS، Windows 10 یا جدیدتر، یا Android 7.0 یا جدیدتر (Termux)
- **ابزارها:** `curl`، `unzip` و `bash`؛ نصب‌کننده در بیشتر سیستم‌ها وابستگی‌های لازم را خودش آماده می‌کند
- **Windows:** PowerShell نسخهٔ 5.1 یا جدیدتر
- **Termux:** نسخهٔ F-Droid را نصب کنید؛ نسخهٔ Play Store قدیمی است
