---
title: "نصب و راه‌اندازی"
weight: 3
bookFlatSection: false
---

# نصب و راه‌اندازی

ابزار bgscan روی سیستم‌عامل‌های لینوکس، مک، ویندوز و اندروید (ترموکس) اجرا می‌شود. روشی را که با محیط کاربری‌تان سازگارتر است انتخاب کنید.

## نصب سریع

**لینوکس / مک**

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash

```

**ویندوز (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex

```

**اندروید (Termux)**

```bash
pkg update -y && pkg install bash curl unzip -y && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash

```

اسکریپت نصب‌کننده، آخرین نسخه منتشر شده را دانلود کرده، آن را در پوشه `bgscan/` استخراج می‌کند و فایل باینری را به حالت قابل اجرا (executable) در می‌آورد. اگر نسخه‌ای از قبل نصب شده باشد، از شما می‌پرسد که آیا تمایل به حذف آن دارید یا می‌خواهید یک نسخه پشتیبان از آن با نام `bgscan_old` بسازید.

## نصب دستی

۱. فایل ZIP مربوط به پلتفرم خود را از [صفحه ریلیزها (Releases)](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.

۲. فایل فشرده را استخراج (Extract) کنید.

۳. **برنامه را اجرا کنید:**

* **لینوکس/مک/ترموکس:** ترمینال را باز کنید، به پوشه برنامه بروید و دستور `./bgscan` را اجرا کنید.
* **ویندوز:** برای اجرا کافیست روی فایل `bgscan.exe` دو بار کلیک کنید، یا دستور `.\bgscan.exe` را در پاورشل بزنید.

در اولین اجرا، پوشه پیش‌فرض `settings/` حاوی فایل‌های پیکربندی و پوشه `ips/` شامل لیست آی‌پی‌های همراه برنامه، به‌طور خودکار ساخته می‌شوند.

## ساخت از روی سورس (Build from Source)

> **نکته:** ابزار bgscan را نمی‌توان از طریق دستور `go install` نصب کرد. برای این کار باید از ابزار `bg-builder` که داخل خود ریپازیتوری قرار دارد استفاده کنید.

#### پیش‌نیازها

* Go نسخه 1.26.3 به بالا
* Git

#### کلون و بیلد کردن

```bash
# کلون کردن ریپازیتوری
git clone [https://github.com/MohsenBg/bgscan.git](https://github.com/MohsenBg/bgscan.git)
cd bgscan

# نصب ابزار بیلدر
# لینوکس/مک
curl -fsSL [https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh](https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh) | bash

# ویندوز (PowerShell)
irm [https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1](https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1) | iex

# ساخت برنامه برای سیستم فعلی شما
./bg-builder

# یا ساخت برای یک هدف (Target) مشخص
./bg-builder --os linux --arch amd64
./bg-builder --os windows --arch amd64
./bg-builder --os darwin --arch arm64
./bg-builder --os android --arch arm64

```

## ارتقا (Upgrade)

برای ارتقا به نسخه‌های جدیدتر، کافیست دوباره اسکریپت [نصب سریع](https://www.google.com/search?q=%23%D9%86%D8%B5%D8%A8-%D8%B3%D8%B1%DB%8C%D8%B9) را اجرا کنید. این اسکریپت نسخه فعلی شما را شناسایی کرده و پیشنهاد جایگزینی یا بک‌آپ‌گیری از آن را به شما می‌دهد.

اگر تنظیمات شخصی‌سازی شده دارید:

* فایل‌های کانفیگ خود را از مسیر `settings/*.toml` به محل نصب جدید کپی کنید.
* لیست آی‌پی‌های کاستوم خود را از پوشه `ips/` منتقل کنید.
* قبل از جایگزینی فایل‌ها، مطمئن شوید که هیچ پروسه در حال اجرایی از bgscan باز نیست.

## نیازمندی‌های سیستم

* **سیستم‌عامل:** لینوکس، مک، ویندوز ۱۰ به بالا، یا اندروید ۷.۰ به بالا (Termux)
* **ابزارها:** ابزارهای `curl`، `unzip` و `bash` (نصب‌کننده در اکثر سیستم‌ها کمبود این وابستگی‌ها را خودش برطرف می‌کند)
* **ویندوز:** پاورشل نسخه 5.1 به بالا
* **ترموکس:** حتماً از F-Droid نصب شده باشد (نسخه گوگل پلی قدیمی است و پشتیبانی نمی‌شود)
