---
title: "Outboundهای Xray"
weight: 6
---

# Outboundهای Xray

هر Outbound مشخص می‌کنه BGScan موقع اسکن Xray چطور ترافیک رو از پروکسی عبور بده. هر Outbound به‌صورت یه فایل JSON در پوشه `assets/xray/outbounds/` ذخیره می‌شه و پروتکل، ترنسپورت و تنظیمات امنیتی پروکسی رو مشخص می‌کنه.

برای اجرای اسکن Xray، باید حداقل یه Outbound داشته باشید.

برای مدیریت Outboundها، از منوی اصلی وارد **Xray Outbound** بشید یا کلید `x` رو بزنید.

---

{{< img "/bgscan-outbound-menu.webp" "bgscan outbound menu" >}}

## فهرست Outboundها

جدول، Outboundهای موجود رو از جدیدترین به قدیمی‌ترین نمایش می‌ده:

| ستون | توضیح |
|---|---|
| Name | نام Outbound، بدون پسوند `.json` |
| Protocol | پروتکل پروکسی، مثل `vless`، `vmess` یا `socks` |
| Network | نوع ترنسپورت، مثل `ws`، `grpc`، `xhttp` یا `tcp` |
| Security | مقدار `streamSettings.security`، مثل `tls`، `reality` یا مقدار خالی |
| Created | زمان آخرین تغییر فایل |

برای مدیریت Outboundها از این کلیدها استفاده کنید:

| کلید | عملکرد |
|---|---|
| `a` | افزودن Outbound جدید |
| `r` | تغییر نام Outbound انتخاب‌شده |
| `x` | حذف Outbound انتخاب‌شده |

کلیدهای مرتب‌سازی و فیلتر هم در دسترس‌اند:

- `s`: تغییر ستون مرتب‌سازی
- `o`: برعکس کردن ترتیب مرتب‌سازی
- `/`: باز کردن فیلتر
- `b` یا `Esc`: برگشت به صفحه قبل

---

{{< img "/bgscan-outbound-empty.webp" "bgscan outbound empty" >}}

## افزودن Outbound با کلید `a`

با زدن `a`، منوی **Addition Method** باز می‌شه. برای افزودن Outbound دو راه دارید.

---

### روش اول: From Link

گزینه **From Link** رو انتخاب کنید، لینک پروکسی رو وارد کنید و بعد یه اسم برای Outbound بذارید. BGScan لینک رو تجزیه می‌کنه، اطلاعات Outbound رو ازش استخراج می‌کنه و به‌صورت یه فایل JSON ذخیره می‌کنه.

{{< img "/bgscan-outbound-select-link.webp" "bgscan outbound select link" >}}

لینک‌های زیر پشتیبانی می‌شن:

| شِمای لینک | پروتکل |
|---|---|
| `vmess://` | VMess |
| `vless://` | VLESS |
| `trojan://` | Trojan |
| `ss://` | Shadowsocks |
| `socks://` | SOCKS |
| `hysteria2://` یا `hy2://` | Hysteria2 |

لینک‌های اشتراک WireGuard پشتیبانی نمی‌شن، اما می‌تونید Outbound مربوط به اون رو از طریق فایل JSON وارد کنید.

بعد از وارد کردن لینک، اسم Outbound رو مشخص کنید. این اسم باید یکتا باشه و نباید قبلاً برای Outbound دیگه‌ای استفاده شده باشه.

```text
a → From Link → Paste link → Enter name → Done
```

{{< img "/bgscan-outbound-link.webp" "bgscan outbound link" >}}
{{< img "/bgscan-outbound-name.webp" "bgscan outbound name" >}}

---

### روش دوم: From File

گزینه **From File** رو انتخاب کنید و یه فایل JSON یا JSONC از سیستم‌تون باز کنید. BGScan، Outbound پروکسی رو از فایل استخراج می‌کنه، اعتبارش رو بررسی می‌کنه و ذخیره می‌کنه.

{{< img "/bgscan-outbound-select-json.webp" "bgscan outbound select json" >}}

فایل ورودی می‌تونه یکی از این حالت‌ها باشه:

- **یک Outbound تکی:** مثلاً یه شیء JSON که فیلد `protocol` داره.
- **آرایه‌ای از Outboundها:** شامل چند Outbound.
- **Config کامل Xray:** فایلی که آرایه `outbounds` رو داره.

اگه فایل شامل چند Outbound باشه، BGScan اولین Outboundی رو انتخاب می‌کنه که به‌عنوان پروکسی شناخته بشه. اگه هیچ‌کدوم به این شکل تشخیص داده نشن، اولین Outbound با تگ `proxy` انتخاب می‌شه. بقیه تنظیمات Config کامل ذخیره نمی‌شن.

