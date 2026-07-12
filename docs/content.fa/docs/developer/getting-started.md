---
title: "شروع به کار"
weight: 1
bookFlatSection: true
bookCollapseSection: true
---

# شروع به کار - توسعه‌دهندگان (Getting Started)

این راهنما نحوه آماده‌سازی محیط توسعه، کامپایل و اجرای محلی **bgscan** را برای مقاصد توسعه و برنامه‌نویسی توضیح می‌دهد.

> **ابتدا بخوانید:** [راهنمای مشارکت](../contributing/) — سیستم برانچ‌بندی، قوانین ثبت کامیت و فرآیند ثبت درخواست‌های Pull Request (PR).

---

## پیش‌نیازها (Prerequisites)

- زبان برنامه‌نویسی [Go](https://go.dev/) نسخه 1.26.3 به بالا (نسخه دقیق را در فایل `go.mod` بررسی کنید)
- سیستم کنترل نسخه Git
- برای کامپایل نسخه اندروید: ابزار [Android NDK](https://developer.android.com/ndk)

---

## ۱. کلون کردن مخزن (Repository)

```bash
git clone [https://github.com/MohsenBg/bgscan.git](https://github.com/MohsenBg/bgscan.git)
cd bgscan

```

## ۲. ایجاد یک شاخه یا برانچ (Branch)

```bash
git checkout -b feature/my-change

```

برای آشنایی با استانداردهای نام‌گذاری شاخه‌ها، بخش [راهنمای مشارکت](../contributing/) را مطالعه کنید.

## ۳. نصب وابستگی‌ها (Dependencies)

پروژه bgscan از یک ابزار کمکی به نام **`bgscan-builder`** برای دریافت و کامپایل وابستگی‌های پروژه استفاده می‌کند. اسکریپت‌های نصب، این ابزار را برای شما دانلود کرده، در ریشه پروژه قرار می‌دهند و از آن برای دریافت نسخه صحیح وابستگی‌های متناسب با سیستم‌عامل و معماری پردازنده شما استفاده می‌کنند.

**لینوکس / مک (Linux / macOS)**

```bash
./scripts/install-deps.sh

```

**ویندوز (Windows)**

```powershell
./scripts/install-deps.ps1

```

این اسکریپت اقدامات زیر را انجام می‌دهد:

1. ابزار `bgscan-builder` را در پوشه ریشه پروژه دانلود می‌کند.
2. دستور `bgscan-builder setup-dev --project-dir <project-root>` را اجرا می‌کند تا وابستگی‌های صحیح سازگار با سیستم‌عامل/معماری شما را دانلود کرده و در دایرکتوری مناسب قرار دهد.

## ۴. کامپایل و اجرا

پس از نصب کامل وابستگی‌ها، دستورات زیر را اجرا کنید:

```bash
go mod tidy
go run ./cmd/bgscan/

```

در ابتدا [بررسی‌های سلامت راه‌اندازی (Startup Health Checks)](../core/%23startup) اجرا می‌شوند. پس از موفقیت‌آمیز بودن آن‌ها، کلید Enter را فشار دهید تا وارد محیط متنی برنامه (TUI) شوید.

---

## ساخت نسخه‌های ریلیز (Building Releases)

برای کامپایل و ساخت آرتفکت‌های نهایی (Release Artifacts)، به ابزار `bgscan-builder` نیاز دارید. اگر در مرحله نصب وابستگی‌ها آن را دریافت نکرده‌اید، با دستورات زیر آن را نصب کنید:

**لینوکس / مک (Linux / macOS)**

```bash
./scripts/install-builder.sh

```

**ویندوز (Windows)**

```powershell
./scripts/install-builder.ps1

```

#### دستورات کامپایل (Build Commands)

```bash
bgscan-builder release -os linux -arch amd64
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
bgscan-builder release -os all -arch all -dest ./dist

```

#### کامپایل برای اندروید (Android)

کامپایل نسخه اندروید به ابزار Android NDK نیاز دارد. مسیر آن را با پرچم `-ndk-dir` به دستور پاس بدهید:

```bash
bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk

```

---

## مرجع ابزار bgscan-builder

این ابزار خروجی‌های نسخه ریلیز را برای یک یا چند ترکیب مختلف از سیستم‌عامل‌ها و معماری‌های پردازنده کامپایل می‌کند.

**پرچم‌ها (Flags):**

| پرچم | توضیحات |
| --- | --- |
| `-arch string` | معماری پردازنده هدف (`amd64` ،`arm64` ،`arm32` ،`amd32` ،`all`) |
| `-dep-version string` | تگ نسخه وابستگی‌ها (مقدار پیش‌فرض: `"v1.0"`) |
| `-dest string` | دایرکتوری خروجی نسخه ریلیز (مقدار پیش‌فرض: `"./dist"`) |
| `-ndk-dir string` | دایرکتوری ریشه Android NDK |
| `-os string` | سیستم‌عامل هدف (`linux` ،`windows` ،`macos` ،`android` ،`all`) |
| `-project-dir string` | مسیر دایرکتوری پروژه bgscan |
| `-xray-version string` | تگ نسخه Xray (مقدار پیش‌فرض: `"v26.3.27"`) |
