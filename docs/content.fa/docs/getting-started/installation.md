---
title: "نصب و راه‌اندازی"
weight: 3
bookFlatSection: false
---

# نصب

bgscan روی Linux، macOS، Windows و Android از طریق Termux اجرا می‌شه. روش نصب متناسب با سیستم‌عامل خودتون رو انتخاب کنید.

## نصب سریع

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

**Windows با PowerShell**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**Android با Termux**

```bash
{ command -v curl >/dev/null 2>&1 || pkg install -y curl; } && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

اسکریپت‌های نصب از `bgscan-installer` استفاده می‌کنن. این ابزار یه باینری مستقل و کوچیکه که با Rust نوشته شده و مخزن جداگانه‌ای داره که در [bgscan-installer](https://github.com/MohsenBg/bgscan-installer) قرار گرفته.

`bgscan-installer` نسخه مناسب سیستم شما رو از بین نسخه‌های منتشرشده پیدا می‌کنه، چک‌سام SHA256 اون رو با مقدار ثبت‌شده در `checksum.txt` تطبیق می‌ده و bgscan رو داخل پوشه `bgscan/` نصب می‌کنه. اگه باینری `bgscan` از قبل توی پوشه فعلی وجود داشته باشه، نصب همون‌جا انجام می‌شه.

اگه نصب قبلی پیدا بشه، می‌تونید یکی از این گزینه‌ها رو انتخاب کنید:

- **به‌روزرسانی:** باینری جایگزین می‌شه و فایل‌های جدید اضافه می‌شن، اما پوشه‌های `ips`، `assets` و `settings` دست‌نخورده می‌مونن.
- **نصب تمیز:** نصب قبلی با یه نصب تازه جایگزین می‌شه.
- **پشتیبان‌گیری:** از نصب فعلی توی پوشه‌ای با مهر زمانی به اسم `bgscan_<timestamp>` نسخه پشتیبان گرفته می‌شه.
- **انصراف:** عملیات نصب متوقف می‌شه.

اگه نسخه مشخصی مدنظرتونه و نمی‌خواید آخرین نسخه نصب بشه، از گزینه `--version` استفاده کنید:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh -s -- --version v2.12.0
```

## نصب دستی

1. فایل ZIP مناسب سیستم‌عامل و معماری دستگاهتون رو از [صفحه Releases](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.
2. فایل ZIP رو از حالت فشرده خارج کنید.
3. برنامه رو اجرا کنید:
   - **Linux/macOS/Termux:** ترمینال رو باز کنید، وارد پوشه برنامه بشید و دستور `./bgscan` رو اجرا کنید.
   - **Windows:** روی `bgscan.exe` دوبار کلیک کنید یا در PowerShell دستور `.\bgscan.exe` رو اجرا کنید.

در اولین اجرا، پوشه `settings/` با فایل‌های تنظیمات پیش‌فرض و پوشه `ips/` با فهرست‌های IP آماده ساخته می‌شن.

## ساخت از سورس

> **نکته:** نصب bgscan با `go install` امکان‌پذیر نیست، چون برنامه به کتابخانه Slipstream مخصوص هر پلتفرم نیاز داره. برای ساخت برنامه باید از ابزار همراهش، **`bgscan-builder`**، استفاده کنید. این ابزار فقط برای ساخت برنامه است. نصب و به‌روزرسانی نسخه‌های منتشرشده با `bgscan-installer` انجام می‌شه.

> Xray به باینری جداگانه نیاز نداره و از طریق یه کتابخانه داخلی، داخل همون پروسه اجرا می‌شه.

#### پیش‌نیازها

- Go 1.27 یا بالاتر
- Git

#### کلون و ساخت

```bash
# Clone the repository
git clone --recurse-submodules https://github.com/MohsenBg/bgscan.git
cd bgscan

# Install the builder tool
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# Fetch the Slipstream client library and bundled IP lists for your platform
# Linux/macOS
./scripts/install-deps.sh
# Windows (PowerShell)
./scripts/install-deps.ps1

# Run in development
go run ./cmd/bgscan/

# Or build release artifacts targeting a specific platform (see `bgscan-builder release -h` for all flags)
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os macos -arch arm64
./bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

## ارتقا

برای به‌روزرسانی bgscan، کافیه دوباره اسکریپت [نصب سریع](#نصب-سریع) رو اجرا کنید. نصب‌کننده، نصب قبلی رو پیدا می‌کنه و گزینه‌های به‌روزرسانی، نصب تمیز، پشتیبان‌گیری یا انصراف رو در اختیارتون می‌ذاره.

در حالت به‌روزرسانی، پوشه‌های `settings`، `ips` و `assets` حفظ می‌شن.

اگه نصب تمیز رو انتخاب کردید، بهتره قبلش از اطلاعات موردنیاز نسخه قبلی پشتیبان بگیرید:

- فایل‌های `settings/*.toml` که تغییر دادید، به نصب جدید منتقل کنید.
- فهرست‌های IP سفارشی رو از پوشه `ips/` بردارید.
- قبل از جایگزین کردن فایل‌های برنامه، bgscan رو ببندید.

## پیش‌نیازهای سیستم

- **سیستم‌عامل:** Linux، macOS، Windows 10 یا بالاتر، یا Android 7.0 یا بالاتر با Termux
- **ابزارها:** `curl`. نصب‌کننده در بیشتر سیستم‌ها ابزارهای موردنیاز رو خودش نصب می‌کنه.
- **Windows:** PowerShell 5.1 یا بالاتر
- **Termux:** نسخه F-Droid رو نصب کنید. نسخه Play Store قدیمیه.
