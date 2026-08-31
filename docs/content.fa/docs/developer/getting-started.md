---
title: "شروع کار"
weight: 1
---

# شروع کار

این راهنما برای زمانی است که می‌خواهید bgscan را روی سیستم خودتان اجرا و تغییر دهید.

> قبل از شروع، [راهنمای مشارکت](../contributing/) را ببینید. آن‌جا گفته‌ایم Branchها و PRها را چطور نگه می‌داریم.

## چیزهایی که لازم دارید

- [Go](https://go.dev/) نسخهٔ 1.27 یا جدیدتر
- Git
- اگر برای Android Build می‌گیرید: [Android NDK](https://developer.android.com/ndk)

## سورس را بگیرید

```bash
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan
```

برای تغییر خودتان یک Branch بسازید:

```bash
git checkout -b feature/my-change
```

## وابستگی‌ها را آماده کنید

bgscan برای گرفتن وابستگی‌های مخصوص هر سیستم، از `bgscan-builder` استفاده می‌کند.

**Linux / macOS**

```bash
./scripts/install-deps.sh
```

**Windows**

```powershell
./scripts/install-deps.ps1
```

این اسکریپت `bgscan-builder` را به ریشهٔ پروژه می‌آورد و با `setup-dev` وابستگی مناسب OS و معماری فعلی را دانلود می‌کند.

## اجرا

```bash
go mod tidy
go run ./cmd/bgscan/
```

برنامه اول [بررسی‌های اولیه](../core/) را داخل TUI نشان می‌دهد. اگر همه‌چیز درست بود، `Enter` بزنید.

## ساخت Release

اگر هنوز `bgscan-builder` را ندارید، آن را نصب کنید:

```bash
# Linux / macOS
./scripts/install-builder.sh

# Windows
./scripts/install-builder.ps1
```

بعد Release بسازید:

```bash
bgscan-builder release -os linux -arch amd64
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
bgscan-builder release -os all -arch all -dest ./dist
```

برای Android باید مسیر NDK را با `-ndk-dir` بدهید.

| Flag | کار |
| --- | --- |
| `-arch` | معماری هدف: `amd64`، `arm64`، `arm32`، `amd32` یا `all` |
| `-dest` | پوشهٔ خروجی؛ پیش‌فرض `./dist` |
| `-dir` | پوشهٔ نصب bgscan برای install و update؛ پیش‌فرض `bgscan` |
| `-ndk-dir` | مسیر Android NDK |
| `-os` | OS هدف: `linux`، `windows`، `macos`، `android` یا `all` |
| `-project-dir` | مسیر پروژه برای setup-dev |
| `-verbose` | نمایش مرحله‌به‌مرحلهٔ کارها |
| `-version` | تگ Release برای install و update؛ در release نسخهٔ داخل باینری |
| `-xray-version` | نسخهٔ Xray؛ پیش‌فرض `v26.3.27` |
