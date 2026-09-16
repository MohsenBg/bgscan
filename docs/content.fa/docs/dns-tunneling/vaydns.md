---
title: "VayDNS"
weight: 2
---

# تنظیمات VayDNS

VayDNS پروتکل بومی vaydns است. همان پشتهٔ تونل DNSTT را دارد (کانال بسته‌های DNS روی انتقال KCP با رمزنگاری Noise و چندگانه‌سازی smux) اما با قاب‌بندی بومی و پارامترهای تنظیم اضافه برای ساختار QNAME، MTU و نوع Record.

## تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `Domain` | (لازم) | دامین معتبر | Zone واگذارشده به سرور VayDNS شما |
| `PubKey` | (لازم) | ۶۴ کاراکتر hex | کلید عمومی سرور |
| `ResolverType` | `"udp"` | `udp`، `tcp`، `dot` | روش انتقال به Resolver |
| `ResolverPort` | `53` | بزرگ‌تر از 0 | پورت Resolver |
| `Fingerprint` | `"Chrome"` | جدول DNSTT | اثر انگشت uTLS ClientHello |
| `RecordType` | `"TXT"` | پایین را ببینید | نوع Record برای پرس‌وجوهای تونل |
| `RPS` | `0` | 0-500 | محدودیت نرخ (پرس‌وجو در ثانیه)؛ 0 یعنی بدون محدودیت |

## تنظیمات پیشرفته

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `ClientIDSize` | `2` | 1-8 | اندازهٔ Client ID بر حسب بایت |
| `MaxQnameLen` | `101` | 0-253 | بیشترین طول QNAME در بسته‌های DNS (0 = خودکار) |
| `MaxNumLabels` | `0` | 0-4 | بیشترین تعداد Label در QNAME (0 = خودکار) |
| `MTU` | `0` | 0-1452 | بیشترین واحد انتقال (0 = خودکار) |

## نوع Recordهای DNS

مقادیر پشتیبانی‌شدهٔ `RecordType`: `A`، `AAAA`، `CNAME`، `NS`، `MX`، `TXT`، `SRV`، `NULL`، `CAA`.

نوع Record تعیین می‌کند پرس‌وجوهای تونل با کدام نوع Record انجام شوند. `TXT` پیش‌فرض است و بیشترین پشتیبانی را دارد.

## پروکسی و احراز هویت

همان فیلدهای DNSTT است؛ جدول [پروکسی و احراز هویت](../dnstt/#پروکسی-و-احراز-هویت) در صفحهٔ DNSTT را ببینید.

## نمونهٔ Config

```toml
Domain = "example.com"
PubKey = "0000000000000000000000000000000000000000000000000000000000000000"
ResolverType = "udp"
ResolverPort = 53
Fingerprint = "Chrome"
RecordType = "TXT"
ClientIDSize = 2
MaxQnameLen = 101
MaxNumLabels = 0
MTU = 0
RPS = 0
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

## موضوعات مرتبط

- [تونل DNS](../) — نمای کلی پروتکل‌ها و نحوهٔ اسکن
- [DNSTT](../dnstt/) — قاب‌بندی سازگار با DNSTT و Fingerprintهای TLS
