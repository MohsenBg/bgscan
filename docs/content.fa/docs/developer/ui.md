---
title: "رابط کاربری"
weight: 4
---

# رابط کاربری

رابط ترمینالی bgscan با [Bubble Tea](https://github.com/charmbracelet/bubbletea) نسخهٔ ۲ (`charm.land/bubbletea/v2`) ساخته شده و برای استایل از [Lipgloss](https://github.com/charmbracelet/lipgloss) استفاده می‌کنه. ساختار رابط بر پایهٔ درختی از کامپوننت‌هاست؛ Dialogها به‌صورت Overlay روی صفحه قرار می‌گیرن و یک مدیر مرکزی هم چیدمان رو کنترل می‌کنه.

---

## نقشهٔ پکیج‌ها

| پکیج | مسئولیت |
|---|---|
| `ui/main/app` | مدل اصلی Bubble Tea که سه مرحلهٔ اپ رو مدیریت می‌کنه: Splash ← Startup ← Workspace |
| `ui/main/splash` | صفحهٔ شروع با انیمیشن لوگوی ASCII و نمایش نسخه |
| `ui/main/startup` | بررسی‌های اولیه؛ لاگرها رو بررسی می‌کنه، Schemaهای Probe رو ثبت می‌کنه و Config، Xray، DNSTT، Slipstream، VayDNS، MasterDNS، StormDNS، TheFeed و وضعیت اپ رو اعتبارسنجی می‌کنه |
| `ui/main/workspace` | محیط اصلی برنامه شامل Body، Footer و Stack دیالوگ‌های Overlay |
| `ui/main/body` | بخش اصلی محتوا که Stack کامپوننت‌ها رو نگه می‌داره؛ از منوی اصلی تا صفحه‌های مختلف |
| `ui/main/footer` | نوار وضعیت شامل نسخهٔ برنامه، نام صفحهٔ فعلی، تعداد Goroutineها و مصرف حافظه |
| `ui/components/basic` | Widgetهای قابل استفادهٔ مجدد مثل confirm، crud، form، input، inspector، logview، menu، notice، picker، progress، table و tabs |
| `ui/components/form` | فرم‌های تنظیمات تونل DNS برای DNSTT، Slipstream، VayDNS، MasterDNS، StormDNS و TheFeed که بر پایهٔ `formkit` مشترک ساخته شدن |
| `ui/components/inspector` | فرم‌های تنظیمات پروتکل‌ها و بخش‌های مختلف برنامه: DNS، General/Writer، HTTP، ICMP، TCP، Xray و UI |
| `ui/components/menus` | منوهای اصلی و فرعی مثل entry، logs، outbound، scantype، settings، targetsource و dnstunmenu |
| `ui/components/tables` | جدول‌های داده شامل iplist، ipviewer، outbounds، resultlist و dnstun برای مدیریت تونل‌های DNS |
| `ui/components/scanner` | نمای زندهٔ اسکن با نوار پیشرفت مرحله‌بندی‌شده و جدول نتیجه برای هر Stage |
| `ui/shared` | زیرساخت مشترک رابط: Layout، Dialogها، Session/Env، کلیدها، اینترفیس کامپوننت، انیمیشن، Meter و اعتبارسنجی |
| `ui/theme` | ۲۱ تم تیره و روشن، تشخیص خودکار تم از پس‌زمینهٔ ترمینال و آداپتر تم فرم‌های huh |

---

## اینترفیس کامپوننت

هر صفحه، Widget و Dialog اینترفیس `ui.shared.ui.Component` رو پیاده‌سازی می‌کنه:

```go
type Component interface {
    ID() ComponentID
    Name() string
    Init() tea.Cmd
    Update(tea.Msg) (Component, tea.Cmd)
    View() string
    OnClose() tea.Cmd
    Mode() KeyMode
}
```

- **`ID`**: یک UUID که با `NewComponentID()` ساخته می‌شه و برای تشخیص کامپوننت در پیام‌های باز و بسته‌شدن استفاده می‌شه.
- **`Name`**: نام قابل‌خواندن کامپوننت که در نوار وضعیت Footer نمایش داده می‌شه.
- **`Init`**: هنگام اضافه‌شدن کامپوننت به رابط، یک بار اجرا می‌شه. برای کارهای غیرهم‌زمان، یک `tea.Cmd` برمی‌گردونه.
- **`Update`**: پیام‌های دریافتی رو پردازش می‌کنه و کامپوننت به‌روزشده به‌همراه Commandهای موردنیاز رو برمی‌گردونه.
- **`View`**: خروجی متنی کامپوننت رو برای نمایش در ترمینال می‌سازه.
- **`OnClose`**: هنگام برداشته‌شدن کامپوننت از Stack، کارهای پاک‌سازی رو انجام می‌ده.
- **`Mode`**: مشخص می‌کنه کدوم کلیدها برای Back و Quit استفاده بشن. جزئیات بیشتر در بخش [حالت‌های ورودی](#حالتهای-ورودی) اومده.

---

## مدل اصلی اپ (`app`)

پکیج `ui/main/app` مدل اصلی Bubble Tea رو پیاده‌سازی می‌کنه. این مدل مراحل مختلف برنامه و وضعیت مشترک رابط رو مدیریت می‌کنه:

```go
type model struct {
    session   *ui.Session
    splash    ui.Component
    startup   ui.Component
    workspace ui.Component
    stage     AppStage
    // dialog stack
}

func New(uiLogger, coreLogger logger.Logger) Application
```

برنامه از سه مرحله می‌گذره:

| مرحله | کامپوننت | توضیح |
|---|---|---|
| `StageSplash` | `splash` | لوگوی ASCII با انیمیشن Glitch و نمایش نسخه؛ با فشردن هر کلید می‌شه ازش گذشت |
| `StageStartUP` | `startup` | اجرای ترتیبی بررسی‌های اولیه: Logger ← Config ← Xray ← DNSTT ← Slipstream ← VayDNS ← MasterDNS ← StormDNS ← TheFeed ← App |
| `StageWorkspace` | `workspace` | محیط اصلی برنامه شامل Body، Footer و Stack دیالوهای Overlay |

`ui.Session` و `ui.Env` جای ساختار تخت قبلی `AppState` رو گرفتن. Session وابستگی‌های مشترک رو در اختیار کامپوننت‌های مختلف قرار می‌ده:

```go
type Session struct {
    Config  *config.ScannerConfig
    Store   *config.Store
    Program ui.Program
    // env and core logger
}

type Env struct {
    Layout   layout.Layout
    Theme    theme.Theme
    Keys     *keys.Keys
    Log      logger.Logger // ui.log
    Settings UISettings
}
```

کامپوننت‌ها تنظیمات رو از `session.Config` می‌خونن و تغییرات رو از طریق `session.Store` ذخیره می‌کنن. تم، Layout، کلیدها و تنظیمات رابط هم از `session.Env()` در دسترس هستن. برای دسترسی به Config از Singleton سراسری استفاده نمی‌شه.

### مسیریابی پیام‌ها

نحوهٔ پردازش پیام‌ها به مرحلهٔ فعلی برنامه بستگی داره:

1. **Splash:** انیمیشن اجرا می‌شه تا به پایان برسه یا کاربر کلیدی بزنه؛ بعد برنامه وارد Startup می‌شه.
2. **Startup:** پیام‌های بررسی از طریق کانال `logCh` و از یک Goroutine دریافت می‌شن. خطاهای بحرانی ادامهٔ بررسی‌ها رو متوقف می‌کنن. با زدن Enter، برنامه وارد Workspace می‌شه.
3. **Workspace:** پیام‌های رابط در این مرحله پردازش می‌شن:

   - `tea.WindowSizeMsg`: با تغییر اندازهٔ ترمینال، `layout.Update(width, height)` دوباره اجرا می‌شه.
   - `tea.KeyPressMsg`: کلید `ctrl+c` فوراً برنامه رو می‌بنده؛ `ctrl+t` هم اطلاعات Goroutineها رو در سطح Debug داخل `ui.log` ثبت می‌کنه.
   - **مدیریت Overlay:** وقتی Dialog بازه، بالاترین Dialog ورودی‌های کیبورد رو دریافت می‌کنه. کلیدهای Back و Quit برای بستن اون استفاده می‌شن.
   - **ناوبری در ریشه:** در منوی اصلی، `b` یا `q` به‌جای خروج فوری، Dialog تأیید خروج رو باز می‌کنن.
   - `dialog.OpenDialogMsg`: یک Overlay جدید به‌همراه اطلاعات موقعیتش به Stack اضافه می‌کنه.
   - `ui.CloseComponentMsg`: Overlay متناظر با ID رو از Stack برمی‌داره و `OnClose` اون رو اجرا می‌کنه.
   - **ارسال پیام‌های عمومی:** اگر Overlayی پیام رو دریافت نکنه، پیام به Body و Footer فرستاده می‌شه.

### نحوهٔ نمایش

در مرحلهٔ **Workspace**، چیدمان برنامه به این صورته:

```text
┌─────────────────────────────────┐
│                                 │
│           body                  │  ← component stack (menu → screen)
│                                 │
├─────────────────────────────────┤
│           footer                │  ← version · status · goroutines · mem
└─────────────────────────────────┘
    + overlay dialogs (centered/positioned)
```

در نسخهٔ 2.12 هدر جداگانه‌ای وجود نداره و Banner داخل منوی اصلی نمایش داده می‌شه. Overlayها با استفاده از `bubbletea-overlay` روی محتوای اصلی قرار می‌گیرن.

حداقل اندازهٔ قابل‌قبول ترمینال ۷۵×۳۵ هست. اگر اندازه از این مقدار کمتر بشه، پیام `terminal too small` نمایش داده می‌شه.

---

## صفحهٔ Splash

پکیج `ui/main/splash` لوگوی ASCII متحرک «BGSCAN» رو با افکت Glitch نمایش می‌ده:

- ۸۰ فریم انیمیشن با فاصلهٔ ۳۰ میلی‌ثانیه‌ای بین Tickها، سپس ۱۵۰ فریم ثابت
- جایگزین‌شدن بعضی کاراکترهای لوگو با کاراکترهای Glitch مثل `▓▒░╬╪┼╫┃━╳▄▀`؛ احتمال این جایگزینی در طول انیمیشن کمتر می‌شه
- نمایش یک هایلایت افقی با رنگ Accent در ۷۲٪ پیشرفت
- لرزش افقی و ایجاد ردیف‌های تصویری معیوب در طول انیمیشن
- نمایش نسخه با افکت Fade-in بعد از پایان انیمیشن

---

## بررسی‌های اولیهٔ Startup

پکیج `ui/main/startup` ده بررسی رو به‌ترتیب و در یک Goroutine اجرا می‌کنه. وضعیت هر بررسی در نوار کناری با نشانگرهای Pending، Running، Success، Warn و Error نمایش داده می‌شه. نوار پیشرفت و بخش قابل‌اسکرول هم جزئیات بررسی‌ها رو نشون می‌دن.

| # | دسته | چه چیزی بررسی می‌شه؟ |
|---|---|---|
| 1 | Logger | لاگرهای Core و UI بررسی می‌شن و Schemaهای داخلی Probe از طریق `core.Init()` ثبت می‌شن |
| 2 | Config | تنظیمات بارگذاری می‌شن، `NormalizeAll` اجرا می‌شه و مقادیری که اصلاح شدن گزارش می‌شن |
| 3 | Xray | نسخهٔ هستهٔ داخلی نمایش داده می‌شه و Outboundها فهرست و اعتبارسنجی می‌شن |
| 4 | DNSTT | همهٔ فایل‌های تنظیمات DNSTT اعتبارسنجی می‌شن |
| 5 | Slipstream | کتابخانهٔ کلاینت پیدا و بارگذاری می‌شه، بررسی می‌شه و تنظیمات Slipstream اعتبارسنجی می‌شن |
| 6 | VayDNS | همهٔ فایل‌های تنظیمات VayDNS اعتبارسنجی می‌شن |
| 7 | MasterDNS | همهٔ فایل‌های تنظیمات MasterDNS اعتبارسنجی می‌شن |
| 8 | StormDNS | همهٔ فایل‌های تنظیمات StormDNS اعتبارسنجی می‌شن |
| 9 | TheFeed | همهٔ فایل‌های تنظیمات TheFeed اعتبارسنجی می‌شن |
| 10 | App | برنامه منتظر می‌مونه تا کاربر Enter بزنه |

خطاهای بحرانی باعث توقف بررسی‌های بعدی می‌شن. با کلیدهای `j` و `k` می‌تونید در خروجی اسکرول کنید و بعد از پایان بررسی‌ها با زدن Enter وارد Workspace بشید.

---

## Body و Stack کامپوننت‌ها

پکیج `ui/main/body` یک Stack از کامپوننت‌ها نگه می‌داره. کامپوننت بالای Stack صفحهٔ فعاله. Body فقط در مرحلهٔ `workspace` فعاله و در مراحل Splash و Startup استفاده نمی‌شه.

```text
components[0] = mainMenu (entry)    ← همیشه در Stack هست
components[1] = scan source picker  ← روی Stack قرار می‌گیره
components[2] = scan type picker   ← روی Stack قرار می‌گیره
components[3] = scanner view       ← روی Stack قرار می‌گیره
```

**پیام‌های ناوبری**

| پیام | عملکرد |
|---|---|
| `ui.OpenComponentMsg` | یک کامپوننت جدید به بالای Stack اضافه می‌کنه |
| `ui.CloseComponentMsg` | کامپوننت مشخص‌شده با ID رو برمی‌داره |
| `ui.ResetComponentStacksMsg` | Stack رو تا منوی اصلی خالی می‌کنه |
| کلید Back (`b` / `Esc`) | کامپوننت بالایی رو برمی‌داره، به‌شرط اینکه Stack بیشتر از یک کامپوننت داشته باشه |
| کلید Quit (`q`) | Dialog تأیید خروج رو باز می‌کنه |

فقط کامپوننت بالای Stack پیام `Update` دریافت می‌کنه. وقتی این کامپوننت برداشته بشه، کامپوننت بعدی فعال می‌شه و نامش در نوار وضعیت Footer نمایش داده می‌شه.

---

## سیستم Dialogهای Overlay

پکیج `ui/shared/dialog` امکان نمایش Dialog روی Layout اصلی رو فراهم می‌کنه.

```go
type OpenDialogMsg struct {
    Component ui.Component
    XPos      DialogPosition // Top, Right, Bottom, Left, Center
    YPos      DialogPosition
    XOffset   int
    YOffset   int
}
```

برای تعیین موقعیت Dialogها از کتابخانهٔ `rmhubbert/bubbletea-overlay` استفاده می‌شه. بیشتر Dialogها وسط صفحه نمایش داده می‌شن، اما File Picker و بعضی Dialogهای تأیید ممکنه موقعیت متفاوتی داشته باشن.

Stack دیالوگ‌ها در `app.model` امکان بازبودن چند Overlay روی هم رو فراهم می‌کنه. بالاترین Dialog همهٔ ورودی‌ها رو دریافت می‌کنه. باز و بسته‌شدن Dialogها هم انیمیشن داره؛ هنگام بسته‌شدن، اجرای `OnClose` تا پایان انیمیشن به تعویق می‌افته.

اگر گزینهٔ **Animation** رو در تنظیمات رابط خاموش کنید، انیمیشن Dialogها و سایر انیمیشن‌های رابط غیرفعال می‌شن.

---

## حالت‌های ورودی

فایل `ui/shared/ui/mode.go` سه حالت برای کنترل رفتار کلیدهای کیبورد تعریف می‌کنه:

| حالت | کاربرد | کلیدهای Back | نکته |
|---|---|---|---|
| `KeyModeNormal` | ناوبری عادی | `b`، `Esc` | کلید `q` تأیید خروج رو باز می‌کنه؛ در منوی اصلی، `b` و `Esc` هم همین کار رو انجام می‌دن |
| `KeyModeInput` | تایپ در فیلد متنی یا Picker | `Esc` | وقتی فیلد در حالت ورودی باشه، `q` و `b` به‌عنوان متن در نظر گرفته می‌شن |
| `KeyModeManage` | صفحه‌هایی با کلیدهای اختصاصی، مثل جدول‌ها و اسکنر | بسته به کامپوننت | اسکنر از `Esc`، `b` یا `q` برای بازکردن تأیید لغو استفاده می‌کنه |

تعریف کلیدها در `ui/shared/keys/defaults.go` قرار داره و از طریق `session.Env().Keys` در اختیار کامپوننت‌هاست. کامپوننت‌ها به‌جای مقایسهٔ مستقیم رشتهٔ کلیدها، از Actionهای تعریف‌شده استفاده می‌کنن. به این ترتیب رفتار کلیدهای Back، Quit، ناوبری، جدول و ورودی در بخش‌های مختلف یکسان می‌مونه.

Footer و راهنمای Dialogها هم از همین تعریف‌ها استفاده می‌کنن. با زدن `?` می‌تونید فهرست کلیدها رو ببینید. کلید `Ctrl+C` مستقل از حالت فعلی و در سطح اپ، برنامه رو فوراً می‌بنده.

### هشدار چیدمان کیبورد

اگر در حالت `KeyModeNormal` یک کاراکتر غیرلاتین به Workspace برسه، bgscan اعلان **Keyboard Layout** رو نمایش می‌ده و توضیح می‌ده که میان‌برها به چیدمان انگلیسی نیاز دارن.

این هشدار در حالت‌های Input و Manage، هنگام بازبودن Dialog یا خارج از Workspace نمایش داده نمی‌شه. همچنین اعلان‌های تکراری روی هم جمع نمی‌شن.

---

## Layout

فایل `ui/shared/layout/layout.go` محاسبهٔ اندازه و موقعیت بخش‌های رابط رو انجام می‌ده. با هر تغییر اندازهٔ ترمینال، ابعاد ناحیه‌های مختلف دوباره محاسبه می‌شن.

```go
type Layout struct {
    terminal TerminalSize // ابعاد واقعی ترمینال
    content  ContentSize  // فضای قابل رسم پس از حاشیه‌ها و Padding
    body     ComponentSize
    footer   ComponentSize
}
```

در نسخهٔ 2.12، Workspace فقط دو بخش قابل رسم داره: Body و Footer. هدر قدیمی و پکیج `ui/main/header` دیگه استفاده نمی‌شن.

با هر `tea.WindowSizeMsg`، متد `Layout.Update(width, height)` اجرا می‌شه. کامپوننت‌ها هم Layout مشترک رو از `session.Env().Layout` می‌گیرن.

اندازهٔ پیش‌فرض ترمینال ۸۰×۳۵ و حداقل اندازهٔ قابل‌قبول ۷۵×۳۵ هست.

---

## Widgetهای قابل استفادهٔ مجدد (`basic`)

پوشهٔ `ui/components/basic/` شامل Widgetهایی هست که در صفحه‌های مختلف استفاده می‌شن:

| Widget | کاربرد |
|---|---|
| `confirm` | Dialog تأیید بله/خیر برای کارهایی مثل خروج و حذف |
| `crud` | اینترفیس عمومی CRUD برای جدول‌هایی مثل فهرست IPها و Outboundها، با استفاده از انتزاع `Provider` |
| `input` | پوستهٔ Dialog ورودی و Widgetهای تایپ‌دار مثل textinput، textarea، selectinput، multiselect و toggleinput |
| `inspector` | ویرایشگر عمومی کلید-مقدار برای تنظیمات، همراه با قالب‌بندی فیلدها و آداپترهای نوع |
| `logview` | نمایش زنده و قابل‌اسکرول لاگ‌های Core و UI |
| `menu` | منوی عمودی قابل‌کنترل با کیبورد |
| `notice` | اعلان موقت در قالب Banner |
| `picker` | Dialog انتخاب فایل |
| `progress` | نوار پیشرفت انیمیشنی همراه با ETA و نمایش نرخ |
| `table` | جدول داده با قابلیت مرتب‌سازی و اسکرول |
| `tabs` | محفظهٔ Tabبندی‌شده؛ نمای اسکنر برای نمایش هر Stage از این Widget استفاده می‌کنه |

هر Widget اینترفیس استاندارد `Component` رو پیاده‌سازی می‌کنه و می‌تونه در صفحه‌های بزرگ‌تر به‌کار بره.

---

## Inspectorهای تنظیمات و فرم‌ها

### Inspectorهای تنظیمات (`ui/components/inspector/`)

| فرم | فایل تنظیمات |
|---|---|
| `general` | `general_settings.toml` |
| `icmp` | `icmp_settings.toml` |
| `tcp` | `tcp_settings.toml` |
| `http` | `http_settings.toml` |
| `dns` | `dns_settings.toml` |
| `xray` | `xray_settings.toml` |
| `ui` | `ui_settings.toml` |

هر Inspector بخش مربوط به خودش رو از `session.Config` می‌خونه، فیلدهای قابل ویرایش رو نمایش می‌ده و هنگام ذخیره، متد متناظر `session.Store.SaveXxx()` رو صدا می‌زنه.

Inspector تنظیمات UI از `session.Env().Settings` استفاده می‌کنه و تغییرات رو با `ui.SaveSettings()` ذخیره می‌کنه.

فیلدها می‌تونن به انتخاب‌های قبلی وابسته باشن. مثلاً تنظیمات TLS فقط وقتی نمایش داده می‌شن که HTTPS انتخاب شده باشه؛ فیلدهای مخصوص تونل هم فقط برای پروتکل مربوطه ظاهر می‌شن.

هر فیلد پیش از ثبت اعتبارسنجی می‌شه. Inspector تغییرات رو روی یک کپی از بخش تنظیمات اعمال می‌کنه و `validate.ValidateXxx` رو روی اون اجرا می‌کنه. اگر فیلد نامعتبر باشه، تغییر رد می‌شه و مقدار نامعتبر وارد Config یا فایل تنظیمات نمی‌شه.

Widget عمومی `basic/inspector` مسئول نمایش فیلدها، قالب‌بندی و تبدیل نوعه. Inspector هر پروتکل فقط مشخص می‌کنه چه فیلدهایی نمایش داده بشن و چطور به Struct تنظیمات نگاشت بشن.

### فرم‌های تنظیمات تونل DNS (`ui/components/form/`)

| فرم | پروتکل | Struct تنظیمات |
|---|---|---|
| `dnstt` | DNSTT | `dns.DNSTTConfig` |
| `slipstream` | Slipstream | `dns.SlipstreamConfig` |
| `vaydns` | VayDNS | `dns.VayDNSConfig` |
| `masterdns` | MasterDNS | `dns.MasterDNSConfig` |
| `stormdns` | StormDNS | `dns.StormDNSConfig` |
| `thefeed` | TheFeed | `dns.TheFeedConfig` |

هر فرم Inspector مخصوص پروتکل خودش رو می‌سازه و فیلدهای مربوط به اتصال رو نمایش می‌ده. گزینه‌های پیشرفته فقط در فرم VayDNS وجود دارن و تنظیمات پروکسی و احراز هویت در فرم‌های DNSTT، VayDNS و Slipstream ارائه می‌شن.

این فرم‌ها هم برای ساخت Config جدید و هم برای ویرایش Config موجود استفاده می‌شن. هنگام ذخیره، متد `SaveConfig` یا `EditConfig` سرویس مربوطه فراخوانی می‌شه. سازنده‌های فیلدها هم در `ui/components/form/formkit` قرار دارن.

---

## منوها و جدول‌ها

### منوها (`ui/components/menus`)

| منو | محل دسترسی | کاربرد |
|---|---|---|
| `entry` | شروع برنامه | منوی اصلی شامل Run Scan، IP Files، Result Files، Xray Outbound، DNS Tunneling، Settings و Logs |
| `targetsource` | Run Scan | انتخاب ورودی اسکن از فهرست IPها یا نتایج قبلی |
| `scantype` | بعد از انتخاب ورودی | انتخاب نوع اسکن: ICMP، TCP، HTTP، Xray، DNS Resolve یا DNS Tunneling |
| `settings` | گزینهٔ Settings | انتخاب بخش تنظیمات |
| `outboundmenu` | گزینهٔ Xray Outbound | مدیریت Outboundها |
| `dnstunmenu` | نوع اسکن DNS Tunneling | انتخاب پروتکل از بین DNSTT، Slipstream، VayDNS، MasterDNS، StormDNS و TheFeed |
| `logs` | گزینهٔ Logs | انتخاب لاگ‌های Core یا UI |

### جدول‌ها (`ui/components/tables`)

| جدول | منبع داده | Provider |
|---|---|---|
| `iplist` | پوشهٔ `ips/` | `iplist.Registry` |
| `resultlist` | پوشهٔ `result/` | `result.GetResultFiles` |
| `outbounds` | تنظیمات Outbound در Xray | پکیج `xray` |
| `dnstun` | تنظیمات تونل DNS | `dns.GetAllDNSTunsFile` |
| `ipviewer` | مجموعهٔ نتایج موجود در حافظه | کامپوننت اسکنر |

جدول‌ها از Widget عمومی `basic/table` و اینترفیس `Provider` استفاده می‌کنن. Provider ردیف‌ها رو فراهم می‌کنه و مرتب‌سازی ستون‌ها رو انجام می‌ده.

مرورگر نتایج ابتدا اطلاعات کلی فایل‌ها رو نشون می‌ده. وقتی فایلی رو باز می‌کنید، `ipviewer` ستون‌ها رو بر اساس Schema متصل به فایل می‌سازه؛ بنابراین نتایج Probeهای ثبت‌شده بدون نیاز به تغییر کد جدول قابل نمایش هستن.

Schemaهای MasterDNS و StormDNS در Registry ثبت نمی‌شن؛ به همین دلیل فایل‌های نتایج این دو پروتکل در مرورگر نتایج نمایش داده نمی‌شن. جدول `dnstun` هم امکان مدیریت Configهای هر شش پروتکل تونل DNS رو فراهم می‌کنه.

---

## نمای اسکنر

پکیج `ui/components/scanner` داشبورد زندهٔ اسکنه و اسکنر Core رو به رابط کاربری وصل می‌کنه.

امکانات این بخش:

- **نوار Tab مراحل اسکن:** وضعیت هر Stage رو به‌صورت زنده با حالت‌هایی مثل `waiting`، `scanning`، `paused`، `done` و `error` نمایش می‌ده.
- **نوار پیشرفت:** تعداد IPهای پردازش‌شده و کل، زمان تخمینی باقی‌مونده (ETA)، نرخ پردازش برحسب IP/s و یک Sparkline از نرخ‌های اخیر رو نشون می‌ده.
- **پنل‌های Scan و Errors:** پنل Scan نرخ برخورد رو نمایش می‌ده. پنل Errors هم خطاها رو به دسته‌هایی مثل timeout، refused، unreachable، closed، badresponse و سایر خطاها تقسیم می‌کنه و تعداد و درصد هر دسته رو نشون می‌ده.
- **جدول نتایج هر Stage:** از `tables/ipviewer` استفاده می‌کنه و ستون‌ها رو بر اساس Schema همون Stage می‌سازه. مقادیر زمان و پهنای باند هم بر اساس کیفیت با رنگ‌های متفاوت نمایش داده می‌شن.
- **وضعیت مستقل هر Stage:** وضعیت هر مرحله از `Waiting` به `PreProcessing`، سپس `Scanning` و در پایان به `Ended` یا `Error` تغییر می‌کنه.
- **به‌روزرسانی زنده:** اطلاعات پیشرفت از `engine.ScanHooks.OnProgress` دریافت می‌شن و نتایج موفق به‌صورت دسته‌ای وارد جدول می‌شن.

کلیدهای اصلی هنگام اسکن:

- `p` یا `space`: توقف موقت و ادامهٔ اسکن
- `l`: بازکردن نمایشگر لاگ Core
- `Tab`: جابه‌جایی بین Stageها
- `Esc`، `b` یا `q`: بازکردن Dialog تأیید لغو اسکن

---

## تم

فایل `ui/theme/theme.go` پالت رنگی مرکزی رابط رو تعریف می‌کنه:

```go
type Theme struct {
    Primary       color.Color
    Secondary     color.Color
    Border        color.Color
    BorderActive  color.Color
    Text          color.Color
    Muted         color.Color
    Timestamp     color.Color
    Info          color.Color
    Error         color.Color
    Success       color.Color
    Orange        color.Color
    Yellow        color.Color
    Purple        color.Color
    ProgressStart color.Color
    ProgressEnd   color.Color
}
```

### تم‌های داخلی

پکیج `ui/theme` شامل ۲۱ تم داخلیه: ۱۳ تم تیره و ۸ تم روشن. نام تم از گزینهٔ `theme` در `settings/ui_settings.toml` خونده می‌شه.

اگر این گزینه خالی باشه، برنامه با استفاده از متغیر محیطی `$COLORFGBG` پس‌زمینهٔ ترمینال رو تشخیص می‌ده و تم `bgscan-dark` یا `bgscan-light` رو انتخاب می‌کنه. اگر نام تم ناشناخته باشه، تنظیمات به حالت انتخاب خودکار برمی‌گرده.

فهرست تم‌ها:

`bgscan-dark`، `tokyo-night`، `gruvbox-dark`، `one-dark`، `dracula`، `monokai`، `solarized-dark`، `rose-pine`، `kanagawa`، `github-dark`، `material-dark`، `everforest-dark`، `ayu-dark`، `bgscan-light`، `gruvbox-light`، `one-light`، `solarized-light`، `rose-pine-dawn`، `everforest-light`، `ayu-light` و `github-light`.

کامپوننت‌ها رنگ‌ها رو از `session.Env().Theme` می‌گیرن و نباید رنگی رو مستقیماً داخل کد ثابت کنن. متد `Session.SetTheme` تم انتخاب‌شده رو ذخیره می‌کنه و پیام `theme.ThemeChanged` رو می‌فرسته تا فرم‌ها و Widgetهای Cacheشده دوباره ساخته بشن.

هر تم رنگ‌های `TierExcellent`، `TierGood`، `TierFair` و `TierPoor` رو هم تعریف می‌کنه. مقادیر ناموفق یا خارج از محدوده از رنگ‌های Muted و Error همون تم استفاده می‌کنن. IP Viewer از این رنگ‌ها برای نمایش کیفیت تأخیر و پهنای باند استفاده می‌کنه.

### هماهنگی با huh

فایل `ui/theme/huh.go` یک `huh.Theme` بر اساس پالت مرکزی می‌سازه. به این ترتیب فیلدهای فرم مثل select، confirm و text input که کتابخانهٔ `huh` رسم می‌کنه، با ظاهر بقیهٔ رابط هماهنگ می‌مونن. Widgetهای ورودی و Inspector هم از همین تم استفاده می‌کنن.

---

## عیب‌یابی

### گرفتن Dump از Goroutineها

با زدن `Ctrl+T`، تابع `dumpGoroutines()` در `app/update.go` اجرا می‌شه و اطلاعات زیر رو ثبت می‌کنه:

- تعداد کل Goroutineها رو در لاگ می‌نویسه.
- پروفایل کامل Goroutineها رو از طریق `runtime/pprof` در سطح Debug داخل `ui.log` ثبت می‌کنه.
- خط‌هایی رو که شامل `[chan receive]`، `[chan send]`، `[select]`، `[IO wait]` یا `[sleep]` هستن جدا می‌کنه و لاگ می‌گیره. این اطلاعات برای پیدا کردن Workerهای گیرکرده یا Probeهایی که منابعشون آزاد نشده مفیده.

چون این اطلاعات در سطح `Debug` ثبت می‌شن، برای مشاهده‌شون مقدار زیر رو در `settings/ui_settings.toml` قرار بدید:

```toml
log_level = "debug"
```

### آمار زمان اجرا

کامپوننت Footer هر ثانیه آمار Runtime زبان Go رو می‌خونه:

- `runtime.NumGoroutine()`: تعداد Goroutineها
- `runtime.ReadMemStats()`: اطلاعات حافظهٔ Heap و Sys

این مقادیر در نوار وضعیت نمایش داده می‌شن و برای بررسی مصرف منابع هنگام اسکن کاربرد دارن.

---

## صفحه‌های مرتبط

- [معماری](./architecture/): آشنایی با ساختار کلی پروژه
- [Core](./core/): آشنایی با موتور اسکن و اینترفیس Probe

