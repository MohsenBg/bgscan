---
title: "تنظیمات HTTP"
weight: 7
---

# تنظیمات HTTP

> 💡 **نکته:** از **Settings → HTTP Settings** هم می‌توانید این گزینه‌ها را تغییر دهید.

فایل تنظیمات: `settings/http_settings.toml`

پروب HTTP برای هر هدف یک درخواست می‌فرستد و کد وضعیت، نسخهٔ توافق‌شدهٔ پروتکل و استفاده‌شدن TLS را ثبت می‌کند. HTTP/1.1 و HTTP/2 روی TCP هستند؛ HTTP/3 روی QUIC است.

## خلاصهٔ گزینه‌ها

| گزینه | پیش‌فرض | توضیح |
|---|---:|---|
| `workers` | `50` | درخواست هم‌زمان، ۱ تا ۵۰۰۰ |
| `host` | `"example.com"` | Host درخواست؛ می‌تواند Path هم داشته باشد |
| `server_name` | `""` | SNI دلخواه؛ خالی یعنی از `host` گرفته شود |
| `port` | `443` | پورت هدف، ۱ تا ۶۵۵۳۵ |
| `protocol` | `"https"` | `http` یا `https` |
| `version` | `"h1,h2"` | نسخهٔ HTTP برای مذاکره |
| `tls_validation` | `true` | بررسی اعتبار Certificate |
| `min_tls_version` | `"tls1.1"` | کمترین نسخهٔ TLS |
| `max_tls_version` | `"tls1.3"` | بیشترین نسخهٔ TLS |
| `timeout` | `4000` | زمان درخواست، میلی‌ثانیه |
| `accepted_status_codes` | `[]` | کدهای وضعیت قابل قبول؛ خالی یعنی همه |
| `prefix_output` | `"http_"` | پیشوند فایل نتیجه |

## Workers

```toml
workers = 50
```

تعداد درخواست‌هایی است که هم‌زمان اجرا می‌شوند. هر Worker در HTTP اتصال را نگه می‌دارد و پاسخ را می‌خواند، پس نسبت به TCP هزینهٔ بیشتری دارد.

- `10-50` برای منابع محدود
- `50-100` برای استفادهٔ معمول
- `100-200` برای اسکن پرسرعت

## Host

```toml
host = "example.com"
```

Host درخواست است. می‌توانید Path هم بدهید، مثل `example.com/path`؛ Path به URL اضافه می‌شود، اما هدر `Host` فقط بخش Domain را می‌گیرد.

## Port و Protocol

```toml
port = 443
protocol = "https"
```

برای HTTP ساده معمولاً پورت 80 و برای HTTPS پورت 443 را بگذارید. مقدار `protocol` فقط `http` یا `https` است و مشخص می‌کند اتصال با TLS پوشانده شود یا نه.

## TLS Validation

```toml
tls_validation = true
```

اگر `true` باشد، Certificate باید معتبر و مورداعتماد سیستم باشد. اگر `false` باشد، Certificate منقضی یا self-signed هم قبول می‌شود. هنگام اسکن مستقیم IP معمولاً Certificate با آدرس جور نیست، پس خاموش‌کردن آن کاربردی است.

## HTTP Version

```toml
version = "h1,h2"
```

| مقدار | معنی |
|---|---|
| `h1` | فقط HTTP/1.1 |
| `h2` | فقط HTTP/2 |
| `h1,h2` | انتخاب از طریق ALPN |
| `h3` | HTTP/3 روی QUIC |

نام‌های بلندتر هم پذیرفته می‌شوند: `http1`، `http1.1`، `http2`، `http3`، `http1,http2` و `http2,http1`. HTTP/3 با QUIC جداگانه اجرا می‌شود و همیشه TLS 1.3 دارد؛ پس `min_tls_version` و `max_tls_version` برای `h3` اثری ندارند.

## Timeout

```toml
timeout = 4000
```

بودجهٔ زمانی هر درخواست بر حسب میلی‌ثانیه است.

- `2000-5000` برای شبکه‌های مطمئن
- `5000-10000` برای شبکه‌های دور یا ناپایدار

## محدودهٔ TLS

```toml
min_tls_version = "tls1.1"
max_tls_version = "tls1.3"
```

حد پایین و بالای TLS برای Handshake هستند. مقدارهای مجاز `tls1.0`، `tls1.1`، `tls1.2` و `tls1.3` است و مقدار پایین نباید از مقدار بالا بزرگ‌تر باشد. در حالت `h3` نادیده گرفته می‌شوند.

## Server Name Indication

```toml
server_name = ""
```

SNIای است که هنگام TLS Handshake فرستاده می‌شود. اگر خالی باشد، Domain داخل `host` استفاده می‌شود. برای Domain Fronting یا اسکن یک IP با نام Host مشخص از آن استفاده کنید.

## Accepted Status Codes

```toml
accepted_status_codes = []
```

فهرست کدهای وضعیت HTTP که نتیجهٔ موفق حساب می‌شوند. فهرست خالی یا فهرستی که همهٔ کدها را پوشش دهد یعنی هیچ فیلتری وجود ندارد. پاسخ خارج از این فهرست دور ریخته می‌شود و به مرحلهٔ بعد Pipeline نمی‌رود.

```toml
accepted_status_codes = [200, 204, 301, 302, 307, 308]
```

## Prefix Output

```toml
prefix_output = "http_"
```

پیشوند فایل‌های نتیجهٔ این پروب است. فایل‌ها داخل `result/http/` ذخیره می‌شوند.

## فایل‌های مرتبط

- [`general_settings.toml`](./general.md)
- [`icmp_settings.toml`](./icmp.md)
- [`tcp_settings.toml`](./tcp.md)
- [`dns_settings.toml`](./dns.md)
- [`xray_settings.toml`](./xray.md)
- [`writer_settings.toml`](./writer.md)
