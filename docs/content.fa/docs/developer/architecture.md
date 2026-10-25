---
title: "معماری"
weight: 2
---

# معماری

BGScan ساختار لایه‌ای داره. نقطه ورود برنامه، `main`، کنترل رو به رابط کاربری مبتنی بر Bubble Tea می‌سپاره. رابط کاربری سه مرحله داره: Splash، Startup و Workspace.

در مرحله **Startup**، لاگرها بررسی می‌شن، Schemaهای نتایج ثبت می‌شن، تنظیمات بارگذاری و اعتبارسنجی می‌شن و وضعیت هسته داخلی Xray، فایل‌های تنظیمات تونل و کتابخانه اختیاری Slipstream بررسی می‌شه.

بعد از اینکه Startup با موفقیت تموم شد، برنامه وارد Workspace می‌شه. از اینجا می‌تونید اسکن رو اجرا کنید. موتور اسکن از چند مرحله تشکیل شده و می‌شه Probeها و Writerهای اون رو تغییر داد.

## چیدمان پوشه‌ها

```text
.
├── assets
│   ├── dns-tunneling
│   │   ├── dnstt
│   │   ├── masterdns
│   │   ├── slipstream
│   │   ├── stormdns
│   │   ├── thefeed
│   │   └── vaydns
│   ├── xray
│   │   └── outbounds
│   └── slipstream-client
├── cmd
│   └── bgscan
│       └── main.go
├── docs
├── internal
│   ├── core
│   │   ├── core.go          schema registration
│   │   ├── config
│   │   │   └── validate
│   │   ├── dns
│   │   ├── ffi
│   │   ├── fileutil
│   │   ├── iplist
│   │   ├── netutil
│   │   ├── process
│   │   ├── result
│   │   ├── scanner
│   │   │   ├── engine
│   │   │   ├── portmgr
│   │   │   └── probe
│   │   │       ├── dnsttprobe
│   │   │       ├── httpprobe
│   │   │       ├── icmpprobe
│   │   │       ├── masterdnsprobe
│   │   │       ├── resolveprobe
│   │   │       ├── slipstreamprobe
│   │   │       ├── stormdnsprobe
│   │   │       ├── tcpprobe
│   │   │       ├── thefeedprobe
│   │   │       ├── vaydnsprobe
│   │   │       └── xrayprobe
│   │   ├── socks
│   │   ├── ssh
│   │   ├── speedtest
│   │   └── xray
│   ├── logger
│   └── ui
│       ├── components
│       ├── main
│       │   ├── app
│       │   ├── body
│       │   ├── footer
│       │   ├── splash
│       │   ├── startup
│       │   └── workspace
│       ├── shared
│       └── theme
├── ips
├── scripts
└── settings
```

## نمای کلی لایه‌ها

```text
┌─────────────────────────────────────────────┐
│                 cmd/bgscan                   │
│   theme.Init → app.New → tea.Run             │
├─────────────────────────────────────────────┤
│                internal/ui                   │
│   main                                       │
│     splash  startup  workspace               │
│     app (root model, stage transitions)      │
│     body, footer                             │
│   components (menus, tables, inspector)      │
│   shared (layout, env, dialog, ui, valid.)   │
│   theme                                      │
│                                              │
│   ui/main/startup performs (inside the TUI): │
│     logger, core.Init, config, shared-lib    │
│     and tunnel-config checks                 │
├─────────────────────────────────────────────┤
│               internal/core                  │
│  ┌──────────┐ ┌────────┐ ┌───────────────┐  │
│  │ scanner  │ │ config │ │    result     │  │
│  │ engine   │ │ Store  │ │ writer+schema │  │
│  │ probe    │ │validate│ │   registry    │  │
│  │ portmgr  │ └────────┘ └───────────────┘  │
│  └──────────┘                                │
│  ┌────────┐ ┌─────┐ ┌──────┐ ┌───────────┐  │
│  │ iplist │ │ dns │ │ xray │ │ speedtest │  │
│  └────────┘ └─────┘ └──────┘ └───────────┘  │
│  ┌─────────┐ ┌──────────┐ ┌─────────────┐   │
│  │ netutil │ │ process  │ │  fileutil   │   │
│  └─────────┘ └──────────┘ └─────────────┘   │
├─────────────────────────────────────────────┤
│              internal/logger                 │
├─────────────────────────────────────────────┤
│          assets/  ips/  settings/           │
└─────────────────────────────────────────────┘
```