**شرایط فایل ورودی:**

- فایل باید JSON یا JSONC معتبر باشه. کامنت‌های `//` و `/* */` و BOM از نوع UTF-8 هم پذیرفته می‌شن.
- Outbound باید یه آدرس مقصد داشته باشه که BGScan بتونه با IP هر هدف جایگزینش کنه. این آدرس می‌تونه در `settings.address`، یکی از ورودی‌های `settings.vnext` یا `settings.servers` یا endpoint یک Peer در WireGuard باشه.
- پروتکل پروکسی باید معتبر باشه.
- Config باید از اعتبارسنجی هسته داخلی Xray عبور کنه. BGScan برای این کار از باینری خارجی `xray` استفاده نمی‌کنه.

می‌تونید از `$ADDRESS` به‌عنوان جای‌نگهدار آدرس استفاده کنید، اما اجباری نیست. هر آدرسی که در یکی از فیلدهای قابل‌جایگزینی قرار داشته باشه، موقع اسکن با IP هدف عوض می‌شه و پورت Outbound هم حفظ می‌شه. بنابراین، اگه یه Config آماده دارید، لازم نیست قبل از وارد کردن، آدرس سرور اصلی رو حذف کنید.

نمونه حداقلی از یک Outbound معتبر:

```json
{
  "protocol": "vless",
  "settings": {
    "vnext": [
      {
        "address": "$ADDRESS",
        "port": 443,
        "users": [
          { "id": "your-uuid-here", "encryption": "none" }
        ]
      }
    ]
  },
  "streamSettings": {
    "network": "ws",
    "security": "tls",
    "tlsSettings": { "serverName": "example.com" },
    "wsSettings": { "path": "/ws", "headers": { "Host": "example.com" } }
  }
}
```

بعد از انتخاب فایل، یه اسم برای Outbound وارد کنید. این اسم هم باید یکتا باشه.

```text
a → From File → Select file → Enter name → Done
```

{{< img "/bgscan-outbound-name.webp" "bgscan outbound name" >}}
{{< img "/bgscan-outbound-item.webp" "bgscan outbound item" >}}

---

## محل ذخیره Outboundها

Outboundها به‌صورت فایل `.json` در این پوشه ذخیره می‌شن:

```text
<bgscan-root>/
└── assets/
    └── xray/
        └── outbounds/
            ├── my-vless-ws.json
            ├── cloudflare-trojan.json
            └── ...
```

اسمی که در TUI می‌بینید، همون نام فایل بدون پسوند `.json` هست.

BGScan چند Outbound نمونه هم با پسوند `.json.example` در همین پوشه قرار می‌ده. این فایل‌ها به‌صورت خودکار بارگذاری نمی‌شن.

اگه خواستید از یکی از نمونه‌ها استفاده کنید، فایل رو کپی کنید، پسوند `.example` رو بردارید و مقادیرش رو با تنظیمات خودتون جایگزین کنید. بعد فایل رو در همین پوشه قرار بدید یا از طریق گزینه **From File** واردش کنید.

---

## جایگزینی آدرس و اعتبارسنجی

موقع اسکن Xray، BGScan برای هر IP یه Config در حافظه می‌سازه و آدرس مقصد Outbound رو با IP هدف جایگزین می‌کنه. پورت بدون تغییر باقی می‌مونه و فایل موقتی روی دیسک ساخته نمی‌شه.

هنگام افزودن Outbound، BGScan بررسی می‌کنه که آدرس قابل‌جایگزینی داشته باشه و Config ساخته‌شده از اعتبارسنجی هسته داخلی Xray عبور کنه.

Outboundهای ذخیره‌شده هنگام شروع برنامه هم دوباره بررسی می‌شن. اگه Outboundی معتبر نباشه، خطای مربوط به اون در بررسی‌های اولیه برنامه گزارش می‌شه.

---

## تغییر نام و حذف Outbound

**تغییر نام با کلید `r`:** Outbound موردنظرتون رو انتخاب کنید، `r` رو بزنید و اسم جدید رو وارد کنید. اسم جدید نباید تکراری باشه.

**حذف با کلید `x`:** Outbound رو انتخاب کنید و `x` رو بزنید. فایل مربوط به اون به‌طور دائمی از پوشه `assets/xray/outbounds/` حذف می‌شه و امکان برگردوندنش از داخل برنامه وجود نداره.

---

## موضوعات مرتبط

- [انواع اسکن](./scan-types.md)
- [ورودی اسکن](./scan-source.md)
- [فایل‌های نتیجه](./result-files.md)

