---
title: "DNSTT"
weight: 1
---

# تنظیمات DNSTT

DNSTT از کتابخانه VayDNS با فریمینگ سازگار با DNSTT استفاده می‌کنه، یعنی DNS با TXT و فریمینگ قدیمی. ساختار تونل به این صورته: کانال بسته‌های DNS ← KCP ← رمزنگاری Noise ← مالتی‌پلکس smux.

## تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیح |
| --- | --- | --- | --- |
| `domain` | الزامی | دامنه معتبر | Zoneای که به سرور DNSTT شما واگذار شده |
| `pub_key` | الزامی | ۶۴ کاراکتر Hex | کلید عمومی سرور |
| `resolver_type` | `"udp"` | `udp`، `tcp`، `dot` | ترنسپورت ارتباط با ریزالور |
| `resolver_port` | `53` | بزرگ‌تر از `0` | پورت ریزالور |
| `fingerprint` | `"Chrome"` | طبق جدول پایین | Fingerprint مربوط به ClientHello در uTLS برای DNS با TLS |
| `rps` | `0` | ۰ تا ۵۰۰ | محدودیت نرخ کوئری در ثانیه، مقدار `0` یعنی بدون محدودیت |

## پروکسی و احراز هویت

| فیلد | پیش‌فرض | توضیح |
| --- | --- | --- |
| `proxy_type` | `"socks"` | نوع پروکسی: `socks` یا `ssh` |
| `proxy_port` | `1080` | پورت پروکسی |
| `auth_method` | `"none"` | روش احراز هویت: `none`، `password` یا `key` |
| `username` | `""` | نام کاربری. برای احراز هویت با Password یا Key لازمه |
| `password` | `""` | رمز عبور. برای احراز هویت با Password لازمه |
| `private_key` | `""` | کلید خصوصی SSH با کدگذاری PEM. برای احراز هویت با Key لازمه |
| `known_hosts_file` | `""` | مسیر فایل `known_hosts` مربوط به SSH. برای `proxy_type = "ssh"` لازمه و هنگام اجرا اعتبارسنجی می‌شه |

**نکات مربوط به پروکسی:**

- پروکسی SSH به احراز هویت با Password یا Key و یک فایل `known_hosts_file` نیاز داره. BGScan به‌صورت خودکار سراغ `~/.ssh/known_hosts` نمی‌ره و این فیلد هم در فرم TUI نمایش داده نمی‌شه. برای استفاده از SSH باید مسیر فایل رو مستقیماً در فایل TOML وارد کنید.
- پروکسی SOCKS از احراز هویت با Key پشتیبانی نمی‌کنه.

## Fingerprintهای TLS

فیلد `fingerprint` مشخص می‌کنه هنگام TLS Handshake از کدوم پروفایل ClientHello در uTLS استفاده بشه. نام‌ها بدون حساسیت به حروف بزرگ و کوچک مقایسه می‌شن.

| دسته | مقادیر قابل استفاده |
| --- | --- |
| Chrome | `Chrome`، `Chrome_58`، `Chrome_62`، `Chrome_70`، `Chrome_72`، `Chrome_83`، `Chrome_87`، `Chrome_96`، `Chrome_100`، `Chrome_102`، `Chrome_120` |
| Firefox | `Firefox`، `Firefox_55`، `Firefox_56`، `Firefox_63`، `Firefox_65`، `Firefox_99`، `Firefox_102`، `Firefox_105`، `Firefox_120` |
| iOS | `iOS`، `iOS_11_1`، `iOS_12_1`، `iOS_13`، `iOS_14` |
| سایر | `random` |

## نمونه Config

```toml
domain = "example.com"
pub_key = "0000000000000000000000000000000000000000000000000000000000000000"
resolver_type = "udp"
resolver_port = 53
fingerprint = "Chrome"
rps = 0
proxy_type = "socks"
proxy_port = 1080
auth_method = "none"
```

## موضوعات مرتبط

- [تونل DNS](../): آشنایی با پروتکل‌ها و روند اسکن
- [VayDNS](./vaydns/): پروتکل بومی VayDNS با گزینه‌های مشابه برای پروکسی

