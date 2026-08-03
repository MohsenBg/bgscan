---
title: "مشارکت"
weight: 6
---

# مشارکت در bgscan

رفع باگ، بهترکردن مستندات و قابلیت جدید، همگی خوش‌آمدند.

## قبل از شروع

- اول Issueهای باز را نگاه کنید؛ شاید موضوع قبلاً مطرح شده باشد.
- برای تغییر بزرگ یا قابلیت جدید، اول یک Issue باز کنید تا دربارهٔ مسیر کار هماهنگ شویم.
- هر PR را روی یک موضوع نگه دارید.

## Branch

روی `main` مستقیم Commit نکنید:

```bash
git checkout main
git pull origin main
git checkout -b feature/my-feature
```

چند نمونهٔ خوب:

```text
feature/add-xray-parser
fix/windows-build
docs/update-readme
refactor/scanner
test/unit-tests
```

## کد و Commit

کد را ساده، خوانا و نزدیک به سبک فعلی پروژه نگه دارید. تغییرات فرمت نامرتبط ندهید و Comment را فقط وقتی بگذارید که واقعاً چیزی را روشن می‌کند. Commitهای کوچک و متمرکز بهتر از یک Commit بزرگ‌اند.

برای پیام Commit از این سبک استفاده کنید:

```text
feat: add HTTP/3 support
fix: resolve race condition in scanner
docs: improve installation guide
refactor: simplify writer pipeline
test: add unit tests for parser
```

## Pull Request

قبل از فرستادن PR مطمئن شوید Build و تست‌ها پاس می‌شوند. اگر قابلیت جدید اضافه کرده‌اید، تست و مستنداتش را هم اضافه کنید.

داخل PR کوتاه بنویسید چه چیزی عوض شده و چرا. برای تغییر UI Screenshot بگذارید و Issue مرتبط را لینک کنید.

## باگ و قابلیت جدید

برای باگ، OS، معماری، نسخهٔ bgscan، تنظیمات، روش تکرار، چیزی که انتظار داشتید، چیزی که دیدید و در صورت امکان Logها را بفرستید.

برای قابلیت جدید، مسئله‌ای که حل می‌کند، راه‌حل پیشنهادی و گزینه‌های دیگر را توضیح دهید.

## رفتار

بحث فنی و مخالفت محترمانه خوب است. توهین، آزار یا بی‌احترامی جایی در پروژه ندارد.
