<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/logo-light.svg">
  <img src="./images/logo-light.svg" alt="BGSCAN" width="520" style="max-width:100%;">
</picture>

یک اسکنر سریع IP برای چند پروتکل، با قابلیت وصل‌کردن چند مرحله به هم

[English](./README.md) | [فارسی](./README.fa.md)

---

[![Go Version](https://img.shields.io/badge/Go-1.26.3+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-6366f1?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20|%20Windows%20|%20macOS%20|%20Termux-64748b?style=flat-square)](https://github.com/MohsenBg/bgscan/releases)
[![UI](https://img.shields.io/badge/UI-BubbleTea%20TUI-ec4899?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-22c55e?style=flat-square)](https://github.com/MohsenBg/bgscan/releases/latest)

[نصب](#نصب) · [شروع سریع](#شروع-سریع) · [مستندات](#مستندات) · [پروتکل‌های پشتیبانی‌شده](#پروتکل‌های-پشتیبانی‌شده)

</div>

## bgscan چیست؟

**bgscan** یک اسکنر شبکه است که داخل ترمینال اجرا می‌شود و با Go نوشته شده است. با آن می‌توانید IPها را با ICMP، TCP، HTTP، DNS و Xray بررسی کنید و چند مرحلهٔ اسکن را پشت سر هم اجرا کنید. همه‌چیز هم از طریق یک رابط متنی و با صفحه‌کلید انجام می‌شود.

با bgscan می‌توانید IPهای فعال را پیدا کنید، سرویس‌های وب را بررسی کنید، Resolverهای DNS را آزمایش کنید، تونل‌ها را بررسی کنید و اتصال Xray را بسنجید. نتیجه‌ها روی دیسک ذخیره می‌شوند و بعداً می‌توانید دوباره از آن‌ها برای اسکن استفاده کنید. مثلاً اول یک محدوده را با ICMP اسکن کنید و بعد فقط IPهایی را که جواب داده‌اند با TCP یا HTTP بررسی کنید.

bgscan برای توسعه‌دهنده‌ها و پژوهشگرهایی ساخته شده که می‌خواهند سریع کار کنند و برای اسکن‌کردن مجبور نباشند از ترمینال بیرون بروند.

<img width="1258" height="690" alt="bgscan" src="https://github.com/user-attachments/assets/998c2c7c-f960-4a71-a022-72d86b13c6fb" />

---

## قابلیت‌ها

**موتور اسکن**

- پروب‌های چندپروتکله: ICMP، TCP، HTTP/1.1، HTTP/2، HTTP/3 (QUIC)، TLS، DNS، DNSTT، Slipstream و Xray
- وصل‌کردن مراحل اسکن، مثلاً ICMP → TCP → HTTP، در حالت‌های Streaming و Batch
- چند Worker برای هر بخش، تا اسکن‌ها هم‌زمان انجام شوند
- به‌هم‌زدن ترتیب IPها و محدودکردن تعداد IPهایی که باید اسکن شوند
- پشتیبانی کامل از IPv4 و IPv6 در هدف‌ها و فهرست‌های IP

**رابط کاربری ترمینال**

- رابط متنی BubbleTea برای اجرای اسکن، دیدن وضعیت و مدیریت نتیجه‌ها
- داشبورد زنده با نوارهای پیشرفت تب‌بندی‌شده و جدول نتایج هر مرحله
- بخش تنظیمات داخل برنامه برای تغییر تنظیمات و ذخیرهٔ فوری آن‌ها
- تم‌های روشن، تیره و خودکار بر پایهٔ Catppuccin

**داده و ورودی/خروجی**

- ذخیرهٔ نتیجه‌ها در CSV و استفاده از آن‌ها برای اسکن‌های بعدی
- فهرست‌های IP آماده برای Cloudflare (IPv4 و IPv6)، AWS، Azure، Google، Akamai، Fastly، Bunny، G-Core و ایران
- مدیریت Outboundهای Xray؛ اضافه‌کردن آن‌ها از لینک اشتراک‌گذاری یا فایل JSON و تست اتصال و سرعتشان

**پایداری**

- ذخیره‌کردن نتیجه‌ها به شکلی امن، طوری که قطع‌شدن برنامه باعث خراب‌شدن فایل‌ها نشود و حافظهٔ زیادی مصرف نشود
- مدیریت خودکار لاگ‌های هسته، رابط کاربری و دیباگ؛ هر لاگ بعد از ۵۰ مگابایت چرخش پیدا می‌کند و ۳ نسخهٔ قبلی تا ۷ روز نگه داشته می‌شود
- بررسی تنظیمات و پیدا کردن برنامه‌های کمکی هنگام اجرا؛ اگر یکی از آن‌ها نصب نباشد، کل bgscan از کار نمی‌افتد و فقط همان نوع اسکن غیرفعال می‌شود

## چرا bgscan؟

- بیشتر اسکنرها فقط یک پروتکل را بلدند. bgscan می‌تواند چند پروتکل را پشت سر هم اجرا کند: یک محدوده را Ping کنید، IPهای جواب‌گو را با TCP بررسی کنید و بعد سراغ HTTP بروید.
- موتور اسکن به نوع اسکن کاری ندارد؛ IPها را می‌گیرد، نتیجه‌ها را جمع می‌کند و روی دیسک می‌نویسد. برای اضافه‌کردن یک اسکن جدید، کافی است رابط `Probe` را با چهار متد `Init`، `Run`، `Schema` و `Close` پیاده کنید و مرحلهٔ آن را به برنامه اضافه کنید.
- ICMP، TCP، HTTP و DNS به برنامهٔ دیگری نیاز ندارند. Xray، DNSTT و Slipstream جداگانه نصب می‌شوند و اگر وجود نداشته باشند فقط همان اسکن در دسترس نخواهد بود.
- تنظیمات در فایل‌های سادهٔ TOML هستند. می‌توانید آن‌ها را دستی تغییر دهید یا از بخش تنظیمات داخل برنامه استفاده کنید.
- bgscan برای ترمینال ساخته شده: حرکت با صفحه‌کلید، پنجره‌های ساده، نمایش زندهٔ پیشرفت و لاگ‌ها؛ بدون مرورگر و وب‌سرور.

## پروتکل‌های پشتیبانی‌شده

| پروتکل | لایه | توضیح |
| :---: | :---: | :--- |
| ICMP | 3 | شناسایی میزبان و بررسی دسترسی با Ping در IPv4 و IPv6 |
| TCP | 4 | اسکن اتصال و اعتبارسنجی TCP Handshake |
| HTTP | 7 | HTTP/1.1، HTTP/2 و HTTP/3 روی QUIC از طریق ALPN |
| TLS | 7 | بررسی نسخه‌های TLS 1.0 تا TLS 1.3 |
| DNS | 7 | پرس‌وجوی پیشرفتهٔ DNS با UDP، TCP و DNS-over-TLS، همراه با Fallback و بررسی Hijacking |
| DNSTT | 7 | اعتبارسنجی تونل DNS با SOCKS و بدون احراز هویت |
| Slipstream | 7 | اعتبارسنجی تونل Slipstream با SOCKS و بدون احراز هویت |
| Xray | 7 | اعتبارسنجی Outboundهای Xray و تست سرعت اتصال |

## نصب

### نصب سریع

لینوکس، macOS و Termux:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

ویندوز (PowerShell):

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

برای Android/Termux ابتدا وابستگی‌ها را نصب کنید:

```bash
pkg update -y && pkg install bash curl unzip -y
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

نصب‌کننده سیستم‌عامل را تشخیص می‌دهد، آخرین نسخه را دانلود می‌کند، آن را در پوشهٔ `bgscan/` باز می‌کند و برنامه را آمادهٔ اجرا می‌کند. اگر دوباره اجرا شود، نصب قبلی را پیدا می‌کند و از شما می‌پرسد که آن را جایگزین کند یا اول از آن پشتیبان بگیرد.

### نصب دستی

1. فایل ZIP مناسب سیستم‌عامل خود را از [صفحهٔ Releases](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.
2. آرشیو را استخراج کنید.
3. باینری را اجرا کنید:
   - لینوکس، macOS و Termux: `./bgscan`
   - ویندوز: `bgscan.exe`

در اولین اجرا، پوشهٔ `settings/` با فایل‌های تنظیمات پیش‌فرض TOML و پوشهٔ `ips/` با فهرست‌های IP آماده ساخته می‌شود.

### ساخت از سورس

bgscan برای دریافت وابستگی‌های مخصوص هر پلتفرم، از ابزار همراه `bgscan-builder` استفاده می‌کند. این وابستگی‌ها شامل باینری‌های Xray، DNSTT و Slipstream هستند.

```bash
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan

# نصب ابزار builder
# Linux / macOS:
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash
# Windows (PowerShell):
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# دریافت وابستگی‌های پلتفرم
# Linux / macOS:
./scripts/install-deps.sh
# Windows:
./scripts/install-deps.ps1

# ساخت و اجرا
go run ./cmd/bgscan/
```

برای ساخت Release یک پلتفرم مشخص:

```bash
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os darwin -arch arm64
```

> bgscan را نمی‌توان با `go install` نصب کرد، چون به باینری‌های خارجی و مخصوص پلتفرم نیاز دارد. برای نصب از ابزار builder یا اسکریپت نصب سریع استفاده کنید.

## شروع سریع

1. bgscan را از پوشهٔ نصب اجرا کنید (`./bgscan` در Unix و `bgscan.exe` در Windows).
2. گزینهٔ **Run Scan** را انتخاب و `Enter` را فشار دهید.
3. منبع هدف را انتخاب کنید: **IP List** از فهرست‌های واردشده یا **Result List** از نتیجهٔ اسکن قبلی.
4. نوع اسکن را انتخاب کنید: ICMP، TCP، HTTP، DNS یا Xray.
5. با فشردن `Enter` اسکن را شروع کنید. پیشرفت و نتایج به‌صورت زنده در داشبورد نمایش داده می‌شوند.
6. از منوی اصلی وارد **Result Files** شوید تا نتایج ذخیره‌شده را ببینید، تغییر نام دهید یا حذف کنید.

| کلید | عملکرد |
| :---: | :--- |
| `↑` `↓` | جابه‌جایی بین گزینه‌ها |
| `Enter` | انتخاب یا شروع |
| `b` یا `Esc` | بازگشت |
| `q` | خروج |

## مستندات

- صفحهٔ اصلی: <https://mohsenbg.github.io/bgscan>
- مستندات کامل: <https://mohsenbg.github.io/bgscan/docs>

مستندات شامل راه‌اندازی سریع، انواع اسکن، منابع اسکن، فهرست‌های IP، فایل‌های نتیجه، Pipeline اسکن، Outboundهای Xray، تمام فایل‌های تنظیمات TOML، لاگ‌ها و راهنمای توسعه است.

## تنظیمات

تمام تنظیمات به‌صورت فایل‌های TOML در پوشهٔ `settings/` و در کنار باینری ذخیره می‌شوند:

| فایل | کاربرد |
| :--- | :--- |
| `general_settings.toml` | حالت Pipeline، اندازهٔ Batch، سقف IP هر مرحله، توقف پس از یافتن نتیجه، Shuffle و فاصلهٔ نمایش وضعیت |
| `writer_settings.toml` | Buffer نتیجه، فاصلهٔ Flush، اندازهٔ Channel و Batch و پوشهٔ نتایج |
| `icmp_settings.toml` | Timeout، تعداد Retry و Workerهای ICMP |
| `tcp_settings.toml` | پورت، Timeout، Retry و Workerهای TCP |
| `http_settings.toml` | نسخهٔ HTTP/HTTPS/HTTP3، محدودهٔ TLS و کدهای وضعیت قابل قبول |
| `dns_settings.toml` | تنظیمات Resolver، DNSTT و Slipstream |
| `xray_settings.toml` | نوع تست اتصال Xray، تست سرعت و Pre-scan |

می‌توانید فایل‌ها را دستی تغییر دهید یا از بخش تنظیمات داخل برنامه استفاده کنید. هر دو روش همان فایل‌ها را تغییر می‌دهند و تغییرات بخش تنظیمات فوراً ذخیره می‌شوند. توضیح همهٔ گزینه‌ها در [مستندات تنظیمات](https://mohsenbg.github.io/bgscan/docs/settings/) آمده است.

## نمونهٔ استفاده

**اسکن یک فهرست IP آماده:** پس از اجرای برنامه، **Run Scan → IP List → cloudflare_IPv4** را انتخاب کنید؛ برای IPv6 از `cloudflare_IPv6` استفاده کنید. سپس نوع اسکن، مثلاً `t` برای TCP، را انتخاب کنید. نتایج در `result/tcp/` یا پوشه‌ای که در `result_directory` تعیین کرده‌اید ذخیره می‌شوند.

**زنجیرهٔ ICMP → TCP → HTTP:** در `general_settings.toml` مقدار `pipeline_mode` را روی `"streaming"` بگذارید. وقتی چند نوع اسکن فعال باشد، مراحل به‌صورت خودکار زنجیره می‌شوند و فقط IPهایی که معیار موفقیت مرحلهٔ قبلی را داشته باشند به مرحلهٔ بعد می‌روند.

**اسکن مجدد نتیجهٔ قبلی:** از مسیر **Run Scan → Result List** یک فایل نتیجه را انتخاب کنید. bgscan فقط IPهای همان فایل را دوباره بررسی می‌کند؛ این روش برای تحلیل عمیق‌تر میزبان‌هایی مناسب است که قبلاً از یک مرحله عبور کرده‌اند.

**اعتبارسنجی Outboundهای Xray:** از منوی اصلی وارد **Xray → Outbounds** شوید و با فشردن `a` یک قالب را از لینک‌هایی مانند `vless://`، `vmess://`، `trojan://`، `ss://`، `hysteria2://` و `wireguard://` یا از فایل JSON اضافه کنید. سپس اسکن Xray را اجرا کنید تا اتصال و پهنای باند بررسی شود.

## پلتفرم‌های پشتیبانی‌شده

| پلتفرم | معماری‌ها |
| :--- | :--- |
| Linux | amd64، arm64، arm32، 386 |
| Windows | amd64، arm64 (نسخهٔ 10 به بعد) |
| macOS | amd64، arm64 |
| Android (Termux) | arm64، arm32، x86_64، x86 |

> برای Termux از نسخهٔ F-Droid استفاده کنید؛ نسخهٔ موجود در Play Store قدیمی است.

## ساختار پروژه

```text
bgscan/
├── cmd/bgscan/              # نقطهٔ ورود برنامه
├── internal/
│   ├── core/
│   │   ├── config/          # تنظیمات TOML، مقادیر پیش‌فرض و اعتبارسنجی
│   │   │   └── validate/    # اعتبارسنج‌های بخش‌ها و اعتبارسنجی کلی
│   │   ├── scanner/         # هماهنگ‌کنندهٔ اسکنر و Stage Builderها
│   │   │   ├── engine/      # اجرای Pipeline در حالت Streaming و Batch
│   │   │   ├── portmgr/     # انتخاب پورت برای مراحل TCP و HTTP
│   │   │   └── probe/       # رابط Probe و پیاده‌سازی پروتکل‌ها
│   │   ├── result/          # رابط نتیجه، Schema، نویسنده و بارگذاری CSV
│   │   ├── iplist/          # بارگذاری، تجزیه، ثبت و Shuffle فهرست‌های IP
│   │   ├── netutil/         # ابزارهای netip و محاسبات CIDR برای IPv4/IPv6
│   │   ├── dns/             # ابزار DNS، DNSTT، Slipstream و SOCKS5
│   │   ├── speedtest/       # تست پهنای باند، تأخیر و Transport در Xray
│   │   ├── xray/             # اجرای Xray و تجزیهٔ Outbound و لینک
│   │   ├── process/         # مدیریت چرخهٔ عمر Process در چند پلتفرم
│   │   └── fileutil/        # ابزار CSV، JSON، TOML، متن و فایل موقت
│   ├── logger/              # لاگ‌گیری سطح‌بندی‌شده با چرخش lumberjack
│   ├── startup/             # بررسی سلامت logger، config، xray، dnstt و slipstream
│   └── ui/                  # TUI مبتنی بر BubbleTea
├── assets/                  # باینری‌های Xray، DNSTT و Slipstream و قالب‌های Outbound
├── ips/                     # فهرست‌های آماده و واردشدهٔ IP در قالب CSV
├── settings/                # فایل‌های پیش‌فرض تنظیمات TOML
├── result/                  # خروجی اسکن، تفکیک‌شده بر اساس نوع اسکن
├── logs/                    # فایل‌های core.log، ui.log و debug.log
├── scripts/                 # اسکریپت‌های نصب، ساخت، وابستگی و انتشار
├── docs/                    # سایت مستندات ساخته‌شده با Hugo Book
└── go.mod
```

## کمک به توسعهٔ پروژه

اگر می‌خواهید در پروژه کمک کنید، می‌توانید باگ‌ها را برطرف کنید، مستندات را بهتر کنید یا قابلیت جدیدی اضافه کنید.

- از `main` شاخه بگیرید و نامی توصیفی مانند `feature/`، `fix/`، `docs/` یا `refactor/` انتخاب کنید.
- سبک فعلی پروژه را حفظ کنید، کد را ساده و خوانا بنویسید و تغییرات را در کامیت‌های کوچک نگه دارید.
- برای پیام کامیت از پیشوندهایی مثل `feat:`، `fix:`، `docs:`، `refactor:` و `test:` استفاده کنید.
- قبل از فرستادن Pull Request، مطمئن شوید پروژه ساخته می‌شود، تست‌ها پاس می‌شوند و مستندات هم به‌روز هستند.
- اگر تغییر بزرگی دارید، اول یک Issue باز کنید تا دربارهٔ روش انجام آن صحبت کنیم.

راهنمای کامل در [راهنمای مشارکت](https://mohsenbg.github.io/bgscan/docs/developer/contributing/) قرار دارد.

با فرستادن کد یا مستندات برای bgscan، قبول می‌کنید که این تغییرات می‌توانند همراه پروژه و تحت مجوز آن منتشر شوند. همچنین تأیید می‌کنید که حق فرستادن آن‌ها را دارید.

## مجوز

این پروژه تحت [مجوز MIT](LICENSE) منتشر می‌شود — Copyright (c) 2026 Mohsen Bagheri

## حمایت مالی

اگر bgscan برایتان مفید بوده است، می‌توانید از توسعهٔ آن حمایت کنید:

| شبکه | ارز | آدرس |
| :---: | :---: | :--- |
| Bitcoin | `BTC` | `bc1qdwh57dm97nmx5jzdr7lrc9cxe5xh3zc59er7z9` |
| Ethereum | `ETH` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| BNB Smart Chain | `BNB / BEP20` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| TRON | `TRX / TRC20` | `TNW6pbfY8zZVZezZWyYXo7h12MycRsVJK7` |

<div align="center">

ساخته‌شده با Go · مجوز MIT

</div>