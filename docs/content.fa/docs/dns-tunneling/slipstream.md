---
title: "Slipstream"
weight: 3
---

# تنظیمات Slipstream

Slipstream داخل خود BGScan و از طریق `libslipstream-client` اجرا می‌شه. این کتابخانه کلاینت با `purego` بارگذاری می‌شه. بنابراین، نیازی به باینری جداگانه نیست و BGScan هم چیزی رو در `PATH` جست‌وجو نمی‌کنه.

هر پروب یک تونل راه‌اندازی می‌کنه و عبور ترافیک از اون رو از طریق یک شنونده محلی SOCKS5 روی پورتی که به پروب اختصاص داده شده، بررسی می‌کنه.

## تنظیمات اتصال

| فیلد | پیش‌فرض | محدوده | توضیح |
| --- | --- | --- | --- |
| `domain` | الزامی | دامنه معتبر | دامنه‌ای که سرور Slipstream شما سرویس می‌ده |
| `resolver_port` | `53` | بزرگ‌تر از `0` | پورت ریزالور |
| `cert_path` | `""` | مسیر فایل | گواهی CA برای اعتبارسنجی سرور، مقدار خالی یعنی استفاده از مخزن گواهی‌های سیستم |
| `dns_resolution` | `"recursive"` | `recursive`، `authoritative` | روش Resolve کردن نام‌ها توسط کلاینت |
| `gso` | `false` | `true`، `false` | فعال کردن Generic Segmentation Offload برای UDP |
| `congestion_control` | `"cubic"` | `bbr`، `cubic` | الگوریتم کنترل ازدحام تونل، مقدار خالی یعنی استفاده از پیش‌فرض کتابخانه |

## پروکسی و احراز هویت

Slipstream از همون فیلدهای پروکسی DNSTT استفاده می‌کنه. برای جزئیات، بخش [پروکسی و احراز هویت](./dnstt/#پروکسی-و-احراز-هویت) رو ببینید.

برای استفاده از پروکسی SSH، فیلد `known_hosts_file` باید به یک فایل معتبر `known_hosts` اشاره کنه. این فیلد در فرم TUI نمایش داده نمی‌شه. بنابراین، باید مسیرش رو مستقیماً در فایل TOML وارد کنید.

## محل کتابخانه

BGScan پوشه‌های زیر رو به‌ترتیب برای پیدا کردن کتابخانه کلاینت متناسب با پلتفرم بررسی می‌کنه:

1. `<bgscan-root>/assets/slipstream-client/`
2. `<bgscan-root>/assets/slipstream/`
3. `<bgscan-root>/assets/`
4. `<bgscan-root>/`

نام فایل‌های مورد جست‌وجو بر اساس پلتفرم:

| پلتفرم | نام فایل‌ها |
| --- | --- |
| Linux / Android | `libslipstream-client.so`، `libslipstream-client-*.so` |
| macOS | `libslipstream-client.dylib`، `libslipstream-client-*.dylib` |
| Windows | `slipstream-client.dll`، `libslipstream-client.dll`، `libslipstream-client-*.dll` |

اسکریپت‌های Build کتابخانه رو دانلود می‌کنن و فایل اون داخل مخزن نگه‌داری نمی‌شه. هنگام شروع برنامه، BGScan بارگذاری کتابخانه رو بررسی می‌کنه و نسخه‌اش رو گزارش می‌ده. اگه کتابخانه پیدا نشه، اسکن Slipstream اجرا نمی‌شه و خطا گزارش می‌شه.

## نمونه Config

```toml id="k4v9sz"
domain = "example.com"
resolver_port = 53
cert_path = ""
dns_resolution = "recursive"
gso = false
congestion_control = "cubic"
proxy_type = "socks"
proxy_port = 1080
auth_method = "none"
username = ""
password = ""
private_key = ""
known_hosts_file = ""
```

## موضوعات مرتبط

- [تونل DNS](../): آشنایی با پروتکل‌ها و روند اسکن
- [DNSTT](./dnstt/): توضیحات فیلدهای پروکسی و احراز هویت