## معرفی پوشه‌ها

| مسیر | توضیح |
|---|---|
| `cmd/bgscan` | نقطه ورود برنامه. لاگرها رو می‌سازه، اپلیکیشن رو همراه با تم و تنظیمات رابط کاربری راه‌اندازی می‌کنه و اجرای برنامه رو به Bubble Tea می‌سپاره. |
| `internal/core/core.go` | تابع `core.Init()` نُه Schema مربوط به Probeها رو در `result.DefaultRegistry` ثبت می‌کنه. Schemaهای MasterDNS و StormDNS تعریف شدن، اما ثبت نمی‌شن. این تابع در مرحله Startup اجرا می‌شه. |
| `internal/core/config` | تعریف انواع `ScannerConfig`، پیش‌فرض‌های کامپایل‌شده و `Store` برای خواندن و نوشتن فایل‌های `settings/*.toml`. |
| `internal/core/config/validate` | اعتبارسنج‌ها و Normalizerهای بخش‌های مختلف تنظیمات که در `aggregate.go` کنار هم قرار گرفتن. |
| `internal/core/scanner` | اینترفیس `Scanner`، ساختار `StageConfig` و توابع ساخت Stageها. |
| `internal/core/scanner/engine` | اجرای مراحل اسکن در حالت‌های Sequential، Streaming و Batch، به‌همراه مدیریت Pause. همه اسکن‌ها از این بخش عبور می‌کنن. |
| `internal/core/scanner/probe` | اینترفیس `Probe` و یک زیربسته برای هر نوع Probe. |
| `internal/core/scanner/portmgr` | مدیریت اجاره پورت محلی برای Probeهای تونل و Xray. |
| `internal/core/result` | اینترفیس `Result`، Schemaها، Registry، Writer غیرهم‌زمان، ادغام فایل‌های CSV و Loader. |
| `internal/core/iplist` | خواندن و Parse فهرست IPها، Registry، Shuffle و پردازش جریانی ورودی‌ها. |
| `internal/core/netutil` | نرمال‌سازی Host، Parse نسخه TLS و استخراج SNI برای Probeهای HTTP. |
| `internal/core/dns` | کوئری‌های DNS، ترنسپورت‌ها، سرویس تنظیمات تونل‌های DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS و TheFeed، همراه با مدیریت تونل‌ها. |
| `internal/core/socks` | کلاینت SOCKS5 برای اعتبارسنجی تونل. |
| `internal/core/ssh` | کلاینت SSH برای اتصال‌های تونل‌شده از طریق پروکسی SSH. |
| `internal/core/xray` | سرویس Xray که در همان پروسه اجرا می‌شه، مدیریت تنظیمات ورودی و خروجی و Parse لینک‌های اشتراک. |
| `internal/core/speedtest` | اندازه‌گیری تأخیر، سرعت دانلود و آپلود، مورد استفاده Probe مربوط به Xray. |
| `internal/core/process` | ابزارهای قدیمی مدیریت پروسه برای چند پلتفرم که اسکنر فعلی ازشون استفاده نمی‌کنه. |
| `internal/core/ffi` | اتصال از طریق purego به کتابخانه کلاینت داخلی Slipstream. |
| `internal/core/fileutil` | ابزارهای کمکی برای CSV، JSON، TOML، متن، فایل‌های موقت، مرتب‌سازی و مدیریت مسیرها. |
| `internal/logger` | دو جریان لاگ سطح‌بندی‌شده در `core.log` و `ui.log`، با چرخش فایل به کمک lumberjack و امکان دریافت زنده لاگ‌ها. |
| `internal/ui/main/app` | مدل اصلی Bubble Tea که مراحل Splash، Startup و Workspace رو مدیریت می‌کنه. |
| `internal/ui/main/splash` | صفحه شروع انیمیشنی که لوگوی ASCII و نسخه برنامه رو نمایش می‌ده. |
| `internal/ui/main/startup` | بررسی‌های اولیه به‌ترتیب اجرا می‌شن: Logger، Config، Xray، DNSTT، Slipstream، VayDNS، MasterDNS، StormDNS، TheFeed و App. وضعیت هر بررسی به‌صورت زنده نمایش داده می‌شه. |
| `internal/ui/main/workspace` | پوسته اصلی Workspace که بعد از موفقیت Startup نمایش داده می‌شه و Body و Footer رو در بر می‌گیره. |
| `internal/ui/main/body` | مدیریت صفحات داخل Workspace، مثل منوی اصلی، اسکن، تنظیمات و لاگ‌ها. |
| `internal/ui/main/footer` | نوار وضعیت پایین رابط کاربری. Workspace هدر جداگانه‌ای نداره. |
| `internal/ui/components` | کامپوننت‌های رابط کاربری، Inspectorهای تنظیمات، منوها، جدول‌ها و نمای زنده اسکن. |
| `internal/ui/shared` | ابزارهای Layout، سیستم Dialog، مدیریت کلیدها، اینترفیس کامپوننت‌ها و اعتبارسنجی. |
| `internal/ui/theme` | تم‌های روشن و تیره داخلی و آداپتر تم فرم‌های huh. |
| `assets/xray` | داده‌های مسیریابی Xray و Outboundها. باینری Xray یا پوشه تنظیمات تولیدشده در این مسیر قرار نداره، Xray با هسته داخلی و در همان پروسه اجرا می‌شه. |
| `assets/slipstream-client` | کتابخانه کلاینت Slipstream که با purego بارگذاری می‌شه، پروسه جداگانه‌ای اجرا نمی‌کنه. |
| `assets/dns-tunneling` | فایل‌های تنظیمات تونل DNS برای DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS و TheFeed. |
| `ips` | فهرست‌های IP ارائه‌دهنده‌ها در قالب CSV. |
| `settings` | فایل‌های تنظیمات `.toml` که برنامه ازشون استفاده می‌کنه. |
| `scripts` | اسکریپت‌های کمکی برای نصب، Build و انتشار نسخه‌ها. |

