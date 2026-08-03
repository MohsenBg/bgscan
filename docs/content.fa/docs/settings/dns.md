---
title: "تنظیمات DNS"
weight: 8
---

# تنظیمات DNS

> 💡 **نکته:** از **Settings → DNS Settings** هم می‌توانید این گزینه‌ها را تغییر دهید.

فایل تنظیمات: `settings/dns_settings.toml`

این فایل سه بخش جدا دارد: Resolver برای DNS معمولی، و DNSTT و Slipstream برای بررسی Tunnel DNS. دو بخش Tunnel به Client خارجی نیاز دارند.

## خلاصهٔ گزینه‌ها

| گزینه | پیش‌فرض | توضیح |
|---|---:|---|
| `resolver.workers` | `100` | پرس‌وجوی DNS هم‌زمان، ۱ تا ۲۵۰۰ |
| `resolver.protocol` | `"udp"` | `udp`، `tcp`، `dot` یا `doh` |
| `resolver.domain` | `"google.com"` | Domain مورد پرس‌وجو |
| `resolver.port` | `53` | پورت Resolver |
| `resolver.check_types` | `["A"]` | Record typeها به‌ترتیب آزمایش |
| `resolver.ends_buffer_size` | `1234` | EDNS0 UDP buffer |
| `resolver.timeout` | `2000` | Timeout هر پرس‌وجو |
| `resolver.tries` | `1` | تلاش برای هر Record type |
| `resolver.random_subdomain` | `true` | Subdomain تصادفی برای دورزدن Cache |
| `resolver.accepted_rcodes` | `["noerror", "nxdomain"]` | RCodeهای قابل قبول |
| `resolver.check_dpi` | `true` | بررسی Hijacking قبل از اسکن |
| `resolver.dpi_timeout` | `500` | Timeout بررسی DPI |
| `resolver.dpi_tries` | `2` | تلاش بررسی DPI |
| `resolver.prefix_output` | `"dns_resolver_"` | پیشوند فایل Resolver |
| `dnstt.enabled` | `false` | فعال‌کردن مرحلهٔ DNSTT |
| `dnstt.workers` | `20` | DNSTT هم‌زمان، ۱ تا ۵۰۰ |
| `dnstt.domain` | `"ns.example.com"` | Zone سرور DNSTT |
| `dnstt.public_key` | ۶۴ صفر | Public key سرور، ۶۴ کاراکتر hex |
| `dnstt.timeout` | `8000` | Timeout Handshake |
| `dnstt.prefix_output` | `"dns_dnstt_"` | پیشوند فایل DNSTT |
| `slip_stream.enabled` | `false` | فعال‌کردن Slipstream |
| `slip_stream.workers` | `20` | Slipstream هم‌زمان، ۱ تا ۵۰۰ |
| `slip_stream.domain` | `"ns.example.com"` | Zone سرور Slipstream |
| `slip_stream.cert_path` | `""` | Certificate TLS اختیاری Client |
| `slip_stream.timeout` | `8000` | Timeout پروب |
| `slip_stream.prefix_output` | `"dns_slipstream_"` | پیشوند فایل Slipstream |

## Resolver

هر IP هدف به‌عنوان Resolver استفاده می‌شود. bgscan `domain` را از طریق آن می‌پرسد و فقط وقتی IP را نتیجهٔ موفق می‌داند که کد پاسخ داخل `accepted_rcodes` باشد.

### Workers

```toml
workers = 100
```

تعداد پرس‌وجوهای هم‌زمان است. UDP سبک است، اما نرخ خیلی بالا روی یک شبکهٔ بالادستی قابل‌مشاهده خواهد بود.

### Protocol

```toml
protocol = "udp"
```

روش انتقال پرس‌وجو `udp`، `tcp`، `dot` یا `doh` است. DoH در تنظیمات پذیرفته می‌شود اما هنگام اجرا به DoT تبدیل می‌شود، چون DoH Endpoint مبتنی بر Domain می‌خواهد و هدف‌های اسکنر IP هستند.

### Domain و Port

```toml
domain = "google.com"
port = 53
```

`domain` باید فقط Domain باشد؛ Scheme، پورت یا Path ندهید. نامی انتخاب کنید که همه‌جا قابل Resolve باشد. برای DoT از پورت 853 استفاده کنید.

### Check Types

```toml
check_types = ["A"]
```

Record typeها به‌ترتیب آزمایش می‌شوند. با اولین نوعی که RCode قابل قبول بدهد، پروب متوقف می‌شود و همان نوع در نتیجه ثبت می‌شود.

### EDNS Buffer Size

```toml
ends_buffer_size = 1234
```

