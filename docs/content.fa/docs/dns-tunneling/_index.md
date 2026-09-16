---
title: "تونل DNS"
weight: 3
bookFlatSection: true
bookCollapseSection: true
---

# تونل DNS

bgscan بررسی می‌کند کدام Resolverها می‌توانند یک اتصال تونل را عبور دهند. شش پروتکل پشتیبانی می‌شود که هر کدام قالب Config خودش را دارد و به‌صورت یک فایل TOML جداگانه زیر `assets/dns-tunneling/` ذخیره می‌شود.

برای مدیریت Configهای تونل به **Main Menu → DNS Tunneling** بروید.

## پروتکل‌ها

| پروتکل | توضیحات | نیاز دارد |
| --- | --- | --- |
| [**DNSTT**](./dnstt/) | تونل DNS با کتابخانهٔ vaydns و قاب‌بندی سازگار با DNSTT | دامین، Public key |
| [**VayDNS**](./vaydns/) | پروتکل بومی vaydns با تنظیم QNAME، MTU و نوع Record | دامین، Public key |
| [**Slipstream**](./slipstream/) | تونل مبتنی بر باینری خارجی `slipstream-client` | دامین، باینری در PATH |
| [**MasterDNS**](./masterdns/) | تونل داخل برنامه با رمزنگاری قابل تنظیم و کشف MTU | دامین، کلید رمزنگاری |
| [**StormDNS**](./stormdns/) | تونل داخل برنامه با رمزنگاری قابل تنظیم، کشف MTU و انتخاب نوع کوئری | دامین، کلید رمزنگاری |
| [**TheFeed**](./thefeed/) | تونل فید رمزنگاری‌شده روی UDP، TCP یا DNS-over-TLS | دامین، Passphrase |

DNSTT، VayDNS، MasterDNS، StormDNS و TheFeed تونل را با کتابخانه‌های داخلی داخل خود برنامه اجرا می‌کنند. Slipstream باینری خارجی را صدا می‌زند و از طریق یک پورت محلی SOCKS5 با آن ارتباط برقرار می‌کند.

DNSTT، VayDNS و Slipstream از مسیریابی اختیاری Proxy SOCKS5 یا SSH با احراز هویت Password یا Key پشتیبانی می‌کنند. Configهای MasterDNS، StormDNS و TheFeed تنظیمات Proxy ندارند.

اسکن با Configهای DNSTT، VayDNS، Slipstream و TheFeed اجرا می‌شود. Configهای MasterDNS و StormDNS را می‌شود در منو ساخت، ویرایش و اعتبارسنجی کرد، ولی اجرای اسکن با آن‌ها هنوز وصل نشده است.

## نحوهٔ کار

اسکن تونل DNS وقتی `check_dns_resolver` فعال باشد در دو مرحله اجرا می‌شود:

1. **پیش‌اسکن Resolver** — هر IP هدف به‌عنوان Resolver DNS تست می‌شود. فقط Resolverهایی که پرس‌وجوهای ساده DNS (و بررسی DPI در صورت فعال‌بودن) را پاس کنند به مرحلهٔ تونل می‌روند.
2. **پروب تونل** — از هر Resolver باقی‌مانده آزمایش گرفته می‌شود که می‌تواند اتصال تونل را عبور دهد یا نه. پروب با پروتکل انتخابی تونل را بالا می‌آورد و بررسی می‌کند که ترافیک از آن رد می‌شود.

وقتی `adaptive_resolver` فعال باشد، پیش‌اسکن Resolver به‌طور خودکار همان روش انتقال، پورت و دامین Config تونل را استفاده می‌کند تا تست Resolver با مسیر واقعی تونل یکی باشد.

زمان پاسخ‌دهی گزارش‌شده بعد از برقرارشدن تونل اندازه گرفته می‌شود و هزینهٔ راه‌اندازی را حساب نمی‌کند.

## مدیریت Configها

جدول **Main Menu → DNS Tunneling** همهٔ تونل‌های ذخیره‌شده را نشان می‌دهد:

| ستون | توضیحات |
| --- | --- |
| Name | نام Config |
| Protocol | DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS یا TheFeed |
| Auth | روش احراز هویت (none، password یا key) |
| Created Time | زمان ساخت فایل |

| کلید | کار |
| --- | --- |
| `a` | افزودن Config جدید (انتخاب‌گر پروتکل باز می‌شود) |
| `r` | تغییر نام Config انتخاب‌شده |
| `x` | حذف Config انتخاب‌شده |
| `Enter` | ویرایش یا شروع اسکن با Config انتخاب‌شده |

## محل ذخیرهٔ Configها

Configهای تونل به‌صورت فایل TOML ذخیره می‌شوند:

<pre dir="ltr">
assets/dns-tunneling/
├── dnstt/
│   └── <config-name>.toml
├── vaydns/
│   └── <config-name>.toml
├── slipstream/
│   └── <config-name>.toml
├── masterdns/
│   └── <config-name>.toml
├── stormdns/
│   └── <config-name>.toml
└── thefeed/
    └── <config-name>.toml
</pre>

## پیش‌فرض‌های پلتفرم

تعداد Worker و Timeout به‌طور خودکار بر اساس پلتفرم تشخیص‌داده‌شده و سطح منابع تنظیم می‌شوند:

| تنظیم | Desktop Low | Desktop Mid | Desktop High |
| --- | --- | --- | --- |
| Workerهای تونل DNS | 8 | 16 | 32 |
| Timeout تونل DNS | 10s | 10s | 10s |
| Workerهای ریزالور DNS | 30 | 150 | 300 |

| تنظیم | Android Low | Android Mid | Android High |
| --- | --- | --- | --- |
| Workerهای تونل DNS | 3 | 6 | 12 |
| Timeout تونل DNS | 10s | 10s | 10s |
| Workerهای ریزالور DNS | 15 | 60 | 100 |

| تنظیم | Server Low | Server Mid | Server High |
| --- | --- | --- | --- |
| Workerهای تونل DNS | 12 | 24 | 64 |
| Timeout تونل DNS | 10s | 10s | 8s |
| Workerهای ریزالور DNS | 100 | 400 | 1000 |

## موضوعات مرتبط

- [تنظیمات DNS](../settings/dns.md) — تنظیمات Resolver و هماهنگی تونل
- [انواع اسکن](../scanner/scan-types.md) — جای تونل DNS در فرآیند اسکن
- [فایل‌های نتیجه](../scanner/result-files.md) — قالب فایل نتیجهٔ تونل