## جریان اجرای برنامه

```text
main()
  │
  ├─ logger.New("core.log"), logger.New("ui.log")
  │
  ├─ config.AppVersion = Version    recorded for the result files
  │
  ├─ app := app.New()               theme.Init + splash/startup/workspace
  ├─ p := tea.NewProgram(app)       hand off to BubbleTea
  │
  └─ inside the TUI:
        │
        ├─ StageSplash              animated ASCII logo + version
        │
        ├─ StageStartup             ui/main/startup runs each check
        │     │                     in a goroutine, reports status live
        │     │
        │     ├─ Logger             verify loggers, then core.Init()
        │     │                     to register probe schemas
        │     ├─ Config             store.Load(), validate.NormalizeAll,
        │     │                     report any clamped values
        │     ├─ Xray               report embedded core version,
        │     │                     validate outbound templates
        │     ├─ DNSTT              validate config files
        │     ├─ Slipstream         locate and load the client library,
        │     │                     validate configs
        │     ├─ Vaydns             validate config files
        │     ├─ MasterDNS          validate config files
        │     ├─ StormDNS           validate config files
        │     ├─ TheFeed            validate config files
        │     └─ App                wait for Enter
        │
        └─ StageWorkspace           main workspace:
              ├─ body               component stack: main menu → scan / settings / logs
              └─ footer

on Run Scan:
  ├─ scanner.NewScanner(ctx, input)
  ├─ AddStage(BuildICMPStage) ...
  ├─ Run()
  │     └─ engine.RunScanWithChain
  │          sequential | streaming | batch
  └─ result.Writer → CSV merge → disk
```

## تصمیم‌های اصلی طراحی

**تنظیمات از طریق ساختارها منتقل می‌شن، نه متغیرهای Global.**

