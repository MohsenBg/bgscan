---
title: "معماری"
weight: 2
---

# معماری پروژه

bgscan چند لایهٔ مشخص دارد. `main` اول Schemaهای نتیجه را ثبت می‌کند، تنظیمات را می‌خواند و بررسی‌های اولیه را انجام می‌دهد. بعد TUI را بالا می‌آورد. TUI اسکنر را می‌سازد و اسکنر با چند Stage، Probe و Writer کار می‌کند.

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
| `internal/startup` | بررسی تنظیمات و باینری‌های لازم هنگام شروع |
| `assets` | Xray و Clientهای DNSTT و Slipstream |
| `ips` | فهرست‌های آمادهٔ IP |
| `settings` | فایل‌های TOML در حال استفاده |

## مسیر اجرا

```text
main
 ├─ core.Init()                 Schemaهای نتیجه را ثبت می‌کند
 ├─ Store.Load()                تنظیمات را می‌خواند
 ├─ RunHealthChecks()           config و باینری‌ها را بررسی می‌کند
 └─ TUI
     └─ Run Scan
         ├─ Scanner و Stageها
         ├─ یک Stage: RunScan
         └─ چند Stage: Sequential / Streaming / Batch
             └─ Writer → فایل CSV
```

## چند تصمیم مهم

تنظیمات Global نیستند. `main` یک `ScannerConfig` می‌سازد و همان را از راه `AppState` به TUI و Scanner می‌دهد.

هر Probe مستقل است و فقط باید `Init`، `Run`، `Schema` و `Close` را پیاده کند. `Run` یک `netip.Addr` می‌گیرد، پس IPv4 و IPv6 مسیر جدا ندارند.

هر نوع نتیجه Schema خودش را دارد: پوشهٔ خروجی، ستون‌های CSV و Parser. وقتی Schema در `core.Init()` ثبت شد، Writer و جدول نتیجه بدون تغییر اضافه آن را می‌شناسند.

موتور اسکن کاری به نوع پروتکل ندارد؛ IPها و نتیجه‌ها را بین Stageها جابه‌جا می‌کند. Xray، DNSTT و Slipstream هم اختیاری‌اند؛ نبودشان فقط همان اسکن را غیرفعال می‌کند.
