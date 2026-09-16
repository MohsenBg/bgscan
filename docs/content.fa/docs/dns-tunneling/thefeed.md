---
title: "TheFeed"
weight: 6
---

# تنظیمات TheFeed

TheFeed با کتابخانهٔ داخلی خودش داخل برنامه اجرا می‌شود. کوئری‌ها و پاسخ‌ها با کلیدهای AES-256 رمزنگاری می‌شوند که از یک Passphrase مشترک ساخته می‌شوند. تنظیمات Proxy ندارد. مهلت زمانی پروب داخل این فایل نیست و از تنظیمات کلی تونل DNS می‌آید (`dns_settings.toml` و `dns_tunneling.timeout`).

## تنظیمات اتصال

| فیلد | پیش‌فرض | توضیحات |
| --- | --- | --- |
| `Domain` | (لازم) | ساب‌دامین DNS که به کوئری‌های فید جواب می‌دهد (مثلاً `t.example.com`) |
| `Passphrase` | (لازم) | همان Passphrase سرور که کلیدهای AES-256 از آن ساخته می‌شوند |
| `ResolverPort` | `53` | پورت Resolver (سرورهای TheFeed معمولاً روی 5300 هستند) |
| `ResolverType` | `'udp'` | روش انتقال: `udp`، `tcp` یا `dot` |
| `QueryMode` | `'single'` | کدگذاری کوئری: `single` (یک لیبل base32) یا `double` (چند لیبل hex) |

## نمونهٔ Config

```toml
domain = 't.example.com'
passphrase = 'your-passphrase'
resolver_port = 53
resolver_type = 'udp'
query_mode = 'single'
```

## موضوعات مرتبط

- [تونل DNS](../) — نمای کلی پروتکل‌ها و نحوهٔ اسکن
- [تنظیمات DNS](../../settings/dns.md) — مهلت کلی تونل و پیش‌اسکن Resolver
