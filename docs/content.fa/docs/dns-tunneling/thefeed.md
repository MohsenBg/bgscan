---
title: "TheFeed"
weight: 6
---

# تنظیمات TheFeed

TheFeed داخل خود BGScan و از طریق یک کتابخانه کلاینت داخلی اجرا می‌شه. کوئری‌ها و پاسخ‌ها با کلیدهای ۲۵۶ بیتی AES رمزنگاری می‌شن که از یک Passphrase مشترک ساخته می‌شن.

این پروتکل تنظیمات پروکسی نداره. مقدار Timeout پروب هم از تنظیمات کلی تونل DNS، یعنی `dns_settings.toml` و گزینه `dns_tunneling.timeout`، گرفته می‌شه. بنابراین، لازم نیست اون رو در فایل Config مربوط به TheFeed تنظیم کنید.

## تنظیمات اتصال

| فیلد | پیش‌فرض | توضیح |
| --- | --- | --- |
| `domain` | الزامی | زیردامنه‌ای که به کوئری‌های Feed پاسخ می‌ده، مثلاً `t.example.com` |
| `passphrase` | الزامی | همون Passphrase سرور که برای ساخت کلیدهای ۲۵۶ بیتی AES استفاده می‌شه |
| `resolver_port` | `53` | پورت ریزالور، سرورهای TheFeed معمولاً از پورت `5300` استفاده می‌کنن |
| `resolver_type` | `'udp'` | ترنسپورت ارتباط با ریزالور: `udp`، `tcp` یا `dot` |
| `query_mode` | `'single'` | روش کدگذاری کوئری: `single` برای یک برچسب Base32 یا `double` برای چند برچسب Hex |

## نمونه Config

```toml id="q6v3nw"
domain = 't.example.com'
passphrase = 'your-passphrase'
resolver_port = 53
resolver_type = 'udp'
query_mode = 'single'
```

## موضوعات مرتبط

- [تونل DNS](../): آشنایی با پروتکل‌ها و روند اسکن
- [تنظیمات DNS](../settings/dns.md): تنظیم Timeout تونل و پیش‌اسکن ریزالور

