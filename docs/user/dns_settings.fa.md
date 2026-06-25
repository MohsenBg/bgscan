<div align="right">
  
  [**English**](./dns_settings.md) &nbsp;|&nbsp; [**فارسی**](./dns_settings.fa.md)
  
</div>

# پیکربندی اسکنر DNS

فایل پیکربندی: `settings/dns_settings.toml`

این فایل رفتار تمام ماژول‌های مبتنی بر DNS را کنترل می‌کند: **اسکنر Resolver عمومی**، **پروب DNSTT**، و **پروب Slipstream**.

---

## فهرست مطالب

- [Resolver — اسکنر DNS عمومی](#resolver--اسکنر-dns-عمومی)
  - [هسته اصلی](#هسته-اصلی)
  - [کوئری](#کوئری)
  - [زمان‌بندی و تلاش مجدد](#زمان‌بندی-و-تلاش-مجدد)
  - [فیلترینگ](#فیلترینگ)
  - [DPI / ضد هایجکینگ](#dpi--ضد-هایجکینگ)
  - [خروجی](#خروجی)
- [DNSTT — پروب تونل DNS](#dnstt--پروب-تونل-dns)
- [Slipstream — پروب تونل DNS](#slipstream--پروب-تونل-dns)

---

## `[resolver]` — اسکنر DNS عمومی

Resolver‌های عمومی DNS را آزمایش می‌کند تا موارد قابل دسترس و دارای رفتار صحیح را شناسایی کند. نتایج به پروب‌های بعدی (DNSTT، Slipstream) منتقل می‌شوند.

---

### هسته اصلی

#### `workers`

```toml
workers = 500
```

تعداد goroutine‌های همزمان برای اسکن resolver‌ها.

| بازه توصیه‌شده | نکته                                  |
| -------------- | ------------------------------------- |
| `50` – `500`   | متناسب با پهنای باند موجود تنظیم کنید |

---

#### `protocol`

```toml
protocol = "udp"
```

پروتکل انتقال استفاده‌شده برای کوئری‌های DNS.

| مقدار   | توضیح                       |
| ------- | --------------------------- |
| `"udp"` | DNS استاندارد UDP (پیش‌فرض) |
| `"tcp"` | DNS از طریق TCP             |
| `"dot"` | DNS-over-TLS                |

> **نکته:** `"doh"` (DNS-over-HTTPS) **پشتیبانی نمی‌شود**.

---

#### `domain`

```toml
domain = ""
```

نام دامنه اصلی که به عنوان هدف کوئری استفاده می‌شود. قبل از اجرا باید تنظیم شود.

---

#### `port`

```toml
port = 53
```

پورت DNS در زمان اسکن resolver‌ها. پورت استاندارد DNS برابر `53` است.

---

### کوئری

#### `check_types`

```toml
check_types = ["txt"]
```

نوع رکوردهای DNS که در طول اسکن کوئری می‌شوند.

| مقدار رایج | توضیح                                             |
| ---------- | ------------------------------------------------- |
| `"txt"`    | رکوردهای TXT — برای سازگاری با DNSTT توصیه می‌شود |
| `"a"`      | رکوردهای آدرس IPv4                                |
| `"aaaa"`   | رکوردهای آدرس IPv6                                |
| `"ns"`     | رکوردهای Nameserver                               |
| `"mx"`     | رکوردهای Mail Exchange                            |

---

#### `ends_buffer_size`

```toml
ends_buffer_size = 0
```

اندازه بافر EDNS (بر حسب بایت) که در رکوردهای OPT ارسال می‌شود. مقدار `0` یعنی از پیش‌فرض resolver استفاده شود.

---

### زمان‌بندی و تلاش مجدد

#### `timeout`

```toml
timeout = 2000
```

زمان انتظار برای دریافت پاسخ DNS بر حسب **میلی‌ثانیه**، قبل از اینکه resolver غیرقابل دسترس تلقی شود.

---

#### `tries`

```toml
tries = 2
```

تعداد تلاش‌های مجدد در سطح شبکه برای پروب اصلی.

> **مهم:** تلاش مجدد فقط در صورت خطای شبکه (timeout، قطع اتصال) انجام می‌شود. اگر **هر** پاسخ DNS دریافت شود — صرف نظر از Rcode — تلاش مجدد فوراً متوقف می‌شود.

---

#### `random_subdomain`

```toml
random_subdomain = true
```

وقتی فعال است، یک زیردامنه تصادفی به کوئری اضافه می‌کند (مثلاً `x1y2z3.example.com`).

**چرا فعال‌سازی توصیه می‌شود؟**

- کش سمت resolver را دور می‌زند
- یک lookup بازگشتی تازه را مجبور می‌کند
- اطمینان می‌دهد نتایج رفتار لحظه‌ای resolver را نشان می‌دهند

---

### فیلترینگ

#### `accepted_rcodes`

```toml
accepted_rcodes = ["noerror", "nxdomain"]
```

مشخص می‌کند کدام کدهای پاسخ DNS نشانه **زنده بودن** resolver هستند.

| مقدار                            | کد  | معنا                         |
| -------------------------------- | --- | ---------------------------- |
| `"noerror"` / `"success"`        | `0` | کوئری با موفقیت پاسخ داده شد |
| `"formerr"` / `"formaterror"`    | `1` | درخواست بدشکل                |
| `"servfail"` / `"serverfailure"` | `2` | خرابی سمت سرور               |
| `"nxdomain"` / `"nameerror"`     | `3` | دامنه وجود ندارد             |
| `"notimp"` / `"notimplemented"`  | `4` | نوع کوئری پشتیبانی نمی‌شود   |
| `"refused"`                      | `5` | کوئری توسط resolver رد شد    |

> **توصیه:** `["noerror", "nxdomain"]`

---

### DPI / ضد هایجکینگ

Resolver‌هایی را شناسایی می‌کند که پاسخ‌های DNS را دستکاری می‌کنند (مثلاً هایجکینگ در سطح ISP). این بررسی یک کوئری برای یک دامنه `.invalid` غیرواقعی ارسال می‌کند — اگر resolver مقدار `NOERROR` (Rcode 0) برگرداند، به عنوان resolver دستکاری‌شده علامت‌گذاری و حذف می‌شود.

#### `check_dpi`

```toml
check_dpi = false
```

پیش‌بررسی ضد هایجکینگ را فعال یا غیرفعال می‌کند.

---

#### `dpi_timeout`

```toml
dpi_timeout = 500
```

Timeout بررسی DPI بر حسب **میلی‌ثانیه**. عمداً کوتاه‌تر از `timeout` اصلی تنظیم شده تا resolver‌های بد سریع‌تر رد شوند.

---

#### `dpi_tries`

```toml
dpi_tries = 2
```

حداکثر تلاش‌های مجدد شبکه برای بررسی DPI. فقط در صورت timeout یا خطای شبکه تلاش مجدد انجام می‌شود.

---

### خروجی

#### `prefix_output`

```toml
prefix_output = "dns_"
```

پیشوند اضافه‌شده به نام فایل خروجی (مثلاً `dns_results.txt`).

---

## `[dnstt]` — پروب تونل DNS

آزمایش می‌کند آیا resolver‌هایی که از اسکن عمومی عبور کردند می‌توانند یک تونل DNS کامل را تحمل کنند. نیاز به یک سرور [DNSTT](https://www.bamsoftware.com/software/dnstt/) دارد که روی Nameserver اقتداری در حال اجرا باشد.

| تنظیم           | پیش‌فرض    | توضیح                                               |
| --------------- | ---------- | --------------------------------------------------- |
| `enabled`       | `false`    | این فاز پروب را فعال یا غیرفعال کنید                |
| `workers`       | `20`       | تعداد worker‌های handshake همزمان (توصیه: `5`–`50`) |
| `domain`        | `""`       | Zone DNS اقتداری که به سرور DNSTT شما واگذار شده    |
| `public_key`    | `""`       | کلید عمومی Ed25519 سرور DNSTT به فرمت Base64        |
| `timeout`       | `10000`    | میلی‌ثانیه مجاز برای یک handshake کامل              |
| `prefix_output` | `"dnstt_"` | پیشوند نام فایل خروجی                               |

**مثال:**

```toml
[dnstt]
enabled    = true
workers    = 20
domain     = "tunnel.example.com"
public_key = "base64encodedpublickey=="
timeout    = 10000
```

---

## `[slip_stream]` — پروب Slipstream DNS

یک تکنیک تونل‌سازی جایگزین که از رفتار DNS بهره می‌برد. مستقل از DNSTT عمل می‌کند و می‌تواند به‌صورت همزمان با آن اجرا شود.

| تنظیم           | پیش‌فرض         | توضیح                                             |
| --------------- | --------------- | ------------------------------------------------- |
| `enabled`       | `true`          | این فاز پروب را فعال یا غیرفعال کنید              |
| `workers`       | `20`            | تعداد worker‌های پروب همزمان (توصیه: `5`–`50`)    |
| `domain`        | `""`            | Zone DNS اقتداری استفاده‌شده توسط سرور Slipstream |
| `cert_path`     | `""`            | مسیر گواهی TLS استفاده‌شده توسط سرور Slipstream   |
| `timeout`       | `8000`          | میلی‌ثانیه مجاز برای یک پروب کامل                 |
| `prefix_output` | `"slipstream_"` | پیشوند نام فایل خروجی                             |

**مثال:**

```toml
[slip_stream]
enabled     = true
workers     = 20
domain      = "slip.example.com"
cert_path   = "/etc/certs/slipstream.crt"
timeout     = 8000
```

---

## مثال کامل

```toml
[resolver]
workers           = 500
protocol          = "udp"
domain            = "example.com"
port              = 53
check_types       = ["txt"]
ends_buffer_size  = 0
timeout           = 2000
tries             = 2
random_subdomain  = true
accepted_rcodes   = ["noerror", "nxdomain"]
check_dpi         = true
dpi_timeout       = 500
dpi_tries         = 2
prefix_output     = "dns_"

[dnstt]
enabled       = true
workers       = 20
domain        = "tunnel.example.com"
public_key    = "base64encodedpublickey=="
timeout       = 10000
prefix_output = "dnstt_"

[slip_stream]
enabled       = true
workers       = 20
domain        = "slip.example.com"
cert_path     = "/etc/certs/slipstream.crt"
timeout       = 8000
prefix_output = "slipstream_"
```
