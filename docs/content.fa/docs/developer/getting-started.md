---
title: "شروع کار"
weight: 1
bookFlatSection: true
bookCollapseSection: true
---

# شروع کار برای توسعه‌دهنده

این راهنما مراحل آماده‌کردن محیط توسعه، ساخت **bgscan** و اجرای محلی پروژه رو توضیح می‌ده.

> **قبل از شروع:** راهنمای [مشارکت](./contributing/) رو بخونید تا با قواعد Branch، نوشتن Commit و روند Pull Request آشنا بشید.

---

## پیش‌نیازها

- [Go](https://go.dev/) نسخهٔ 1.27 یا بالاتر. نسخهٔ دقیق موردنیاز در `go.mod` مشخص شده.
- Git
- برای Build اندروید: [Android NDK](https://developer.android.com/ndk)

---

## ۱. دریافت مخزن

```bash
git clone --recurse-submodules https://github.com/MohsenBg/bgscan.git
cd bgscan
```

استفاده از `--recurse-submodules` ضروریه. هستهٔ Xray و کدهای بالادستی دیگه داخل Submoduleهای پوشهٔ `third_party/` قرار دارن.

## ۲. ساخت Branch

```bash
git checkout -b feature/my-change
```

برای آشنایی با قواعد نام‌گذاری Branch، به راهنمای [مشارکت](./contributing/) مراجعه کنید.

## ۳. نصب وابستگی‌ها

پروژه از ابزار همراهی به اسم **`bgscan-builder`** استفاده می‌کنه. این ابزار وابستگی‌های موردنیاز، از جمله کتابخانهٔ کلاینت Slipstream و فهرست‌های IP مخصوص هر پلتفرم رو آماده می‌کنه.

اسکریپت نصب، `bgscan-builder` رو دانلود می‌کنه و در ریشهٔ پروژه قرار می‌ده. بعد هم با استفاده از اون، کتابخانهٔ مناسب سیستم‌عامل و معماری شما رو دریافت می‌کنه.

برای Xray نیازی به باینری جداگانه نیست. `xray-core` به‌صورت داخلی و در همون پروسه اجرا می‌شه.

**Linux / macOS**

```bash
./scripts/install-deps.sh
```

**Windows**

```powershell
./scripts/install-deps.ps1
```

این اسکریپت دو کار انجام می‌ده:

1. `bgscan-builder` رو در ریشهٔ پروژه دانلود می‌کنه.
2. دستور `bgscan-builder setup-dev --project-dir <project-root>` رو اجرا می‌کنه تا کتابخانهٔ کلاینت Slipstream متناسب با سیستم‌عامل و معماری شما دانلود و در مسیر درست قرار بگیره.

## ۴. ساخت و اجرای پروژه

بعد از نصب وابستگی‌ها، دستورهای زیر رو اجرا کنید:

```bash
go mod tidy
go run ./cmd/bgscan/
```

ابتدا [بررسی‌های اولیه](./core/#startup) اجرا می‌شن. اگر بررسی‌ها بدون مشکل تموم بشن، با زدن Enter وارد TUI می‌شید.

---

## ساخت نسخهٔ Release

برای ساخت خروجی Release هم به `bgscan-builder` نیاز دارید. اگر در مرحلهٔ نصب وابستگی‌ها نصبش نکردید، از اسکریپت مربوط به سیستم‌عامل خودتون استفاده کنید.

**Linux / macOS**

```bash
./scripts/install-builder.sh
```

**Windows**

```powershell
./scripts/install-builder.ps1
```

#### دستورهای Build

```bash
bgscan-builder release -os linux -arch amd64
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
bgscan-builder release -os all -arch all -dest ./dist
```

#### ساخت نسخهٔ اندروید

برای Build اندروید باید Android NDK نصب باشه. مسیر نصب اون رو با فلگ `-ndk-dir` مشخص کنید:

```bash
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

---

## مرجع `bgscan-builder`

`bgscan-builder` یک ابزار CLI کوچک نوشته‌شده با Go برای ساخت bgscan هست و دو Subcommand داره:

- **`setup-dev`**: وابستگی‌های موردنیاز سیستم‌عامل و معماری فعلی رو آماده می‌کنه تا بتونید پروژه رو با `go run ./cmd/bgscan/` اجرا کنید.
- **`release`**: کتابخانهٔ کلاینت Slipstream رو آماده می‌کنه و باینری Release رو برای سیستم‌عامل و معماری مشخص‌شده می‌سازه.

نصب و به‌روزرسانی نسخه‌های منتشرشده وظیفهٔ این ابزار نیست. این کار رو ابزار جداگانهٔ [bgscan-installer](https://github.com/MohsenBg/bgscan-installer) انجام می‌ده. برای جزئیات، راهنمای [نصب](../getting-started/installation/) رو ببینید.

**فلگ‌ها**

| فلگ | Subcommand | توضیح |
|---|---|---|
| `-arch string` | `release` | معماری هدف: `amd64`، `arm64`، `arm32`، `amd32` یا `all` |
| `-dest string` | `release` | مسیر پوشهٔ خروجی Release، مقدار پیش‌فرض `./dist` |
| `-ndk-dir string` | `release` برای اندروید | مسیر ریشهٔ Android NDK |
| `-os string` | `release` | سیستم‌عامل هدف: `linux`، `windows`، `macos`، `android` یا `all` |
| `-project-dir string` | `setup-dev`, `release` | مسیر پروژهٔ bgscan |
| `-verbose` | همه | نمایش جزئیات هر مرحله هنگام اجرا |
| `-version string` | `release` | نسخه‌ای که داخل باینری ساخته‌شده ثبت می‌شه |
