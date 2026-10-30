<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="./images/logo-light.svg">
  <img src="./images/logo-light.svg" alt="BGSCAN" width="520" style="max-width:100%;">
</picture>

اسکنر bgscan برای اسکن IP با چند پروتکل و قابلیت زنجیرکردن مراحل اسکن

[English](./README.md) | [فارسی](./README.fa.md)

---

[نصب](#نصب) · [شروع سریع](#شروع-سریع) · [مستندات](#مستندات) · [پروتکل‌های پشتیبانی‌شده](#پروتکل‌های-پشتیبانی‌شده)

</div>

## پروژهٔ bgscan چیه؟

**این برنامه (bgscan)** یه اسکنر شبکه‌ست که با Go نوشته شده و داخل ترمینال اجرا می‌شه. با پروتکل‌هایی مثل ICMP، TCP، HTTP، DNS، Xray و تونل‌های DNS (مثل DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS و TheFeed) آدرس‌های IP رو بررسی می‌کنه.

هر نوع اسکن رو می‌تونید جدا اجرا کنید یا چند مرحله رو به هم وصل کنید و نتیجهٔ هر مرحله رو به مرحلهٔ بعد بدید. همه‌چیز هم از داخل یه رابط متنی و با کیبورد قابل مدیریته.

یکی از کاربردهای اصلی این برنامه همین زنجیرکردن مرحله‌هاست. مثلاً می‌تونید اول یه محدوده رو با ICMP اسکن کنید، بعد IPهایی که جواب دادن رو برای بررسی TCP بفرستید و در نهایت اتصال HTTP رو روی اون‌ها تست کنید. نتیجهٔ هر مرحله هم ذخیره می‌شه و می‌تونید توی اسکن‌های بعدی دوباره ازش استفاده کنید. به این شکل، با هر مرحله تعداد IPها کمتر می‌شه و لازم نیست همهٔ آدرس‌ها رو از اول تا آخر با تمام پروتکل‌ها بررسی کنید.


<div align="center">
  <video
    src="https://github.com/user-attachments/assets/4a4c216e-0179-427c-8fe1-89a8fa197730"
    autoplay
    muted
    loop
    playsinline
    controls
    width="100%"
    style="max-width: 900px; border-radius: 12px; border: 1px solid #30363d;"
  ></video>
</div>

---

## قابلیت‌ها

**موتور اسکن**

- پشتیبانی از چندین پروتکل: ICMP، TCP، HTTP/1.1، HTTP/2، HTTP/3 (QUIC)، DNS، DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS، TheFeed و Xray
- امکان زنجیرکردن مرحله‌های اسکن، مثل ICMP → TCP → HTTP، در حالت‌های Streaming و Batch
- اجرای هم‌زمان اسکن‌ها با چند Worker برای هر پروب
- امکان Shuffle کردن ترتیب IPها و تعیین سقف تعداد IPهای قابل اسکن
- پشتیبانی از IPv4 و IPv6 در هدف‌ها و فهرست‌های IP

**رابط کاربری ترمینال**

- رابط متنی مبتنی بر BubbleTea که تمام بخش‌هاش با کیبورد قابل مدیریته
- داشبورد زنده با نوار پیشرفت جداگانه و جدول نتیجه‌های هر مرحله
- تنظیمات داخل برنامه برای تغییر گزینه‌ها و ذخیرهٔ فوری تغییرات روی دیسک
- تم روشن، تیره و خودکار بر پایهٔ Catppuccin

**داده و ورودی/خروجی**

- ذخیرهٔ نتیجه‌ها در CSV و امکان استفادهٔ مجدد از اون‌ها به‌عنوان ورودی اسکن
- فهرست‌های IP آماده برای Cloudflare (IPv4 و IPv6)، AWS، Azure، Google، Akamai، Fastly، Bunny، G-Core، ایران و DNSهای عمومی
- مدیریت Outboundهای Xray؛ می‌تونید از لینک اشتراک یا فایل JSON اضافه‌شون کنید و اتصال و سرعتشون رو تست کنید

**پایداری**

- ذخیرهٔ امن نتیجه‌ها با ادغام اتمیک، `fsync` و جابه‌جایی فایل موقت؛ مصرف حافظه هم در طول اجرا کنترل می‌شه
- لاگ‌های core، UI و debug به‌صورت خودکار Rotate می‌شن. هر فایل بعد از ۵۰ مگابایت چرخش پیدا می‌کنه، ۳ نسخهٔ قبلی نگه داشته می‌شه و لاگ‌ها تا ۷ روز باقی می‌مونن
- تنظیمات و برنامه‌های کمکی هنگام اجرا بررسی می‌شن. اگه وابستگی‌ای وجود نداشته باشه، برنامه متوقف نمی‌شه؛ فقط هشدار می‌ده و همون نوع اسکن غیرفعال می‌مونه

## چرا از bgscan استفاده کنیم؟

تفاوت اصلی این برنامه با خیلی از اسکنرهای دیگه، امکان وصل‌کردن چند نوع اسکن به همدیگه‌ست. مثلاً می‌تونید یه محدودهٔ IP رو Ping کنید، IPهایی که جواب دادن رو برای بررسی TCP بفرستید و در مرحلهٔ بعد پورت‌های باز رو با HTTP تست کنید. همهٔ این مراحل توی یه اجرا انجام می‌شن.

موتور اسکن وابستگی خاصی به نوع پروتکل نداره. IPها رو بین مرحله‌ها منتقل می‌کنه، نتیجه‌ها رو جمع می‌کنه و روی دیسک می‌نویسه. برای اضافه‌کردن یه نوع اسکن جدید هم کافیه رابط `Probe` رو با چهار متد `Init`، `Run`، `Schema` و `Close` پیاده‌سازی کنید و مرحلهٔ جدید رو ثبت کنید.

از نظر وابستگی هم بیشتر قابلیت‌ها داخل خود برنامه اجرا می‌شن. ICMP، TCP، HTTP و DNS از کتابخانه‌های Go استفاده می‌کنن. تنها باینری خارجی مربوط به Slipstream هست که موقع اجرا بررسی می‌شه. اگه وجود نداشته باشه، فقط اسکن Slipstream غیرفعال می‌شه و بقیهٔ برنامه به کارش ادامه می‌ده. Xray و تونل‌های DNSTT، VayDNS، MasterDNS، StormDNS و TheFeed هم داخل همون پروسه و با کتابخانه‌های داخلی اجرا می‌شن.

تنظیمات برنامه توی فایل‌های سادهٔ TOML نگه‌داری می‌شن. بخش تنظیمات داخل TUI هم مستقیماً همین فایل‌ها رو می‌خونه و تغییر می‌ده؛ یعنی فرقی نمی‌کنه تنظیمات رو دستی عوض کنید یا از داخل برنامه، در نهایت هر دو روش روی یه فایل کار می‌کنن.

کل رابط کاربری هم داخل ترمیناله. از جابه‌جایی بین بخش‌ها و پنجره‌های شناور گرفته تا نمایش زندهٔ نتیجه‌ها و لاگ‌ها. خبری از وب‌سرور یا نیاز به بازکردن مرورگر نیست.

## پروتکل‌های پشتیبانی‌شده

| پروتکل | لایه | توضیح |
| :---: | :---: | :--- |
| ICMP | 3 | شناسایی میزبان و بررسی دسترسی با Ping در IPv4 و IPv6 |
| TCP | 4 | بررسی اتصال و اعتبارسنجی TCP Handshake |
| HTTP | 7 | پشتیبانی از HTTP/1.1، HTTP/2 و HTTP/3 روی QUIC از طریق ALPN |
| DNS | 7 | پرس‌وجوی DNS با UDP، TCP و DNS-over-TLS، همراه با Fallback و بررسی Hijacking |
| DNSTT | 7 | اعتبارسنجی تونل DNS با اتصال پایهٔ SOCKS و SSH |
| VayDNS | 7 | اعتبارسنجی تونل DNS با اتصال پایهٔ SOCKS و SSH |
| Slipstream | 7 | اعتبارسنجی تونل DNS با اتصال پایهٔ SOCKS و SSH |
| MasterDNS | 7 | اعتبارسنجی تونل DNS با رمزنگاری قابل تنظیم و کشف MTU |
| StormDNS | 7 | اعتبارسنجی تونل DNS با رمزنگاری قابل تنظیم و کشف MTU |
| TheFeed | 7 | اعتبارسنجی تونل فید رمزنگاری‌شدهٔ DNS روی UDP، TCP یا DNS-over-TLS |
| Xray | 7 | اعتبارسنجی Outboundهای Xray و تست سرعت اتصال |

## نصب

### نصب سریع

**لینوکس، macOS :**

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

**ویندوز (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**اندروید (Termux):**

```bash
{ command -v curl >/dev/null 2>&1 || pkg install -y curl; } && curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

نصب‌کننده، ابزار `bgscan-installer` رو دانلود می‌کنه. این ابزار با Rust نوشته شده، مستقل از bgscan منتشر می‌شه ([bgscan-installer](https://github.com/MohsenBg/bgscan-installer)) و وظیفه‌ش دانلود و نصب نسخهٔ مناسب سیستم‌عامل شماست. قبل از نصب هم چک‌سام SHA-256 فایل رو بررسی می‌کنه.

برنامه داخل پوشهٔ `bgscan/` نصب می‌شه. اگه نصب‌کننده رو دوباره اجرا کنید، نصب قبلی رو تشخیص می‌ده و چهار گزینه در اختیارتون می‌ذاره:

- **Update:** به‌روزرسانی در همون مسیر. باینری جایگزین می‌شه و فایل‌های جدید اضافه می‌شن، ولی پوشه‌های `ips`، `assets` و `settings` دست‌نخورده می‌مونن.
- **Clean Install:** نصب تمیز.
- **Backup:** انتقال نصب قبلی به پوشه‌ای با مهر زمانی، مثل `bgscan_<timestamp>`.
- **Cancel:** انصراف از عملیات.

برای نصب یه نسخهٔ مشخص به‌جای آخرین نسخه:

```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh -s -- --version v2.11.0
```

### نصب دستی

1. فایل ZIP مربوط به سیستم‌عامل خودتون رو از [صفحهٔ Releases](https://github.com/MohsenBg/bgscan/releases/latest) دانلود کنید.
2. آرشیو رو از حالت فشرده خارج کنید.
3. باینری رو اجرا کنید:
   - لینوکس، macOS و Termux: `./bgscan`
   - ویندوز: `bgscan.exe`

در اولین اجرا، پوشهٔ `settings/` با فایل‌های پیش‌فرض TOML و پوشهٔ `ips/` با فهرست‌های آمادهٔ IP ساخته می‌شن.

### ساخت از سورس

برای ساخت bgscan، دریافت باینری Slipstream مخصوص هر پلتفرم و آماده‌کردن فهرست‌های IP پیش‌فرض، از ابزار همراه `bgscan-builder` استفاده می‌شه.

ابزار `bgscan-builder` فقط برای ساخت برنامه‌ست. نصب و به‌روزرسانی نسخه‌های منتشرشده با ابزار جداگانهٔ [`bgscan-installer`](https://github.com/MohsenBg/bgscan-installer) انجام می‌شه.

خود Xray هم نیازی به باینری جداگانه نداره و داخل برنامه اجرا می‌شه.

```bash
git clone https://github.com/MohsenBg/bgscan.git
cd bgscan

# Install builder
# Linux / macOS:
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.sh | bash

# Windows (PowerShell):
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install-builder.ps1 | iex

# Install platform dependencies
# Linux / macOS:
./scripts/install-deps.sh

# Windows:
./scripts/install-deps.ps1

# Build and run
go run ./cmd/bgscan/
```

برای ساخت Release مخصوص یه پلتفرم (لیست کامل فلگ‌ها با `bgscan-builder release -h`):

```bash
./bgscan-builder release -os linux -arch amd64
./bgscan-builder release -os windows -arch amd64
./bgscan-builder release -os macos -arch arm64
./bgscan-builder release -os android -arch arm64 -ndk-dir /opt/android-ndk
```

> نصب bgscan با `go install` ممکن نیست، چون برنامه به باینری Slipstream مخصوص هر پلتفرم نیاز داره. برای نصب از اسکریپت نصب سریع یا ابزار builder استفاده کنید.

## شروع سریع

1. از پوشهٔ نصب، bgscan رو اجرا کنید (`./bgscan` در Unix و `bgscan.exe` در Windows).
2. گزینهٔ **Run Scan** رو انتخاب کنید و `Enter` بزنید.
3. ورودی اسکن رو انتخاب کنید:
   - منو **IP List:** استفاده از فهرست‌های IP واردشده
   -  منو **Result List:** استفاده از نتیجهٔ اسکن قبلی
4. نوع اسکن رو انتخاب کنید؛ مثل DNS Resolver، تونل DNS، ICMP، TCP، HTTP یا Xray.
5. با زدن `Enter` اسکن شروع می‌شه و نتیجه‌ها و میزان پیشرفت رو می‌تونید همون لحظه توی داشبورد ببینید.
6. برای دیدن نتیجه‌های ذخیره‌شده، از منوی اصلی وارد **Result Files** بشید. از همین بخش می‌تونید فایل‌ها رو ببینید، تغییر نام بدید یا حذف کنید.

| کلید | عملکرد |
| :---: | :--- |
| `↑` `↓` | جابه‌جایی بین گزینه‌ها |
| `Enter` | انتخاب یا شروع |
| `b` یا `Esc` | بازگشت |
| `q` | خروج |

## مستندات

برای دسترسی به مستندات کامل پروژه:

- صفحهٔ اصلی: [https://mohsenbg.github.io/bgscan](https://mohsenbg.github.io/bgscan)
- مستندات: [https://mohsenbg.github.io/bgscan/docs](https://mohsenbg.github.io/bgscan/docs)

توی مستندات این بخش‌ها رو پیدا می‌کنید:

- **شروع سریع:** نصب، اجرا و انجام اولین اسکن
- **اسکنر:** انواع اسکن، فهرست‌های IP، فایل‌های نتیجه، Pipeline و Outboundهای Xray
- **تونل DNS:** راهنمای جداگانهٔ DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS و TheFeed، همراه با روند اسکن و پیش‌فرض‌های هر پلتفرم
- **تنظیمات:** توضیح فایل‌های TOML شامل general، writer، ICMP، TCP، HTTP، DNS و Xray، به‌همراه بخش تنظیمات داخل برنامه
- **لاگ‌ها:** انواع لاگ، نمایشگر لاگ و نحوهٔ چرخش فایل‌ها
- **توسعه:** معماری پروژه، بستهٔ core، موتور اسکن، پروب‌ها، تنظیمات، نتیجه‌ها، رابط کاربری، کامپوننت‌ها، تم‌ها و راهنمای مشارکت

## تنظیمات

تنظیمات bgscan داخل پوشهٔ `settings/` و در قالب فایل‌های TOML قرار دارن. این پوشه کنار باینری ساخته می‌شه.

| فایل | کاربرد |
| :--- | :--- |
| `general_settings.toml` | حالت Pipeline، اندازهٔ Batch، سقف IP هر مرحله، سقف IPهای موفق، Shuffle و فاصلهٔ نمایش وضعیت |
| `writer_settings.toml` | Buffer نتیجه‌ها، فاصلهٔ Flush، اندازهٔ Channel و Batch و مسیر ذخیرهٔ نتیجه‌ها |
| `icmp_settings.toml` | Timeout، تعداد Retry و Workerهای ICMP |
| `tcp_settings.toml` | پورت، Timeout، Retry و Workerهای TCP |
| `http_settings.toml` | نسخهٔ HTTP/HTTPS/HTTP3، محدودهٔ TLS، Fingerprint و کدهای وضعیت قابل قبول |
| `dns_settings.toml` | تنظیمات Resolver و هماهنگی اسکن تونل DNS |
| `xray_settings.toml` | نوع تست اتصال Xray، Timeout تست سرعت، سقف Dialهای هم‌زمان و Pre-scan |

برای تغییر تنظیمات می‌تونید فایل‌ها رو مستقیماً ویرایش کنید یا از بخش تنظیمات داخل برنامه استفاده کنید. هر دو روش روی همین فایل‌ها کار می‌کنن و تغییراتی که از داخل TUI انجام می‌دید همون لحظه ذخیره می‌شن؛ نیازی هم به ری‌استارت برنامه نیست.

توضیح کامل گزینه‌ها رو می‌تونید توی [مستندات تنظیمات](https://mohsenbg.github.io/bgscan/docs/settings/) ببینید.

## نمونهٔ استفاده

**اسکن یه فهرست IP آماده**

برای شروع، برنامه رو اجرا کنید و از مسیر **Run Scan → IP List → cloudflare_IPv4** فهرست Cloudflare رو انتخاب کنید. برای IPv6 هم می‌تونید از `cloudflare_IPv6` استفاده کنید.

بعد نوع اسکن، مثلاً `t` برای TCP، رو انتخاب کنید. اسکن با تعداد Workerهای تنظیم‌شده انجام می‌شه و نتیجه‌ها داخل `result/tcp/` یا مسیری که برای `result_directory` مشخص کردید ذخیره می‌شن.

**زنجیرهٔ ICMP → TCP → HTTP**

برای استفاده از این حالت، توی `general_settings.toml` مقدار `pipeline_mode` رو روی `"streaming"` بذارید.

اگه بیشتر از یه نوع اسکن فعال باشه، مرحله‌ها به‌ترتیب اجرا می‌شن و فقط IPهایی که مرحلهٔ قبلی رو با موفقیت رد کردن، به مرحلهٔ بعد می‌رن.

**اسکن مجدد نتیجهٔ قبلی**

برای این کار از مسیر **Run Scan → Result List** یکی از فایل‌های نتیجهٔ ذخیره‌شده رو انتخاب کنید.

این بار bgscan فقط IPهای داخل همون فایل رو بررسی می‌کنه. این روش وقتی به درد می‌خوره که بخواید نتیجه‌های یه مرحله رو با پروتکل دیگه‌ای دوباره بررسی کنید.

**اعتبارسنجی Outboundهای Xray**

برای اضافه‌کردن Outbound، از منوی اصلی وارد **Xray Outbounds** بشید و کلید `a` رو بزنید. می‌تونید Outbound رو از لینک اشتراک یا فایل JSON اضافه کنید.

فرمت‌های پشتیبانی‌شده برای لینک اشتراک:

`vless://`، `vmess://`، `trojan://`، `ss://`، `hysteria2://` و `wireguard://`

بعد از اضافه‌کردن Outboundها، یه اسکن Xray اجرا کنید تا اتصال و پهنای باندشون بررسی بشه.

**تست تونل‌های DNS**

برای ساخت کانفیگ، از منوی اصلی وارد **DNS Tunneling** بشید و یکی از پروتکل‌های DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS یا TheFeed رو انتخاب کنید.

بعد اسکن تونل DNS رو اجرا کنید تا مشخص بشه کدوم Resolverها امکان برقراری تونل رو دارن.

## پلتفرم‌های پشتیبانی‌شده

| پلتفرم | معماری‌ها |
| :--- | :--- |
| Linux | amd64، arm64، arm32، 386 |
| Windows | amd64، arm64 (نسخهٔ 10 به بعد) |
| macOS | amd64، arm64 |
| Android (Termux) | arm64، arm32، x86_64، x86 |

> برای استفاده از bgscan در Termux، نسخهٔ F-Droid رو نصب کنید. نسخهٔ Play Store قدیمیه.

## ساختار پروژه

```text
bgscan/
├── cmd/bgscan/              # نقطهٔ ورود برنامه
├── internal/
│   ├── core/
│   │   ├── config/          # تنظیمات TOML، مقادیر پیش‌فرض و اعتبارسنجی
│   │   │   └── validate/    # اعتبارسنجی بخش‌ها و تنظیمات کلی
│   │   ├── scanner/         # هماهنگ‌کنندهٔ اسکنر و Stage Builderها
│   │   │   ├── engine/      # اجرای Pipeline در حالت Streaming و Batch
│   │   │   ├── portmgr/     # اجارهٔ پورت محلی برای پروب‌های تونل و Xray
│   │   │   └── probe/       # رابط Probe و پیاده‌سازی پروتکل‌ها
│   │   ├── result/          # رابط نتیجه، Schema، نویسنده و بارگذاری CSV
│   │   ├── iplist/          # بارگذاری، تجزیه، ثبت و Shuffle فهرست‌های IP
│   │   ├── netutil/         # ابزارهای netip و محاسبات CIDR برای IPv4/IPv6
│   │   ├── dns/             # ابزار DNS و سرویس‌های Config تونل‌ها، SOCKS5 و SSH
│   │   ├── speedtest/       # تست پهنای باند، تأخیر و Transport در Xray
│   │   ├── xray/            # سرویس داخلی Xray و تجزیهٔ Outboundها و لینک‌ها
│   │   ├── process/         # مدیریت Process جانبی Slipstream در چند پلتفرم
│   │   └── fileutil/        # ابزار CSV، JSON، TOML، متن و فایل موقت
│   ├── logger/              # لاگ‌گیری سطح‌بندی‌شده با چرخش lumberjack
│   └── ui/                  # رابط TUI مبتنی بر BubbleTea
├── assets/                  # باینری Slipstream، Configها و داده‌های Xray و تونل DNS
├── ips/                     # فهرست‌های آماده و واردشدهٔ IP در قالب CSV
├── settings/                # فایل‌های پیش‌فرض تنظیمات TOML
├── result/                  # خروجی اسکن، تفکیک‌شده بر اساس نوع اسکن
├── logs/                    # فایل‌های core.log، ui.log و debug.log
├── scripts/                 # اسکریپت‌های نصب، ساخت، وابستگی و انتشار
├── docs/                    # سایت مستندات ساخته‌شده با Hugo Book
└── go.mod
```

## کمک به توسعهٔ پروژه

برای مشارکت توی bgscan لازم نیست حتماً قابلیت بزرگی اضافه کنید. رفع باگ، بهترکردن مستندات و اضافه‌کردن تست هم کمک می‌کنه.

- از شاخهٔ `main` یه شاخهٔ جدید با اسم مشخص بسازید؛ مثلاً `feature/`، `fix/`، `docs/` یا `refactor/`.
- سعی کنید سبک فعلی پروژه رو حفظ کنید؛ کد ساده و خوانا باشه و تغییرات رو توی کامیت‌های کوچیک نگه دارید.
- برای پیام کامیت از پیشوندهایی مثل `feat:`، `fix:`، `docs:`، `refactor:` و `test:` استفاده کنید.
- قبل از ارسال Pull Request، مطمئن بشید پروژه بدون مشکل ساخته می‌شه، تست‌ها پاس می‌شن و مستندات مربوط به تغییرات به‌روز شدن.
- اگه تغییر بزرگی مدنظرتونه، بهتره اول یه Issue باز کنید تا دربارهٔ روش انجامش هماهنگ بشیم.

جزئیات بیشتر توی [راهنمای مشارکت](https://mohsenbg.github.io/bgscan/docs/developer/contributing/) قرار داره.

با ارسال کد یا مستندات برای bgscan، قبول می‌کنید که تغییرات شما می‌تونن همراه پروژه و تحت مجوز خود پروژه منتشر بشن. همچنین تأیید می‌کنید که حق ارسال اون تغییرات رو دارید.

## مجوز

این پروژه تحت [مجوز GPL-3.0](LICENSE) منتشر می‌شه.

Copyright (c) 2026 Mohsen‌‌Bg

## حمایت مالی

اگه bgscan براتون مفید بوده و دوست دارید از توسعه‌ش حمایت کنید، می‌تونید از آدرس‌های زیر استفاده کنید:

| شبکه | ارز | آدرس |
| :---: | :---: | :--- |
| Bitcoin | `BTC` | `bc1qdwh57dm97nmx5jzdr7lrc9cxe5xh3zc59er7z9` |
| Ethereum | `ETH` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| BNB Smart Chain | `BNB / BEP20` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| TRON | `TRX / TRC20` | `TNW6pbfY8zZVZezZWyYXo7h12MycRsVJK7` |

<div align="center">

ساخته‌شده با Go · مجوز GPL-3.0

</div>

