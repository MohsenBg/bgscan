---
title: "تنظیمات DNS"
weight: 8
---

# تنظیمات DNS

> 💡 **نکته:** از **Settings → DNS Settings** هم می‌توانید این گزینه‌ها را تغییر دهید.

فایل تنظیمات: `settings/dns_settings.toml`

این فایل دو بخش دارد: `resolver` برای تست Resolver معمولی، و `dns_tunneling` برای هماهنگی اسکن Tunnel DNS. تنظیمات مخصوص هر پروتکل Tunnel (DNSTT، VayDNS و Slipstream) داخل این فایل نیست و در فایل‌های TOML جداگانه زیر `assets/dns-tunneling/` ذخیره می‌شود.

## خلاصهٔ گزینه‌ها

| گزینه | پیش‌فرض | توضیح |
|---|---:|---|
| `resolver.workers` | وابسته به سیستم | پرس‌وجوی DNS هم‌زمان، ۱ تا ۲۵۰۰ |
| `resolver.protocol` | `"udp"` | روش انتقال: `udp`، `tcp` یا `dot` |
| `resolver.domain` | `"example.com"` | Domainی که از هر Resolver پرسیده می‌شود |
| `resolver.port` | `53` | پورت Resolver، ۱ تا ۶۵۵۳۵ |
| `resolver.check_types` | `["TXT"]` | Record typeها به‌ترتیب آزمایش |
| `resolver.edns_buffer_size` | `1232` | اندازهٔ بافر EDNS0 بر حسب بایت؛ `0` خاموش |
| `resolver.timeout` | وابسته به سیستم | Timeout هر پرس‌وجو، میلی‌ثانیه |
| `resolver.tries` | `1` | تلاش برای هر Record type، ۱ تا ۱۰ |
| `resolver.random_subdomain` | `true` | افزودن Label تصادفی برای دورزدن Cache |
| `resolver.accepted_rcodes` | `["NOERROR","NXDOMAIN","SERVFAIL"]` | RCodeهای حساب‌شده به‌عنوان Resolver زنده |
| `resolver.output_prefix` | `"dns_"` | پیشوند فایل نتیجه |
| `resolver.dpi.enabled` | `true` | اجرای بررسی Hijacking قبل از پرس‌وجو |
| `resolver.dpi.timeout` | `2000` | Timeout بررسی DPI، میلی‌ثانیه |
| `resolver.dpi.tries` | `1` | تلاش بررسی DPI، ۱ تا ۱۰ |
| `dns_tunneling.workers` | وابسته به سیستم | Workerهای تست Tunnel هم‌زمان، ۱ تا ۵۰۰ |
| `dns_tunneling.tries` | `1` | تلاش برای هر هدف، ۱ تا ۱۰ |
| `dns_tunneling.timeout` | وابسته به سیستم | Timeout تست Tunnel، میلی‌ثانیه |
| `dns_tunneling.check_dns_resolver` | `true` | اجرای اسکن Resolver قبل از تست Tunnel |
| `dns_tunneling.adaptive_resolver` | `true` | تطبیق تنظیمات Resolver با Config Tunnel |
| `dns_tunneling.output_prefix` | `"dns_tun_"` | پیشوند فایل نتیجه |

## Resolver

هر IP هدف به‌عنوان Resolver استفاده می‌شود. پروب `domain` را از طریق آن می‌پرسد و فقط وقتی هدف را نگه می‌دارد که کد پاسخ داخل `accepted_rcodes` باشد.

### Workers

```toml
[resolver]
workers = 150
```

تعداد پرس‌وجوهای هم‌زمان. پیش‌فرض مؤثر به پلتفرم و منابع سیستم بستگی دارد. پرس‌وجوی UDP سبک است، پس این مقدار بیشتر از پروب HTTP جا دارد، اما نرخ خیلی بالا روی یک شبکهٔ بالادستی قابل‌مشاهده خواهد بود.

### Protocol

```toml
[resolver]
protocol = "udp"
```

