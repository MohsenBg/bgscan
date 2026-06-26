<div align="left">

[**English**](./xray_outbounds.md)  |  [**فارسی**](./xray_outbounds.fa.md)

</div>

<div dir="rtl">

# اضافه کردن یک Outbound سفارشی برای Xray

## فهرست مطالب

- [روش ۱: ویزارد تعاملی Outbound](#روش-۱-ویزارد-تعاملی-outbound)
  - [گزینه الف: اضافه کردن از طریق لینک](#گزینه-الف-اضافه-کردن-از-طریق-لینک)
  - [گزینه ب: اضافه کردن از طریق فایل JSON](#گزینه-ب-اضافه-کردن-از-طریق-فایل-json)
  - [نتیجه](#نتیجه)
- [روش ۲: ویرایش دستی قالب (پیشرفته)](#روش-۲-ویرایش-دستی-قالب-پیشرفته)
  - [۱. رفتن به پوشه Outbounds](#۱-رفتن-به-پوشه-outbounds)
  - [۲. کپی کردن و تغییر نام قالب](#۲-کپی-کردن-و-تغییر-نام-قالب)
  - [۳. ویرایش فایل پیکربندی](#۳-ویرایش-فایل-پیکربندی)
  - [۴. استفاده از Outbound در BGScan](#۴-استفاده-از-outbound-در-bgscan)

---

اسکنر BGScan دو روش برای اضافه کردن یک Outbound سفارشی Xray دارد:

1. **ویزارد تعاملی Outbound** (توصیه‌شده) — اضافه کردن مستقیم از اپلیکیشن با استفاده از یک لینک اشتراک‌گذاری (VLESS / VMess / Trojan / Shadowsocks) یا یک فایل JSON.
2. **ویرایش دستی قالب** — کپی کردن یک فایل قالب و ویرایش دستی آن.

---

## روش ۱: ویزارد تعاملی Outbound

این سریع‌ترین روش برای اضافه کردن یک Outbound است — بدون نیاز به ویرایش دستی هیچ فایلی.

### مراحل

1. اسکنر **BGScan** را باز کنید.
2. به مسیر زیر بروید:

</div>

```
Xray → Outbounds
```

<div dir="rtl">

3. دکمه/کلید **a** را بزنید تا پنجره **Add Outbound** باز شود.

![menu](https://github.com/user-attachments/assets/d2a38b21-bf09-44b7-bca4-4267a284740b)


4. سپس در پنجره باز شده، روش اضافه کردن Outbound را انتخاب کنید:
   - **From Link** — لینک اشتراک‌گذاری VLESS، VMess، Trojan یا 
   Shadowsocks
   - **From JSON File** — یک فایل JSON سازگار با Xray

---

### گزینه الف: اضافه کردن از طریق لینک

1. گزینه **From Link** را انتخاب کنید.

![add_via_link](https://github.com/user-attachments/assets/eb23f4de-945d-4739-a504-176d8b73edf7)

2. لینک اشتراک‌گذاری خود را در جای مناسب وارد کنید (مثلاً `vless://...`، `vmess://...`، `trojan://...`، `ss://...`).

![paste link](https://github.com/user-attachments/assets/c5683fd3-b777-46b4-ba24-3f9938db5b78)

3. اسکنر BGScan به‌طور خودکار لینک را تجزیه کرده و پروتکل، سرور، پورت، اعتبارنامه‌ها و تنظیمات انتقال را استخراج می‌کند.
4. یک **نام** برای Outbound وارد کنید.

![enter name](https://github.com/user-attachments/assets/5ecea6ee-6602-4671-8a4d-6b5cd2510029)

5. درنهایت Outbound ذخیره شده و بلافاصله در لیست **Outbounds** قابل مشاهده است.

</div>

```
Add Outbound → From Link → وارد کردن لینک → وارد کردن نام → تمام
```

<div dir="rtl">

> فرمت‌های لینک پشتیبانی‌شده: `vless://`، `vmess://`، `trojan://`، `ss://`

---

### گزینه ب: اضافه کردن از طریق فایل JSON

1. گزینه **From JSON File** را انتخاب کنید.

![add via json](https://github.com/user-attachments/assets/2ef7ad7c-52b9-4603-b757-cd2bd3e826f3)

2. فایل `.json` مربوط به Outbound را پیدا کرده و انتخاب کنید.

![select json file](https://github.com/user-attachments/assets/a5cd36d3-bdfa-4ed0-a348-0a7962336e56)

3. یک **نام** برای Outbound وارد کنید.
4. Outbound ذخیره شده و بلافاصله در لیست **Outbounds** قابل مشاهده است.

</div>

```
Add Outbound → From JSON File → انتخاب فایل → وارد کردن نام → تمام
```

<div dir="rtl">

> ⚠️ **الزامات فرمت فایل JSON:**
> فایل JSON انتخاب‌شده باید حاوی **یک شیء Outbound منفرد** باشد و دقیقاً همان فرمت قالب دستی (مراجعه کنید به [روش ۲](#روش-۲-ویرایش-دستی-قالب-پیشرفته)) را داشته باشد — از جمله placeholder مقدار `"address": "$ADDRESS"` که BGScan در طول آزمایش آن را به‌طور خودکار جایگزین می‌کند. یک پیکربندی کامل Xray (با `outbounds: [...]`، `inbounds`، `routing` و غیره) ارائه ندهید — فقط بلوک Outbound منفرد را وارد کنید.
>
> مثال معتبر حداقلی:

</div>

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

<div dir="rtl">

> اگر فایل با این فرمت مطابقت نداشته باشد (مثلاً `$ADDRESS` وجود نداشته باشد یا در یک پیکربندی کامل قرار گرفته باشد)، import ناموفق خواهد بود یا Outbound قابل آزمایش نخواهد بود.

---

### نتیجه

پس از اضافه شدن (از هر روشی)، Outbound جدید شما در لیست نمایش داده می‌شود:

![outbounds](./images/outbounds.png)

می‌توانید در هر زمان ویزارد را دوباره اجرا کرده و Outbound‌های بیشتری اضافه کنید — بدون نیاز به دسترسی به فایل‌سیستم.

---

## روش ۲: ویرایش دستی قالب (پیشرفته)

اگر کنترل کامل دستی روی پیکربندی ترجیح می‌دهید، می‌توانید یک فایل قالب را مستقیماً ویرایش کنید.

### ۱. رفتن به پوشه Outbounds

تمام قالب‌های Outbound در مسیر زیر ذخیره هستند:

</div>

```
assets/xray/outbounds/
```

<div dir="rtl">

نمونه‌ای از ساختار پوشه:

</div>

```
assets/xray/outbounds/
├── vless_grpc.json.example
├── vless_ws.json.example
├── vless_ws_no_tls.json.example
├── vless_xhttp.json.example
└── vless_xhttp_no_tls.json.example
```

<div dir="rtl">

فایل‌هایی با پسوند `.example` قالب هستند.

---

### ۲. کپی کردن و تغییر نام قالب

یک قالب انتخاب کنید (مثلاً: `vless_ws.json.example`) و آن را کپی کنید:

</div>

```bash
cp vless_ws.json.example config.json
```

<div dir="rtl">

ساختار پوشه پس از این مرحله:

</div>

```
assets/xray/outbounds/
├── vless_grpc.json.example
├── vless_ws.json.example
├── vless_ws_no_tls.json.example
├── vless_xhttp.json.example
├── vless_xhttp_no_tls.json.example
└── config.json
```

<div dir="rtl">

می‌توانید هر نامی برای فایل انتخاب کنید (`my_ws.json`، `tls_ws.json` و غیره).

---

### ۳. ویرایش فایل پیکربندی

فایل خود را باز کنید:

</div>

```
assets/xray/outbounds/config.json
```

<div dir="rtl">

تمام فیلدهایی که با `?` علامت‌گذاری شده‌اند را با مقادیر واقعی Outbound خود جایگزین کنید.

> ⚠️ **مهم:** فیلد `address` را **ویرایش نکنید**:

</div>

```json
"address": "$ADDRESS"
```

<div dir="rtl">

> اسکنر مقدار `$ADDRESS` را در طول آزمایش به‌طور خودکار جایگزین می‌کند.

---

#### نمونه قالب (قبل از ویرایش)

</div>

```json
{
  "tag": "proxy",
  "protocol": "vless",
  "settings": {
    "vnext": [
      {
        "address": "$ADDRESS",
        "port": 443,
        "users": [
          {
            "id": "?",
            "encryption": "none"
          }
        ]
      }
    ]
  },
  "streamSettings": {
    "network": "ws",
    "security": "tls",
    "tlsSettings": {
      "allowInsecure": false,
      "serverName": "?",
      "alpn": ["h2", "http/1.1"],
      "fingerprint": "firefox"
    },
    "wsSettings": {
      "path": "?",
      "headers": {
        "Host": "?"
      }
    }
  }
}
```

<div dir="rtl">

#### نمونه (پس از پر کردن مقادیر)

</div>

```json
{
  "tag": "proxy",
  "protocol": "vless",
  "settings": {
    "vnext": [
      {
        "address": "$ADDRESS",
        "port": 443,
        "users": [
          {
            "id": "3f1e6f4c-9f1c-4a3a-bf10-9e2c8a123456",
            "encryption": "none"
          }
        ]
      }
    ]
  },
  "streamSettings": {
    "network": "ws",
    "security": "tls",
    "tlsSettings": {
      "allowInsecure": false,
      "serverName": "example.com",
      "alpn": ["h2", "http/1.1"],
      "fingerprint": "firefox"
    },
    "wsSettings": {
      "path": "/ws",
      "headers": {
        "Host": "example.com"
      }
    }
  }
}
```

<div dir="rtl">

فیلدهایی که باید پر شوند:

| فیلد | توضیح |
|---|---|
| `id` | UUID مربوط به VLESS شما |
| `serverName` | نام سرور TLS (SNI) |
| `path` | مسیر WebSocket |
| `Host` | هدر Host |

---

### ۴. استفاده از Outbound در BGScan

پس از ذخیره فایل JSON:

1. اسکنر **BGScan** را باز کنید.
2. به مسیر زیر بروید:

</div>

```
Xray → Outbounds
```

<div dir="rtl">

3. در نهایت Outbound جدید شما به‌طور خودکار در لیست نمایش داده می‌شود.

![menu](https://github.com/user-attachments/assets/583d4d51-a718-40c0-9272-292c4d0c9c1b)
![outbounds](https://github.com/user-attachments/assets/99c9e713-2811-4787-8247-4056dfeefb81)

</div>
