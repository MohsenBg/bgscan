---
title: "تونل DNS"
weight: 7
---

# تونل DNS

bgscan بررسی می‌کند کدام Resolverها می‌توانند یک اتصال تونل را عبور دهند. سه پروتکل پشتیبانی می‌شود که هر کدام قالب Config خودش را دارد و به‌صورت یک فایل TOML جداگانه زیر `assets/dns-tunneling/` ذخیره می‌شود.

برای مدیریت Configهای تونل به **Main Menu → DNS Tunneling** بروید.

## پروتکل‌ها

| پروتکل | توضیحات | نیاز دارد |
| --- | --- | --- |
| **DNSTT** | تونل DNS با کتابخانهٔ vaydns و قاب‌بندی سازگار با DNSTT | دامین، Public key |
| **VayDNS** | پروتکل بومی vaydns با تنظیم QNAME، MTU و نوع Record | دامین، Public key |
| **Slipstream** | تونل مبتنی بر باینری خارجی `slipstream-client` | دامین، باینری در PATH |

DNSTT و VayDNS تونل را با کتابخانهٔ vaydns داخل خود برنامه اجرا می‌کنند. Slipstream باینری خارجی را صدا می‌زند و از طریق یک پورت محلی SOCKS5 با آن ارتباط برقرار می‌کند.

هر سه پروتکل از مسیریابی اختیاری Proxy SOCKS5 یا SSH با احراز هویت Password یا Key پشتیبانی می‌کنند.

## نحوهٔ کار

اسکن تونل DNS وقتی `check_dns_resolver` فعال باشد در دو مرحله اجرا می‌شود:

1. **پیش‌اسکن Resolver** — هر IP هدف به‌عنوان Resolver DNS تست می‌شود. فقط Resolverهایی که پرس‌وجوهای ساده DNS (و بررسی DPI در صورت فعال‌بودن) را پاس کنند به مرحلهٔ تونل می‌روند.
2. **پروب تونل** — از هر Resolver باقی‌مانده آزمایش گرفته می‌شود که می‌تواند اتصال تونل را عبور دهد یا نه. پروب کل پشتهٔ تونل را برپا می‌کند: کانال بسته‌های DNS روی انتقال KCP با رمزنگاری Noise و چندگانه‌سازی smux؛ و بعد اتصال را اعتبارسنجی می‌کند.

وقتی `adaptive_resolver` فعال باشد، پیش‌اسکن Resolver به‌طور خودکار همان روش انتقال، پورت و دامین Config تونل را استفاده می‌کند تا تست Resolver با مسیر واقعی تونل یکی باشد.

زمان پاسخ‌دهی گزارش‌شده بعد از برقرارشدن تونل اندازه گرفته می‌شود و هزینهٔ راه‌اندازی را حساب نمی‌کند.

## مدیریت Configها

جدول **Main Menu → DNS Tunneling** همهٔ تونل‌های ذخیره‌شده را نشان می‌دهد:

| ستون | توضیحات |
| --- | --- |
| Name | نام Config |
| Protocol | DNSTT، VayDNS یا Slipstream |
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
└── slipstream/
    └── <config-name>.toml
</pre>

---

## تنظیمات DNSTT

DNSTT از کتابخانهٔ vaydns با قاب‌بندی سازگار با DNSTT استفاده می‌کند (DNS-over-TXT با قاب‌بندی قدیمی). پشتهٔ تونل این است: کانال بسته‌های DNS روی انتقال KCP با رمزنگاری Noise و چندگانه‌سازی smux.

### تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `Domain` | (لازم) | دامین معتبر | Zone واگذارشده به سرور DNSTT شما |
| `PubKey` | (لازم) | ۶۴ کاراکتر hex | کلید عمومی سرور |
| `ResolverType` | `"udp"` | `udp`، `tcp`، `dot` | روش انتقال به Resolver |
| `ResolverPort` | `53` | بزرگ‌تر از 0 | پورت Resolver |
| `Fingerprint` | `"Chrome"` | جدول پایین | اثر انگشت uTLS ClientHello |
| `RPS` | `0` | 0-500 | محدودیت نرخ (پرس‌وجو در ثانیه)؛ 0 یعنی بدون محدودیت |

### پروکسی و احراز هویت

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

### TLS Fingerprint

فیلد `Fingerprint` یک پروفایل uTLS ClientHello انتخاب می‌کند که در TLS Handshake تقلید شود. برچسب‌ها به بزرگی و کوچکی حروف حساس نیستند.