روش انتقال پرس‌وجوها: `udp`، `tcp` یا `dot`. خواندن مقدار حساس به بزرگی و کوچکی حروف نیست و مقدار ناشناخته به `udp` برمی‌گردد.

### Domain

```toml
[resolver]
domain = "example.com"
```

Domainی که از هر Resolver پرسیده می‌شود. باید فقط Domain باشد؛ Scheme، پورت یا Path ندهید. نامی انتخاب کنید که همه‌جا قابل Resolve باشد، وگرنه Resolverهای سالم هم خراب به نظر می‌رسند.

### Port

```toml
[resolver]
port = 53
```

پورتی که Resolver روی آن گوش می‌دهد. با `protocol = "dot"` پورت 853 را بگذارید.

### Check Types

```toml
[resolver]
check_types = ["TXT"]
```

Record typeها به‌ترتیب آزمایش می‌شوند. پروب با اولین نوعی که RCode قابل قبول بدهد متوقف می‌شود و همان نوع در نتیجه ثبت می‌شود. اگر ممکن است یک Resolver برای نوعی جواب بدهد و نوع دیگری را رد کند، نوع بیشتری به فهرست اضافه کنید.

Record typeهای پشتیبانی‌شده: `A`، `AAAA`، `CNAME`، `NS`، `MX`، `TXT`، `SRV`، `NULL`، `CAA`.

### EDNS Buffer Size

```toml
[resolver]
edns_buffer_size = 1232
```

اندازهٔ UDP payload اعلام‌شده در رکورد OPT. `0` یعنی EDNS0 خاموش.

### Timeout

```toml
[resolver]
timeout = 2000
```

Timeout هر پرس‌وجو بر حسب میلی‌ثانیه.

### Tries

```toml
[resolver]
tries = 1
```

تعداد تلاش برای هر Record type. Retry فقط خطاهای شبکه را پوشش می‌دهد. به‌محض رسیدن هر پاسخ DNS، حتی با RCode ردشده، پروب بدون تلاش دوباره سراغ Record type بعدی می‌رود.

### Random Subdomain

```toml
[resolver]
random_subdomain = true
```

برای هر پروب یک Label تصادفی ۱۰کاراکتری به اول `domain` اضافه می‌کند. این کار Cache Resolver را دور می‌زند و Lookup بازگشتی واقعی را اجباری می‌کند، پس Latency نشان‌دهندهٔ کار واقعی Resolver است نه پاسخ Cache‌شده.

### Accepted RCodes

```toml
[resolver]
accepted_rcodes = ["NOERROR", "NXDOMAIN", "SERVFAIL"]
```

RCodeهایی که Resolver را زنده حساب می‌کنند.

| مقدار | نام دیگر | کد |
|---|---|---:|
| `NOERROR` | `success` | 0 |
| `FORMERR` | `formaterror` | 1 |
| `SERVFAIL` | `serverfailure` | 2 |
| `NXDOMAIN` | `nameerror` | 3 |
| `NOTIMP` | `notimplemented` | 4 |
| `REFUSED` | | 5 |

با فعال‌بودن `random_subdomain`، پاسخ `NXDOMAIN` برای Label ساختگی طبیعی است و به همین دلیل به‌صورت پیش‌فرض پذیرفته می‌شود.

### Output Prefix

```toml
[resolver]
output_prefix = "dns_"
```

پیشوند فایل‌های نتیجهٔ Resolver. فایل‌ها داخل `result/dns_resolver/` ذخیره می‌شوند.

### DPI Check

```toml
[resolver.dpi]
enabled = true
timeout = 2000
tries = 1
```

بررسی DPI (Deep Packet Inspection) قبل از پرس‌وجوهای واقعی اجرا می‌شود. پروب از Resolver یک نام تصادفی `.invalid` می‌پرسد که نمی‌تواند وجود داشته باشد. Resolverی که به چنین نامی پاسخ `NOERROR` بدهد دارد نتیجه ساخت می‌کند، پس هدف کنار گذاشته می‌شود. هر RCode دیگری سالم حساب می‌شود. نتیجهٔ این بررسی با `passed` یا `skipped` در هر رکورد ثبت می‌شود.

