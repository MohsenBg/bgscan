---
title: "رابط کاربری"
weight: 4
---

# رابط کاربری

رابط ترمینالی bgscan با [BubbleTea](https://github.com/charmbracelet/bubbletea) v2 و [Lipgloss](https://github.com/charmbracelet/lipgloss) ساخته شده است. پایهٔ کار یک درخت کامپوننت است؛ صفحه‌ها در Stack قرار می‌گیرند و Dialogها روی صفحهٔ فعلی می‌آیند.

## بخش‌های اصلی

| پکیج | کار |
|---|---|
| `ui/main/app` | مدل اصلی، Header، Body، Footer و Dialogها |
| `ui/main/body` | نگه‌داری Stack صفحه‌ها |
| `ui/components/basic` | Widgetهایی مثل input، table، menu، dialog و progress |
| `ui/components/inspector` | فرم تنظیمات پروتکل‌ها |
| `ui/components/menus` | منوهای اصلی، اسکن، لاگ و تنظیمات |
| `ui/components/tables` | جدول IP list، Outbound، نتیجه و IP viewer |
| `ui/components/scanner` | صفحهٔ اسکن زنده |
| `ui/shared` | Layout، Dialog، کلیدها و قرارداد Component |
| `ui/theme` | تم روشن، تیره و خودکار |

## Componentها

هر صفحه و Widget رابط `Component` را پیاده می‌کند: `ID`، `Name`، `Init`، `Update`، `View`، `OnClose` و `Mode`.

مدل اصلی `AppState` را نگه می‌دارد؛ `AppState` شامل Layout، Config و Store است. کامپوننت‌ها تنظیمات را از Config می‌خوانند و با Store ذخیره می‌کنند. تنظیمات Global نداریم.

## حرکت بین صفحه‌ها

Body یک Stack است و فقط صفحهٔ بالایی پیام می‌گیرد. بازکردن صفحه، آن را Push می‌کند؛ Back یا `CloseComponentMsg` آن را Pop می‌کند. `ResetComponentStacksMsg` شما را به منوی اصلی برمی‌گرداند.

Dialogها هم Stack دارند. اگر Dialog باز باشد، فقط Dialog بالایی ورودی را می‌گیرد.

| حالت | Back | Quit |
|---|---|---|
| `NormalMode` | `b`، `Backspace`، `Esc` | `q`، `ctrl+c` |
| `InputMode` | `Esc` | `ctrl+c` |
| `ScanMode` | ندارد | `ctrl+c` |

## اسکن زنده و تنظیمات

برای هر Stage یک Tab ساخته می‌شود. هر Tab Progress bar و جدول نتیجهٔ زنده دارد. وضعیت Stageها از `Waiting` به `PreProcess`، `Scanning` و بعد `Ended` یا `Error` می‌رسد. در زمان اسکن Pause، Resume، Stop و جابه‌جایی Tab در دسترس است.

Inspector قبل از ذخیره، تغییر را روی یک کپی اعمال و Validate می‌کند. اگر مقدار غلط باشد، نه Config و نه فایل روی دیسک تغییر نمی‌کند. گزینه‌های پویا مثل TLS یا DNSTT فقط وقتی لازم باشند نشان داده می‌شوند.
