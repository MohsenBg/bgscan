---
title: "معماری"
weight: 2
---

# معماری پروژه

bgscan چند لایهٔ مشخص دارد. `main` اول Theme را راه می‌اندازد و TUI را بالا می‌آورد؛ همین. TUI سه مرحله دارد: Splash، Startup و Workspace. مرحلهٔ Startup لاگرها را آماده می‌کند، Schemaهای نتیجه را ثبت می‌کند، تنظیمات را می‌خواند و باینری‌های لازم را بررسی می‌کند. بعد Workspace اسکنر را می‌سازد و اسکنر با چند Stage، Probe و Writer کار می‌کند.

## مسیرهای مهم

| مسیر | چه چیزی آن‌جاست؟ |
|---|---|
| `cmd/bgscan` | نقطهٔ شروع برنامه |
| `internal/core` | اسکنر، تنظیمات، نتیجه‌ها، DNS، Xray و IP list |
| `internal/core/scanner` | ساخت Stageها و اجرای اسکن |
| `internal/core/scanner/engine` | اجرای تک‌مرحله‌ای، Sequential، Streaming و Batch |
| `internal/core/scanner/probe` | قرارداد Probe و پیاده‌سازی هر پروتکل |
| `internal/core/result` | Schema، Writer، خواندن و ادغام CSV |
| `internal/ui` | TUI، منوها، جدول‌ها، Dialogها و Theme |
| `internal/ui/main/startup` | بررسی تنظیمات و باینری‌های لازم هنگام شروع، داخل TUI |
| `assets` | Xray، باینری Slipstream و فایل‌های Config تونل DNS |
| `ips` | فهرست‌های آمادهٔ IP |
| `settings` | فایل‌های TOML در حال استفاده |

## مسیر اجرا

```text
main
 ├─ theme.Init()                پالت رنگی را تعیین می‌کند
 └─ TUI (BubbleTea)
     ├─ Splash                  انیمیشن لوگو و نسخه
     ├─ Startup                 بررسی‌های اولیه، داخل TUI
     │   ├─ لاگرها + core.Init()   ثبت Schemaهای نتیجه
     │   ├─ Store.Load()        تنظیمات را می‌خواند و Validate می‌کند
     │   └─ باینری‌ها            Xray، Slipstream و Configهای تونل
     └─ Workspace
         └─ Run Scan
             ├─ Scanner و Stageها
             ├─ یک Stage: RunScan
             └─ چند Stage: Sequential / Streaming / Batch
                 └─ Writer → فایل CSV
```

## چند تصمیم مهم

تنظیمات Global نیستند. مرحلهٔ Startup یک `Store` می‌سازد و Config را بارگذاری می‌کند و همان را از راه `AppState` به همهٔ کامپوننت‌ها و Scanner می‌رساند.

Config خراب خودش را ترمیم می‌کند. `NormalizeAll` مقدارهای خارج از محدوده را به پیش‌فرض برمی‌گرداند و اصلاحات را گزارش می‌کند؛ پس یک ویرایش اشتباه برنامه را از کار نمی‌اندازد.

هر Probe مستقل است و فقط باید `Init`، `Run`، `Schema` و `Close` را پیاده کند. `Run` یک `netip.Addr` می‌گیرد، پس IPv4 و IPv6 مسیر جدا ندارند.

هر نوع نتیجه Schema خودش را دارد: پوشهٔ خروجی، ستون‌های CSV و Parser. وقتی Schema در `core.Init()` ثبت شد، Writer و جدول نتیجه بدون تغییر اضافه آن را می‌شناسند.

موتور اسکن کاری به نوع پروتکل ندارد؛ IPها و نتیجه‌ها را بین Stageها جابه‌جا می‌کند. Xray، DNSTT و Slipstream هم اختیاری‌اند؛ نبودشان فقط همان اسکن را غیرفعال می‌کند.