اندازهٔ UDP payload اعلام‌شده در رکورد OPT است. `0`، EDNS0 را خاموش می‌کند. نام کلید در فایل دقیقاً `ends_buffer_size` است.

### Timeout و Tries

```toml
timeout = 2000
tries = 1
```

Timeout برای هر پرس‌وجو است. Retry فقط برای خطاهای شبکه انجام می‌شود. وقتی هر پاسخ DNS برسد، حتی با RCode ردشده، پروب بدون Retry سراغ Record type بعدی می‌رود.

### Random Subdomain

```toml
random_subdomain = true
```

برای هر پروب یک Label تصادفی ۱۰کاراکتری به اول Domain اضافه می‌کند. Cache Resolver دور زده می‌شود و Latency کار واقعی Resolver را نشان می‌دهد.

### Accepted RCodes

```toml
accepted_rcodes = ["noerror", "nxdomain"]
```

| مقدار | نام دیگر | کد |
|---|---|---:|
| `noerror` | `success` | 0 |
| `formerr` | `formaterror` | 1 |
| `servfail` | `serverfailure` | 2 |
| `nxdomain` | `nameerror` | 3 |
| `notimp` | `notimplemented` | 4 |
| `refused` | | 5 |

وقتی Subdomain تصادفی فعال باشد، `nxdomain` برای نام ساختگی طبیعی است و به همین دلیل پیش‌فرض قبول می‌شود.

### Check DPI

```toml
check_dpi = true
```

قبل از پرس‌وجوی واقعی اجرا می‌شود. bgscan یک نام تصادفی `.invalid` می‌پرسد که نمی‌تواند وجود داشته باشد. اگر Resolver پاسخ `NOERROR` بدهد، پاسخ ساختگی می‌دهد و IP کنار گذاشته می‌شود. هر RCode دیگر سالم حساب می‌شود. نتیجه با `passed` یا `skipped` ثبت می‌شود.

### DPI Timeout و Tries

```toml
dpi_timeout = 500
dpi_tries = 2
```

Timeout و تعداد تلاش بررسی Hijacking هستند. Timeout را خیلی کمتر از Timeout اصلی بگذارید تا هدف مرده سریع کنار برود.

### Prefix Output

```toml
prefix_output = "dns_resolver_"
```

فایل‌ها در `result/dns_resolver/` ذخیره می‌شوند.

## DNSTT

بررسی می‌کند Resolver می‌تواند Tunnel DNSTT را عبور دهد یا نه. این بخش `dnstt-client` را اجرا می‌کند؛ باینری باید داخل `assets/` یا در `PATH` باشد. برای هر پروب یک پورت محلی SOCKS5 می‌گیرد. اگر باینری نباشد، در شروع برنامه هشدار ثبت و این اسکن غیرفعال می‌شود.

Latency گزارش‌شده بعد از برقرارشدن Tunnel اندازه گرفته می‌شود و زمان راه‌اندازی را حساب نمی‌کند.

```toml
enabled = false
workers = 20
domain = "ns.example.com"
public_key = "0000000000000000000000000000000000000000000000000000000000000000"
timeout = 8000
prefix_output = "dns_dnstt_"
```

هر Worker Process Client خودش و یک پورت محلی دارد، پس از Resolver سنگین‌تر است. `domain` Zone واگذار‌شده به سرور DNSTT شماست. Public key با `-pubkey` به Client می‌رود و باید دقیقاً ۶۴ کاراکتر hexadecimal باشد؛ مقدار پیش‌فرض فقط نمونه است و وصل نمی‌شود. فایل‌ها در `result/dnstt/` می‌روند.

## Slipstream

Slipstream یک روش دیگر Tunnel DNS است و شکل کارش مانند DNSTT است: `slipstream-client` خارجی، یک پورت محلی SOCKS5 برای هر پروب و اندازه‌گیری Latency پس از ایجاد Tunnel.

```toml
enabled = false
workers = 20
domain = "ns.example.com"
cert_path = ""
timeout = 8000
prefix_output = "dns_slipstream_"
```

`domain` Zone سرور Slipstream شماست. `cert_path` مسیر اختیاری Certificate TLS است که با `--cert` به Client داده می‌شود؛ اگر سرور Certificate نمی‌خواهد خالی بگذارید. فایل‌ها در `result/slipstream/` ذخیره می‌شوند.

## فایل‌های مرتبط

- [`general_settings.toml`](./general.md)
- [`icmp_settings.toml`](./icmp.md)
- [`tcp_settings.toml`](./tcp.md)
- [`http_settings.toml`](./http.md)
- [`xray_settings.toml`](./xray.md)
- [`writer_settings.toml`](./writer.md)
