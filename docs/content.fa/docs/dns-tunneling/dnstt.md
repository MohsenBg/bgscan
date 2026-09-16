---
title: "DNSTT"
weight: 1
---

# تنظیمات DNSTT

DNSTT از کتابخانهٔ vaydns با قاب‌بندی سازگار با DNSTT استفاده می‌کند (DNS-over-TXT با قاب‌بندی قدیمی). پشتهٔ تونل این است: کانال بسته‌های DNS روی انتقال KCP با رمزنگاری Noise و چندگانه‌سازی smux.

## تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `Domain` | (لازم) | دامین معتبر | Zone واگذارشده به سرور DNSTT شما |
| `PubKey` | (لازم) | ۶۴ کاراکتر hex | کلید عمومی سرور |
| `ResolverType` | `"udp"` | `udp`، `tcp`، `dot` | روش انتقال به Resolver |
| `ResolverPort` | `53` | بزرگ‌تر از 0 | پورت Resolver |
| `Fingerprint` | `"Chrome"` | جدول پایین | اثر انگشت uTLS ClientHello |
| `RPS` | `0` | 0-500 | محدودیت نرخ (پرس‌وجو در ثانیه)؛ 0 یعنی بدون محدودیت |

## پروکسی و احراز هویت

| فیلد | پیش‌فرض | توضیحات |
| --- | --- | --- |
| `ProxyType` | `"socks"` | نوع Proxy: `socks` یا `ssh` |
| `ProxyPort` | `1080` | پورت Proxy |
| `AuthMethod` | `"none"` | احراز هویت: `none`، `password` یا `key` |
| `Username` | `""` | برای احراز هویت password/key لازم است |
| `Password` | `""` | برای احراز هویت password لازم است |
| `PrivateKey` | `""` | کلید خصوصی SSH با فرمت PEM (برای احراز هویت key لازم است) |
| `KnownHostsFile` | `""` | مسیر فایل known_hosts مربوط به SSH (اختیاری، فقط برای احراز هویت key) |

**قواعد Proxy:**

- Proxy SSH به احراز هویت نیاز دارد (password یا key).
- Proxy SOCKS احراز هویت key ندارد.

## TLS Fingerprint

فیلد `Fingerprint` یک پروفایل uTLS ClientHello انتخاب می‌کند که در TLS Handshake تقلید شود. برچسب‌ها به بزرگی و کوچکی حروف حساس نیستند.

| دسته | برچسب‌ها |
| --- | --- |
| Chrome | `Chrome`, `Chrome_58`, `Chrome_62`, `Chrome_70`, `Chrome_72`, `Chrome_83`, `Chrome_87`, `Chrome_96`, `Chrome_100`, `Chrome_102`, `Chrome_120` |
| Firefox | `Firefox`, `Firefox_55`, `Firefox_56`, `Firefox_63`, `Firefox_65`, `Firefox_99`, `Firefox_102`, `Firefox_105`, `Firefox_120` |
| iOS | `iOS`, `iOS_11_1`, `iOS_12_1`, `iOS_13`, `iOS_14` |
| Other | `random` |

## نمونهٔ Config

```toml
Domain = "example.com"
PubKey = "0000000000000000000000000000000000000000000000000000000000000000"
ResolverType = "udp"
ResolverPort = 53
Fingerprint = "Chrome"
RPS = 0
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

## موضوعات مرتبط

- [تونل DNS](../) — نمای کلی پروتکل‌ها و نحوهٔ اسکن
- [VayDNS](../vaydns/) — پروتکل بومی vaydns با همان گزینه‌های Proxy
