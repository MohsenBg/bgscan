---
title: "Slipstream"
weight: 3
---

# تنظیمات Slipstream

Slipstream به‌صورت یک پروسهٔ خارجی (باینری `slipstream-client`) اجرا می‌شود، نه تونل درون خود برنامه. ارتباط از طریق پورت محلی SOCKS5ای که برای هر پروب گرفته می‌شود برقرار می‌گردد.

## تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `Domain` | (لازم) | دامین معتبر | Zone سرویس‌دهی‌شده توسط سرور Slipstream شما |
| `ResolverPort` | `53` | بزرگ‌تر از 0 | پورت Resolver |
| `CertPath` | `""` | مسیر | Certificate TLS اختیاری (با `--cert` به باینری پاس داده می‌شود) |

## پروکسی و احراز هویت

همان فیلدهای DNSTT است؛ جدول [پروکسی و احراز هویت](../dnstt/#پروکسی-و-احراز-هویت) در صفحهٔ DNSTT را ببینید.

## محل باینری

bgscan باینری `slipstream-client` را در این مسیرها جست‌وجو می‌کند:

1. `<bgscan-root>/assets/slipstream-client/slipstream-client`
2. `<bgscan-root>/assets/slipstream/slipstream-client/slipstream-client`
3. `<bgscan-root>/slipstream-client/slipstream-client`
4. `<bgscan-root>/slipstream-client`
5. `PATH` سیستم

## نمونهٔ Config

```toml
Domain = "example.com"
ResolverPort = 53
CertPath = ""
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

## موضوعات مرتبط

- [تونل DNS](../) — نمای کلی پروتکل‌ها و نحوهٔ اسکن
- [DNSTT](../dnstt/) — فیلدهای Proxy و احراز هویت