`timeout` بر حسب میلی‌ثانیه است (محدودهٔ ۱۰۰ تا ۱۰۰۰۰) و `tries` از ۱ تا ۱۰ می‌گیرد. Timeout بررسی DPI را خیلی کمتر از `timeout` اصلی بگذارید تا هدف‌های مرده سریع کنار بروند.

## DNS Tunneling

بخش `dns_tunneling` هماهنگی اسکن Tunnel را کنترل می‌کند و تنظیمات خود پروتکل‌ها را ندارد. هر Worker یک پروب Tunnel کامل اجرا می‌کند و یک پورت محلی SOCKS5 نگه می‌دارد، پس بسیار سنگین‌تر از پروب Resolver است.

### Workers

```toml
[dns_tunneling]
workers = 16
```

تعداد Workerهای تست Tunnel هم‌زمان. پیش‌فرض مؤثر به پلتفرم و منابع سیستم بستگی دارد.

### Tries

```toml
[dns_tunneling]
tries = 1
```

تعداد تلاش برای هر هدف.

### Timeout

```toml
[dns_tunneling]
timeout = 10000
```

بودجهٔ زمانی برپاکردن Tunnel و اعتبارسنجی آن بر حسب میلی‌ثانیه. تست Tunnel به زمان بیشتری از پرس‌وجوی معمولی DNS نیاز دارد.

### Check DNS Resolver

```toml
[dns_tunneling]
check_dns_resolver = true
```

وقتی `true` باشد، قبل از تست Tunnel یک اسکن Resolver اجرا می‌شود و فقط Resolverهایی که از آن عبور کنند به‌عنوان کاندید Tunnel تست می‌شوند. این کار از هدررفتن پروب‌های Tunnel روی Resolverهایی که حتی به پرس‌وجوی ساده هم جواب نمی‌دهند جلوگیری می‌کند.

### Adaptive Resolver

```toml
[dns_tunneling]
adaptive_resolver = true
```

وقتی `true` باشد، تنظیمات Resolver (روش انتقال، پورت و Domain) به‌طور خودکار با Config Tunnel انتخاب‌شده هماهنگ می‌شوند. مثلاً اگر Config DNSTT با `resolver_type = "tcp"` روی پورت 853 تعریف شده باشد، اسکن Resolver هم همان مسیر را تست می‌کند. وقتی `false` باشد، اسکن Resolver از تنظیمات بخش `[resolver]` همان‌طور که هستند استفاده می‌کند.

### Output Prefix

```toml
[dns_tunneling]
output_prefix = "dns_tun_"
```

پیشوند فایل‌های نتیجهٔ Tunnel. فایل‌ها بسته به پروتکل داخل `result/dnstt/`، `result/vaydns/` یا `result/slipstream/` ذخیره می‌شوند.

## Configهای Tunnel

پروتکل‌های Tunnel DNS (DNSTT، VayDNS و Slipstream) با فایل‌های TOML جداگانه زیر `assets/dns-tunneling/` تنظیم می‌شوند:

```
assets/dns-tunneling/
├── dnstt/
│   ├── my-dnstt-config.toml
│   └── ...
├── vaydns/
│   ├── my-vaydns-config.toml
│   └── ...
└── slipstream/
    ├── my-slipstream-config.toml
    └── ...
```

این Configها از داخل برنامه در **Main Menu → DNS Tunneling** ساخته و مدیریت می‌شوند. هر Config یک نام، نوع پروتکل و فیلدهای مخصوص همان پروتکل دارد. برای تنظیمات هر پروتکل [Tunnel DNS](../scanner/dns-tunneling.md) را ببینید.

## فایل‌های مرتبط

- [`general_settings.toml`](./general.md)
- [`icmp_settings.toml`](./icmp.md)
- [`tcp_settings.toml`](./tcp.md)
- [`http_settings.toml`](./http.md)
- [`xray_settings.toml`](./xray.md)
- [`writer_settings.toml`](./writer.md)
- [Tunnel DNS](../scanner/dns-tunneling.md) — تنظیمات پروتکل‌های Tunnel
