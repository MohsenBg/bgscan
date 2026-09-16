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

اسکریپت‌های نصب از `bgscan-installer` استفاده می‌کنند؛ یک باینری کوچک و مستقل که با Rust نوشته شده و پروژهٔ جداگانهٔ خودش را دارد ([bgscan-installer](https://github.com/MohsenBg/bgscan-installer)). این ابزار نسخهٔ مناسب پلتفرم شما را پیدا می‌کند، چک‌سام SHA-256 آن را با `checksum.txt` بررسی می‌کند و bgscan را داخل پوشهٔ `bgscan/` نصب می‌کند (یا داخل همان پوشهٔ فعلی، اگر باینری `bgscan` از قبل آن‌جا باشد). اگر از قبل نسخه‌ای نصب شده باشد، از شما می‌پرسد که چگونه ادامه دهید: به‌روزرسانی در همان محل (باینری را عوض می‌کند و فایل‌های جدید را اضافه می‌کند، ولی `ips`، `assets` و `settings` شما دست نمی‌خورد)، نصب تمیز، پشتیبان‌گیری از نصب قبلی در پوشه‌ای با مهر زمانی به نام `bgscan_<timestamp>`، یا انصراف.

برای نصب یک نسخهٔ مشخص به‌جای آخرین نسخه، `--version` بدهید:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh -s -- --version v2.11.0
```

## نصب دستی

1. فایل ZIP مناسب سیستم‌عامل خود را از [صفحهٔ Releases](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.
2. فایل ZIP را باز کنید.
3. **برنامه را اجرا کنید:**
   - **Linux، macOS و Termux:** ترمینال را باز کنید، وارد پوشهٔ برنامه شوید و `./bgscan` را اجرا کنید.
   - **Windows:** روی `bgscan.exe` دو بار کلیک کنید یا در PowerShell دستور `./bgscan.exe` را اجرا کنید.

در اولین اجرا، پوشهٔ `settings/` با فایل‌های تنظیمات پیش‌فرض و پوشهٔ `ips/` با فهرست‌های IP آماده ساخته می‌شود.

## ساخت از سورس

> **نکته:** bgscan را نمی‌توان با `go install` نصب کرد چون به باینری Slipstream مخصوص پلتفرم نیاز دارد. برای ساخت آن باید از ابزار همراه **`bgscan-builder`** استفاده کنید. این ابزار فقط برای ساخت است؛ نصب و به‌روزرسانی نسخه‌ها با ابزار جداگانهٔ `bgscan-installer` انجام می‌شود که بالاتر توضیح داده شد. (خود Xray باینری نمی‌خواهد و داخل برنامه اجرا می‌شود.)

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

# گرفتن باینری جانبی Slipstream و فهرست‌های IP مخصوص پلتفرم شما
# Linux / macOS
./scripts/install-deps.sh
# Windows (PowerShell)
./scripts/install-deps.ps1

# اجرا در محیط توسعه
go run ./cmd/bgscan/

# یا ساخت خروجی نهایی برای یک پلتفرم خاص (همهٔ فلگ‌ها را با `bgscan-builder release -h` ببینید)
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os macos -arch arm64
./bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

## ارتقا

برای ارتقا، دوباره اسکریپت [نصب سریع](#نصب-سریع) را اجرا کنید. اسکریپت نصب قبلی را پیدا می‌کند و به‌روزرسانی، نصب تمیز، پشتیبان‌گیری یا انصراف را پیشنهاد می‌دهد.

گزینهٔ به‌روزرسانی `settings`، `ips` و `assets` شما را نگه می‌دارد. اگر نصب تمیز را انتخاب کردید، اول پشتیبان بگیرید:

- فایل‌های `settings/*.toml` را به نصب جدید منتقل کنید.
- فهرست‌های IP شخصی خود را از پوشهٔ `ips/` کپی کنید.
- قبل از جایگزین‌کردن فایل‌ها، مطمئن شوید bgscan در حال اجرا نیست.

## نیازمندی‌ها

- **سیستم‌عامل:** Linux، macOS، Windows 10 یا جدیدتر، یا Android 7.0 یا جدیدتر (Termux)
- **ابزارها:** `curl`؛ نصب‌کننده در بیشتر سیستم‌ها وابستگی‌های لازم را خودش آماده می‌کند
- **Windows:** PowerShell نسخهٔ 5.1 یا جدیدتر
- **Termux:** نسخهٔ F-Droid را نصب کنید؛ نسخهٔ Play Store قدیمی است
