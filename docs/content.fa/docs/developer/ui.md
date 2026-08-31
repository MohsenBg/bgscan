---
title: "رابط کاربری"
weight: 4
---

# رابط کاربری

رابط ترمینالی bgscan با [BubbleTea](https://github.com/charmbracelet/bubbletea) v2 و [Lipgloss](https://github.com/charmbracelet/lipgloss) ساخته شده است. پایهٔ کار یک درخت کامپوننت است؛ صفحه‌ها در Stack قرار می‌گیرند و Dialogها روی صفحهٔ فعلی می‌آیند.

## بخش‌های اصلی

| پکیج | کار |
|---|---|
| `ui/main/app` | مدل اصلی و ماشین مرحله‌ها: Splash → Startup → Workspace |
| `ui/main/splash` | صفحهٔ Splash با انیمیشن لوگو و نمایش نسخه |
| `ui/main/startup` | بررسی‌های اولیه: لاگر، Config، Xray، DNSTT، Slipstream، VayDNS و وضعیت برنامه |
| `ui/main/workspace` | Workspace اصلی؛ Header، Body و Footer را نگه می‌دارد |
| `ui/main/body` | نگه‌داری Stack صفحه‌ها داخل Workspace |
| `ui/components/basic` | Widgetهایی مثل input، table، menu، dialog، crud و progress |
| `ui/components/form` | فرم‌های Config تونل DNS: dnstt، slipstream و vaydns |
| `ui/components/inspector` | فرم تنظیمات پروتکل‌ها |
| `ui/components/menus` | منوهای اصلی، اسکن، لاگ، تونل DNS و تنظیمات |
| `ui/components/tables` | جدول IP list، Outbound، تونل DNS، نتیجه و IP viewer |
| `ui/components/scanner` | صفحهٔ اسکن زنده |
| `ui/shared` | Layout، Dialog، کلیدها و قرارداد Component |
| `ui/theme` | تم روشن، تیره و خودکار + آداپتر فرم huh |

## Componentها

هر صفحه و Widget رابط `Component` را پیاده می‌کند: `ID`، `Name`، `Init`، `Update`، `View`، `OnClose` و `Mode`.

مدل اصلی `AppState` را نگه می‌دارد؛ `AppState` شامل Layout، Config و Store است. کامپوننت‌ها تنظیمات را از Config می‌خوانند و با Store ذخیره می‌کنند. تنظیمات Global نداریم.

## سه مرحلهٔ برنامه

برنامه از سه مرحله می‌گذرد:

| مرحله | کامپوننت | کار |
|---|---|---|
| Splash | `splash` | انیمیشن لوگوی ASCII و نمایش نسخه |
| Startup | `startup` | اجرای بررسی‌های اولیه به‌ترتیب (لاگر → Config → Xray → DNSTT → Slipstream → VayDNS → App)؛ هر قدم در نوار کناری نشان داده می‌شود و بعد از پاس‌شدن همه، با `Enter` مرحلهٔ بعد شروع می‌شود |
| Workspace | `workspace` | صفحهٔ اصلی برنامه با Header، Body (Stack صفحه‌ها)، Footer و Dialogها |

Config در همین مرحلهٔ Startup بارگذاری و Validate می‌شود و بعد از آن برای همهٔ کامپوننت‌ها در `AppState` در دسترس است.

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

Inspector قبل از ذخیره، تغییر را روی یک کپی اعمال و Validate می‌کند. اگر مقدار غلط باشد، نه Config و نه فایل روی دیسک تغییر نمی‌کند. گزینه‌های پویا مثل TLS فقط با انتخاب HTTPS نشان داده می‌شوند و فیلدهای مخصوص هر پروتکل تونل فقط در فرم همان پروتکل ظاهر می‌شوند.

مدیریت Configهای تونل DNS هم در همین رابط انجام می‌شود: جدول **Main Menu → DNS Tunneling** همهٔ تونل‌های ذخیره‌شده را نشان می‌دهد و با کلیدهای `a`، `r` و `x` می‌توانید Config اضافه، تغییر نام یا حذف کنید. با `Enter` هم Config انتخابی ویرایش می‌شود یا اسکن تونل با آن شروع می‌گردد.