برای Config از Singleton استفاده نمی‌شه. مرحله Startup یک `config.Store` می‌سازه، `store.Load()` رو اجرا می‌کنه و مقادیر `*ScannerConfig` و `*Store` رو در اختیار `ui.Session` قرار می‌ده تا کامپوننت‌ها و اسکنر بهشون دسترسی داشته باشن. تست‌ها هم با `WithSettingsDir` یک `Store` روی پوشه موقت می‌سازن. بنابراین تنظیمات واقعی سیستم تغییر نمی‌کنن.

**تنظیمات نامعتبر به‌صورت خودکار اصلاح می‌شن.**

تابع `validate.NormalizeAll` مقادیری رو که خارج از محدوده مجاز هستن به مقدار پیش‌فرض برمی‌گردونه و اصلاحات رو در نوار وضعیت Startup نشون می‌ده. در نتیجه، یک مقدار اشتباه در فایل تنظیمات معمولاً به‌جای متوقف کردن برنامه، با یک مقدار پیش‌فرض قابل استفاده جایگزین می‌شه. مقادیر اصلاح‌شده در حافظه باقی می‌مونن تا کاربر در صورت نیاز، همون بخش رو از طریق Inspector ذخیره کنه.

**Probeها قابل تعویض هستن.**

قرارداد اصلی Probe در اینترفیس `probe.Probe` تعریف شده و شامل متدهای `Init`، `Run`، `Schema` و `Close` هست. متد `Run` یک `netip.Addr` دریافت می‌کنه. به همین دلیل IPv4 و IPv6 از یک مسیر کد پردازش می‌شن.

**ساختار نتایج، اطلاعات لازم برای پردازش خودش رو نگه می‌داره.**

هر Probe یک `result.ResultSchema` داره که مسیر پوشه خروجی، ستون‌ها و Parser مربوط به نتایج رو مشخص می‌کنه. Writer هم مستقیماً از همین Schema استفاده می‌کنه. ثبت Schemaها در `core.Init()` باعث می‌شه مرورگر فایل‌ها و جدول نتایج بتونن هر فایل روی دیسک رو به Schema مربوط به خودش وصل کنن. برای توسعه نتایج هم نیازی به یک Struct مشترک برای همه Probeها نیست.

**موتور اسکن به جزئیات پروتکل‌ها وابسته نیست.**

موتور فقط آدرس‌ها و نتایج رو بین مراحل جابه‌جا می‌کنه و کاری به این نداره که هر Probe دقیقاً چه چیزی رو اندازه می‌گیره. نحوه اجرای مراحل اسکن از طریق تنظیمات انتخاب می‌شه و برای اضافه کردن یک حالت جدید، نیازی به بازنویسی کل موتور نیست.

**اسکنر از پروسه‌های خارجی استفاده نمی‌کنه.**

Probeهای ICMP، TCP، HTTP و ریزالور DNS فقط به کتابخانه استاندارد Go و `golang.org/x/net` نیاز دارن. Xray، DNSTT، VayDNS، MasterDNS، StormDNS و TheFeed هم از طریق کتابخانه‌های داخلی و در همان پروسه اجرا می‌شن.

Slipstream هم به همین شکل اجرا می‌شه: برنامه در زمان شروع، کتابخانه `libslipstream-client` رو با purego بارگذاری می‌کنه. اگر این کتابخانه موجود نباشه، Startup خطا رو گزارش می‌ده و اسکن Slipstream قابل اجرا نخواهد بود.

در Termux، اسکنر دستورهای سیستمی `termux-wake-lock` و `termux-wake-unlock` رو اجرا می‌کنه تا CPU هنگام اسکن در حالت بیدار بمونه. این‌ها تنها پروسه‌های فرزندی هستن که اسکنر اجرا می‌کنه.

**رابط کاربری از کامپوننت‌های مستقل ساخته شده.**

هر صفحه اینترفیس `ui.Component` رو پیاده‌سازی می‌کنه. Overlayها روی هم قرار می‌گیرن و در هر لحظه، بالاترین Overlay ورودی کاربر رو دریافت می‌کنه.

## صفحه‌های مرتبط

- [Core](./core/): جزئیات موتور اسکن، Probeها، تنظیمات و پردازش نتایج
- [رابط کاربری](./ui/): ساختار کامپوننت‌ها، Layout و تم‌ها
- [شروع کار](../getting-started/): راهنمای Build و اجرای پروژه

