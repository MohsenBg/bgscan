<div dir="rtl">

<div align="center">

<br/>

```
██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██╔════╝ ██╔════╝██╔════╝██╔══██╗████╗  ██║
██████╔╝██║  ███╗███████╗██║     ███████║██╔██╗ ██║
██╔══██╗██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
```

**اسکنر چندپروتکلی فوق‌سریع با معماری زنجیره‌ای ماژولار**
<div align="center">

[English](README.md) |
[فارسی](README.fa.md)

</div>
<br/>

[![Go Version](https://img.shields.io/badge/Go-1.26.3+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-6366f1?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20|%20Windows%20|%20macOS%20|%20Termux-64748b?style=flat-square)](https://github.com/MohsenBg/bgscan/releases)
[![UI](https://img.shields.io/badge/UI-BubbleTea%20TUI-ec4899?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-22c55e?style=flat-square)](https://github.com/MohsenBg/bgscan/releases/latest)

<br/>

[📦 نصب](#-نصب) · [📚 مستندات](#-مستندات) · [🔹 پروتکل‌ها](#-پروتکل‌های-پشتیبانی‌شده) · [💖 حمایت](#-حمایت--دونیت)

<br/>

---

</div>

## درباره bgscan

ابزار **bgscan** یک موتور اسکن با کارایی بالا است که با زبان **Go** توسعه یافته.
این ابزار چندین پروتکل را به‌صورت همزمان اجرا می‌کند و مراحل مختلف را به یکدیگر متصل می‌کند تا جریان‌های تشخیص و شناسایی پیشرفته بسازد.

ساخته شده برای توسعه‌دهندگان و پژوهشگرانی که به **سرعت، انعطاف‌پذیری و تجربه‌ای مدرن** در اسکن نیاز دارند.

---
<img width="1552" height="841" alt="bgscan-ui" src="https://github.com/user-attachments/assets/08a50bc0-186d-45a4-8c27-4bb52a2041ee" />

## ✨ ویژگی‌های کلیدی

<table>
<tr>
<td width="50%">

### 🔗 زنجیره‌سازی هوشمند
امکان chain کردن چند مرحله اسکن با حالت‌های اجرای `Stream`، `Sequential` و `Batch Size` برای کنترل کامل جریان داده.

</td>
<td width="50%">

### 🖥️ رابط کاربری TUI
رابط کاربری تعاملی مبتنی بر **Bubble Tea** — اسکن، مانیتورینگ و مدیریت نتایج بدون نیاز به مرورگر.

</td>
</tr>
<tr>
<td width="50%">

### 📡 پشتیبانی گسترده از پروتکل‌ها
از ICMP و TCP تا DNS Tunnel، Slipstream و Xray — همه در یک ابزار.

</td>
<td width="50%">

### 💾 ذخیره و بازاجرا
نتایج اسکن را ذخیره کنید و اسکن‌های جدید را مستقیماً روی داده‌های قبلی اجرا کنید.

</td>
</tr>
<tr>
<td width="50%">

### 🛰️ یکپارچه‌سازی با Xray
ذخیره، مدیریت و اعتبارسنجی Outboundهای Xray به‌صورت کامل.

</td>
<td width="50%">

### 🌐 پروتکل DNS پیشرفته
پشتیبانی از DNS Tunnel و مکانیزم Fallback برای سناریوهای پیچیده.

</td>
</tr>
</table>

---

## 🔹 پروتکل‌های پشتیبانی‌شده

| پروتکل | لایه | توضیحات |
|:------:|:----:|---------|
| **ICMP** | ۳ | شناسایی و بررسی دسترس‌پذیری میزبان از طریق Ping |
| **TCP** | ۴ | اسکن و اعتبارسنجی اتصال TCP Handshake |
| **HTTP** | ۷ | پشتیبانی از HTTP/1.1 و HTTP/2 *(بدون QUIC/HTTP3)* |
| **TLS** | ۷ | پشتیبانی از TLS 1.0 تا TLS 1.3 |
| **DNS** | ۷ | پرس‌وجوی پیشرفته DNS همراه با مکانیزم Fallback |
| **DNSTT** | ۷ | اعتبارسنجی DNS Tunnel *(فقط SOCKS، بدون احراز هویت)* |
| **Slipstream** | ۷ | اعتبارسنجی Slipstream *(فقط SOCKS، بدون احراز هویت)* |
| **Xray** | ۷ | اعتبارسنجی و تست Outboundهای Xray |

---

## 📦 نصب

### نصب خودکار (توصیه‌شده)

**🪟 Windows**
```powershell
irm https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.ps1 | iex
```

**🐧 Linux / 🍎 macOS**
```bash
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | sh
```

**🤖 Android (Termux)**
```bash
pkg update -y && pkg install bash curl unzip -y
curl -fsSL https://raw.githubusercontent.com/MohsenBg/bgscan/refs/heads/main/scripts/install.sh | bash
```

---

### نصب دستی

فایل متناسب با سیستم‌عامل و معماری پردازنده خود را از صفحه Releases دانلود کنید:

👉 **[آخرین نسخه را دانلود کنید](https://github.com/MohsenBg/bgscan/releases/latest)**

**Linux / macOS / Termux**
```bash
unzip bgscan-*.zip
chmod +x bgscan
./bgscan
```

**Windows**

فایل ZIP را Extract کرده و `bgscan.exe` را اجرا کنید.

---

## 📚 مستندات

### 👤 مستندات کاربر
آشنایی با نحوه استفاده، تنظیمات، پروتکل‌های پشتیبانی‌شده، حالت‌های اجرا و مثال‌های کاربردی:

📖 **[→ مستندات کاربر](docs/user/README.fa.md)**

---

### 🛠️ مستندات توسعه‌دهندگان
معماری پروژه، ساختار ماژول‌ها، موتور Pipeline، سیستم Assetها، استانداردهای کدنویسی و راهنمای مشارکت:

📖 **[→ مستندات توسعه‌دهندگان](docs/developer/README.fa.md)**

---

## 💖 حمایت / دونیت

اگر این پروژه برایت مفید بوده، می‌توانی از آدرس‌های زیر کمک مالی ارسال کنی:

| شبکه | ارز | آدرس |
|:----:|:---:|------|
| ₿ بیت‌کوین | `BTC` | `bc1qdwh57dm97nmx5jzdr7lrc9cxe5xh3zc59er7z9` |
| 🟣 اتریوم | `ETH` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| 🟡 بایننس اسمارت چین | `BNB / BEP20` | `0x40Fd22Fff4E059e906A10747Fd0a45A1DB39c985` |
| 🔷 ترون | `TRX / TRC20` | `TNW6pbfY8zZVZezZWyYXo7h12MycRsVJK7` |

---

<div align="center">

ساخته شده با ❤️ و Go &nbsp;|&nbsp; مجوز MIT

</div>

</div>
