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
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**Android (Termux)**

```bash
{ command -v curl >/dev/null 2>&1 || pkg install -y curl; } && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

نصب‌کننده ابزار `bgscan-builder` را دانلود می‌کند که نسخهٔ مناسب پلتفرم شما را دریافت کرده، چک‌سام آن را بررسی می‌کند و bgscan را داخل پوشهٔ `bgscan/` نصب می‌کند. اگر از قبل نسخه‌ای نصب شده باشد، از شما می‌پرسد که چگونه ادامه دهید: به‌روزرسانی در همان محل (که پوشه‌های `ips/`، `assets/` و `settings/` شما را حفظ می‌کند)، نصب تمیز، یا پشتیبان‌گیری از نسخهٔ قبلی در یک پوشه با مهر زمانی به نام `bgscan_bck_*`.

## نصب دستی

1. فایل ZIP مناسب سیستم‌عامل خود را از [صفحهٔ Releases](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.
2. فایل ZIP را باز کنید.
3. **برنامه را اجرا کنید:**
   - **Linux، macOS و Termux:** ترمینال را باز کنید، وارد پوشهٔ برنامه شوید و `./bgscan` را اجرا کنید.
   - **Windows:** روی `bgscan.exe` دو بار کلیک کنید یا در PowerShell دستور `./bgscan.exe` را اجرا کنید.

در اولین اجرا، پوشهٔ `settings/` با فایل‌های تنظیمات پیش‌فرض و پوشهٔ `ips/` با فهرست‌های IP آماده ساخته می‌شود.

## ساخت از سورس

> **نکته:** bgscan را نمی‌توان با `go install` نصب کرد چون برنامه به باینری‌های خارجی (Xray، Slipstream) وابسته است. برای ساخت آن باید از ابزار همراه **`bgscan-builder`** استفاده کنید.

### پیش‌نیازها

- Go نسخهٔ 1.27 یا جدیدتر
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

# دریافت وابستگی‌های پلتفرم شما (Xray، Slipstream و فهرست‌های IP پیش‌فرض)
# Linux / macOS
./scripts/install-deps.sh
# Windows (PowerShell)
./scripts/install-deps.ps1

# اجرا در محیط توسعه
go run ./cmd/bgscan/

# یا ساخت خروجی نهایی برای یک پلتفرم خاص
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os macos -arch arm64
./bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

## ارتقا

برای ارتقا، دوباره اسکریپت [نصب سریع](#نصب-سریع) را اجرا کنید. اسکریپت نسخهٔ قبلی را پیدا می‌کند و به شما اجازه می‌دهد آن را جایگزین کنید یا از آن پشتیبان بگیرید.

اگر تنظیمات خودتان را تغییر داده‌اید:

- فایل‌های `settings/*.toml` را به نصب جدید منتقل کنید.
- فهرست‌های IP شخصی خود را از پوشهٔ `ips/` کپی کنید.
- قبل از جایگزین‌کردن فایل‌ها، مطمئن شوید bgscan در حال اجرا نیست.

## نیازمندی‌ها

- **سیستم‌عامل:** Linux، macOS، Windows 10 یا جدیدتر، یا Android 7.0 یا جدیدتر (Termux)
- **ابزارها:** `curl`؛ نصب‌کننده در بیشتر سیستم‌ها وابستگی‌های لازم را خودش آماده می‌کند
- **Windows:** PowerShell نسخهٔ 5.1 یا جدیدتر
- **Termux:** نسخهٔ F-Droid را نصب کنید؛ نسخهٔ Play Store قدیمی است
