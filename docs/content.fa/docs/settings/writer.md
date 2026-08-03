---
title: "تنظیمات ذخیرهٔ نتیجه"
weight: 4
---

# تنظیمات ذخیرهٔ نتیجه

> 💡 **نکته:** این گزینه‌ها را از **Settings → General Settings** و تب دوم هم می‌توانید تغییر دهید.

فایل تنظیمات: `settings/writer_settings.toml`

این فایل مشخص می‌کند نتیجه‌ها چطور بافر و روی دیسک نوشته شوند و کجا ذخیره شوند.

هر مرحلهٔ اسکن Writer خودش را دارد. Workerها نتیجه را داخل یک Channel می‌فرستند، Writer آن‌ها را در Batch جمع می‌کند و هنگام Flush آن Batch را با فایل CSV مرحله ادغام و مرتب می‌کند.

## خلاصهٔ گزینه‌ها

| گزینه | پیش‌فرض | توضیح |
|---|---:|---|
| `merge_flush_interval` | `2000` | فاصلهٔ Flush، میلی‌ثانیه |
| `chan_size` | `2048` | ظرفیت Channel بین Worker و Writer |
| `batch_size` | `4096` | تعداد نتیجهٔ نگه‌داری‌شده در حافظه تا Flush |
| `result_directory` | `"result"` | پوشهٔ پایهٔ نتیجه‌ها |

## Merge Flush Interval

```toml
merge_flush_interval = 2000
```

فاصلهٔ Flushهای دوره‌ای بر حسب میلی‌ثانیه است. محدودهٔ مجاز ۱۰۰ میلی‌ثانیه تا ۵ دقیقه است. وقتی تعداد نتیجه‌ها به `batch_size` برسد هم Flush انجام می‌شود و هنگام توقف Writer یک Flush نهایی انجام می‌شود؛ بنابراین نتیجه‌ای که قبل از توقف پذیرفته شده گم نمی‌شود.

فاصلهٔ کوتاه‌تر، نتیجه‌ها را زودتر روی دیسک می‌نویسد اما فایل بیشتر بازنویسی می‌شود. فاصلهٔ بلندتر، I/O دیسک را کم می‌کند اما نتیجه‌های بیشتری در حافظه می‌مانند.

## Channel Size

```toml
chan_size = 1024
```

ظرفیت Channelی است که Workerها نتیجه را از طریق آن به Writer می‌دهند. وقتی Channel پر شود، Worker تا رسیدن Writer صبر می‌کند. اگر یک مرحله نتیجه‌ها را Burstی تولید می‌کند و ادغام به آن سرعت نیست، مقدار را بیشتر کنید. محدودهٔ مجاز ۱ تا ۱٬۰۰۰٬۰۰۰ است.

## Batch Size

```toml
batch_size = 4096
```

تعداد نتیجه‌هایی که قبل از Flush اجباری در حافظه می‌مانند. همچنین ظرفیت اولیهٔ Slice مربوط به Batch و اندازهٔ Buffered Writer زمان ادغام را تعیین می‌کند. محدودهٔ مجاز ۱ تا ۱٬۰۰۰٬۰۰۰ است.

## Result Directory

```toml
result_directory = "result"
```

پوشهٔ پایهٔ نتیجه‌ها، نسبت به باینری bgscan است. برای هر Schema یک زیرپوشه ساخته می‌شود: `result/icmp/`، `result/tcp/`، `result/http/`، `result/xray/`، `result/dns_resolver/`، `result/dnstt/` و `result/slipstream/`.

این مقدار باید فقط نام پوشه باشد، نه یک مسیر.

## فایل‌های مرتبط

- [`general_settings.toml`](./general.md)
- [`icmp_settings.toml`](./icmp.md)
- [`tcp_settings.toml`](./tcp.md)
- [`http_settings.toml`](./http.md)
- [`dns_settings.toml`](./dns.md)
- [`xray_settings.toml`](./xray.md)