| دسته | برچسب‌ها |
| --- | --- |
| Chrome | `Chrome`, `Chrome_58`, `Chrome_62`, `Chrome_70`, `Chrome_72`, `Chrome_83`, `Chrome_87`, `Chrome_96`, `Chrome_100`, `Chrome_102`, `Chrome_120` |
| Firefox | `Firefox`, `Firefox_55`, `Firefox_56`, `Firefox_63`, `Firefox_65`, `Firefox_99`, `Firefox_102`, `Firefox_105`, `Firefox_120` |
| iOS | `iOS`, `iOS_11_1`, `iOS_12_1`, `iOS_13`, `iOS_14` |
| Other | `random` |

### نمونهٔ Config

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

---

## تنظیمات VayDNS

VayDNS پروتکل بومی vaydns است. همان پشتهٔ تونل DNSTT را دارد (کانال بسته‌های DNS روی انتقال KCP با رمزنگاری Noise و چندگانه‌سازی smux) اما با قاب‌بندی بومی و پارامترهای تنظیم اضافه برای ساختار QNAME، MTU و نوع Record.

### تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `Domain` | (لازم) | دامین معتبر | Zone واگذارشده به سرور VayDNS شما |
| `PubKey` | (لازم) | ۶۴ کاراکتر hex | کلید عمومی سرور |
| `ResolverType` | `"udp"` | `udp`، `tcp`، `dot` | روش انتقال به Resolver |
| `ResolverPort` | `53` | بزرگ‌تر از 0 | پورت Resolver |
| `Fingerprint` | `"Chrome"` | جدول DNSTT | اثر انگشت uTLS ClientHello |
| `RecordType` | `"TXT"` | پایین را ببینید | نوع Record برای پرس‌وجوهای تونل |
| `RPS` | `0` | 0-500 | محدودیت نرخ (پرس‌وجو در ثانیه)؛ 0 یعنی بدون محدودیت |

### تنظیمات پیشرفته

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `ClientIDSize` | `2` | 1-8 | اندازهٔ Client ID بر حسب بایت |
| `MaxQnameLen` | `101` | 0-253 | بیشترین طول QNAME در بسته‌های DNS (0 = خودکار) |
| `MaxNumLabels` | `0` | 0-4 | بیشترین تعداد Label در QNAME (0 = خودکار) |
| `MTU` | `0` | 0-1452 | بیشترین واحد انتقال (0 = خودکار) |

### نوع Recordهای DNS

مقادیر پشتیبانی‌شدهٔ `RecordType`: `A`، `AAAA`، `CNAME`، `NS`، `MX`، `TXT`، `SRV`، `NULL`، `CAA`.

نوع Record تعیین می‌کند پرس‌وجوهای تونل با کدام نوع Record انجام شوند. `TXT` پیش‌فرض است و بیشترین پشتیبانی را دارد.

### پروکسی و احراز هویت

همان فیلدهای DNSTT است؛ جدول [پروکسی و احراز هویت](#پروکسی-و-احراز-هویت) در بخش DNSTT را ببینید.

### نمونهٔ Config

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

---

## تنظیمات Slipstream

Slipstream به‌صورت یک پروسهٔ خارجی (باینری `slipstream-client`) اجرا می‌شود، نه تونل درون خود برنامه. ارتباط از طریق پورت محلی SOCKS5ای که برای هر پروب گرفته می‌شود برقرار می‌گردد.

### تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیحات |
| --- | --- | --- | --- |
| `Domain` | (لازم) | دامین معتبر | Zone سرویس‌دهی‌شده توسط سرور Slipstream شما |
| `ResolverPort` | `53` | بزرگ‌تر از 0 | پورت Resolver |
| `CertPath` | `""` | مسیر | Certificate TLS اختیاری (با `--cert` به باینری پاس داده می‌شود) |

### پروکسی و احراز هویت

همان فیلدهای DNSTT است؛ جدول [پروکسی و احراز هویت](#پروکسی-و-احراز-هویت) در بخش DNSTT را ببینید.

### محل باینری

bgscan باینری `slipstream-client` را در این مسیرها جست‌وجو می‌کند:

1. `<bgscan-root>/assets/slipstream-client/slipstream-client`
2. `<bgscan-root>/assets/slipstream/slipstream-client/slipstream-client`
3. `<bgscan-root>/slipstream-client/slipstream-client`
4. `<bgscan-root>/slipstream-client`
5. `PATH` سیستم

### نمونهٔ Config

```toml
Domain = "example.com"
ResolverPort = 53
CertPath = ""
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

---

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

---

## موضوعات مرتبط

- [تنظیمات DNS](../settings/dns.md) — تنظیمات Resolver و هماهنگی تونل
- [انواع اسکن](./scan-types.md) — جای تونل DNS در فرآیند اسکن
- [فایل‌های نتیجه](./result-files.md) — قالب فایل نتیجهٔ تونل
