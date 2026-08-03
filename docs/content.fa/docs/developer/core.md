---
title: "هسته"
weight: 3
---

# هسته

هر چیزی که TUI نیست داخل `internal/core` قرار دارد: تنظیمات، موتور اسکن، Probeها، IP list، نتیجه‌ها، DNS، Xray و مدیریت Process.

## پکیج‌ها

| پکیج | کار |
|---|---|
| `config` | ساختار تنظیمات، پیش‌فرض‌ها و خواندن/نوشتن TOML |
| `config/validate` | بررسی و اصلاح تنظیمات نامعتبر |
| `scanner` | Scanner، Stage builder و ساخت Pipeline |
| `scanner/engine` | اجرای اسکن و مدیریت Pause |
| `scanner/probe` | قرارداد Probe و پیاده‌سازی پروتکل‌ها |
| `scanner/portmgr` | دادن پورت محلی به Clientهای Tunnel |
| `result` | Result، Schema، Registry، Writer و CSV merge |
| `iplist` | ورود، خواندن و Stream فهرست‌های IP |
| `dns` | Query DNS، Transport، DNSTT، Slipstream و SOCKS5 |
| `xray` و `speedtest` | اجرای Xray و اندازه‌گیری اتصال و سرعت |

## Probe

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip netip.Addr) (result.Result, error)
    Schema() result.ResultSchema
    Close() error
}
```

`Init` قبل از شروع Workerها اجرا می‌شود و `Close` در پایان. `Schema` به bgscan می‌گوید نتیجهٔ این Probe چه ستون‌هایی دارد و کجا ذخیره شود.

## اسکن و Pipeline

Scanner از Stageها ساخته می‌شود. هر Stage یک Probe، Writer، تعداد Worker و Hookهای UI دارد. برای یک Stage، `RunScan` اجرا می‌شود. برای چند Stage، `RunScanWithChain` یکی از حالت‌های `sequential`، `streaming` یا `batch` را اجرا می‌کند.

## نتیجه‌ها

هر Result کلید، Record CSV و Score خودش را دارد. Writer نتیجه‌های جدید را جمع می‌کند، بر اساس Score مرتب می‌کند و با فایل قبلی ادغام می‌کند. IP تکراری با نتیجهٔ جدید جایگزین می‌شود؛ کل فایل هم‌زمان در حافظه بارگذاری نمی‌شود.

برای Probe جدید باید Result، Schema، Probe و Stage builder بسازید و Schema را در `core.Init()` ثبت کنید. بعد Registry، Writer و TUI خودکار می‌توانند فایل آن را بخوانند و نشان دهند.

## تنظیمات

Config Singleton وجود ندارد. `main` با `config.NewStore()` فایل‌ها را می‌خواند و همان `ScannerConfig` را به UI و Scanner می‌دهد. ذخیرهٔ هر بخش جداست: `SaveGeneral`، `SaveWriter`، `SaveICMP`، `SaveTCP`، `SaveHTTP`، `SaveXray` و `SaveDNS`.
