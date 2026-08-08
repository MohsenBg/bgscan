
---

title: "هسته"
weight: 3
---

# هسته

تمام بخش‌هایی که مربوط به رابط کاربری متنی (TUI) نیستند در `internal/core` قرار دارند: تنظیمات، موتور اسکن، Probeها، فهرست‌های IP، نتیجه‌ها، DNS، Xray و مدیریت Processها.

## نقشه پکیج‌ها

| پکیج | مسئولیت |
| --- | --- |
| `config` | انواع `ScannerConfig`، مقادیر پیش‌فرض داخلی و `Store` برای خواندن/نوشتن `settings/*.toml` |
| `config/validate` | اعتبارسنجی و نرمال‌سازی هر پروتکل، تجمیع شده در `aggregate.go` |
| `scanner` | اینترفیس `Scanner`، ساخت Stageها و سرهم‌بندی Pipeline |
| `scanner/engine` | اجرا: اسکن تکی، زنجیره‌ای ترتیبی، Pipeline استریمینگ و Pipeline دسته‌ای (batch) |
| `scanner/probe` | اینترفیس `Probe` و یک زیرپکیج به ازای هر Probe |
| `scanner/portmgr` | استخر پورت‌های محلی موقت برای Probeهایی که کلاینت تانل اجرا می‌کنند |
| `result` | `Writer`، ادغام فایل‌های CSV، ثبت Schema، بارگذاری و شمارنده‌ها |
| `iplist` | ورود CSV لیست IPها، پارس کردن، ثبت، Shuffle و الاستریم |
| `dns` | توابع کمکی کوئری DNS، پارس Transport، کلاینت‌های DNSTT و SlipStream و SOCKS5 |
| `xray` | اجراکننده Xray، کانفیگ Inbound/Outbound، پارس لینک و تست سرعت |
| `speedtest` | سنجش Latency، دانلود و آپلود برای استفاده در Probe مربوط به Xray |
| `netutil` | نرمال‌سازی Host، پارس نسخه TLS و استخراج SNI برای Probeهای HTTP |
| `process` | اجرای Cross-platform و کشتن Processها |
| `fileutil` | توابع کمکی CSV، JSON، TOML، متن، فایل موقت و Path |

## Scanner

عبارت `scanner.Scanner` یک اینترفیس است، نه Struct. رابط کاربری آن را نگه می‌دارد، Stage به آن اضافه می‌کند و آن را اجرا می‌نماید.

```go
type Scanner interface {
    Run() error
    Close() error

    GetStages() []StageConfig
    AddStage(StageConfig)

    Pause()
    Resume()
    IsPaused() bool
    PausedDuration() time.Duration

    BuildICMPStage(context.Context) (StageConfig, error)
    BuildTCPStage(context.Context) (StageConfig, error)
    BuildHTTPStage(context.Context) (StageConfig, error)
    BuildXrayStage(context.Context, string) (StageConfig, error)
    BuildResolveStage(context.Context) (StageConfig, error)
    BuildDNSTTStage(context.Context) (StageConfig, error)
    BuildSlipStreamStage(context.Context) (StageConfig, error)
}
```

متد `BuildXrayStage` نام قالب Outbound انتخاب‌شده در UI را می‌گیرد. بقیه Builderها فقط یک `context` دریافت می‌کنند.

ساختار یک Stage:

```go
type StageConfig struct {
    Workers int
    Probe   probe.Probe
    Writer  result.Writer
    Rate    int
    Hooks   engine.ScanHooks
}
```

هر Builder تنظیمات پروتکل خود را می‌خواند، Probe را می‌سازد، یک `result.Writer` متصل به Schemaی همان Probe ایجاد می‌کند و Stage را برمی‌گرداند. سپس `AddHooks` Callbackهای UI را متصل می‌کند.

```text
NewScanner(ctx, input, opts...)
  ├─ AddStage(BuildICMPStage(ctx))
  ├─ AddStage(BuildTCPStage(ctx))
  └─ Run()
      ├─ تک Stage   → engine.RunScan
      └─ چند Stage  → engine.RunScanWithChain
```

مقدار concrete برای `scanner` گزینه‌هایی می‌پذیرد: `WithConfig` به جای خواندن از دیسک، کانفیگ را تزریق می‌کند و `WithPauseController` یک کنترل‌کننده Pause مشترک فراهم می‌کند. تست‌ها از `scanRunner` برای ساخت Pipeline بدون شروع Workerها استفاده می‌کنند.

## Engine

پکیج `scanner/engine` مسئول جابه‌جایی IPها و نتیجه‌هاست و اطلاعی از کارکرد داخلی Probeها ندارد.

#### اسکن تکی (Single Scan)

```text
iplist.StreamActiveIPs
   │
   ▼
 کانال ips ──► استخر ورکرها (N گوروتین)
                    ├─ محدودکننده نرخ (rate limiter)
                    ├─ probe.Run(ctx, addr)
                    ├─ موفقیت  → writer + OnSuccess
                    └─ خطا     → log + OnError
                    │
                    ▼
              result.Writer ──► ادغام CSV ──► دیسک

گوروتین پیشرفت → فراخوانی OnProgress در هر status_interval
```

تابع `RunScan` زمانی برمی‌گردد که Reader تمام شود، Workerها خالی شوند، Writer داده‌ها را Flush کند و آخرین گزارش پیشرفت ارسال شود.

#### اسکن زنجیره‌ای (Chain Scan)

تابع `RunScanWithChain` بر اساس `PipelineMode` عمل می‌کند:

| حالت (Mode) | تابع | ارتباط بین Stageها |
| --- | --- | --- |
| `sequential` | `executeSequentialChain` | مرحله N روی فایل می‌نویسد، مرحله N+1 از روی آن می‌خواند. |
| `streaming` | `executeStreamingPipeline` | کانال بافرشده `chan netip.Addr`؛ همه Stageها هم‌زمان اجرا می‌شوند. |
| `batch` | `executeBatchPipeline` | چانک‌های با اندازه ثابت به ترتیب از تمام Stageها عبور می‌کنند. |

تابع `createStageChannels` اندازه هر کانال را برابر `MaxBuffer` قرار می‌دهد (در صورت عدم تنظیم پیش‌فرض ۱۰,۰۰۰ است) و اگر تعداد Workerهای Stage بعدی بیشتر باشد، آن را افزایش می‌دهد.

تابع `calculateBatchSize` برای تک‌مرحله مقدار `BatchSize` را برمی‌گرداند، در غیر این صورت `max(BatchSize, بالاترین تعداد ورکر در مراحل بعدی)` را محاسبه کرده و در صورت عدم تنظیم پیش‌فرض ۱,۰۰۰ را اعمال می‌کند.

#### تایپ‌ها

```go
type ChainConfig struct {
    Mode             PipelineMode
    MaxBuffer        int
    BatchSize        int
    MaxIPsToTest     uint64
    Stages           []ScanConfig
    MinProbeDuration time.Duration
    Pause            PauseController
    Shuffled         bool
    RateLimiter      *rate.Limiter
}

type ScanConfig struct {
    Workers          int
    MaxIPsToTest     uint64
    MinProbeDuration time.Duration
    ProgressInterval time.Duration
    Probe            probe.Probe
    Writer           result.Writer
    Hooks            ScanHooks
    Pause            PauseController
    Shuffled         bool
    RateLimiter      *rate.Limiter
}

type ScanHooks struct {
    OnProgress func(Progress)
    OnSuccess  func(result.Result)
    OnScanEnd  func()
    OnError    func(error)
}
```

استفاده از Hookها اختیاری است. Engine از طریق توابع `callOnSuccess`، `callOnError` و `callOnScanEnd` عمل می‌کند که در صورت `nil` بودن بدون خطا عبور می‌کنند.

تابع `ParsePipelineMode` نام‌های مستعار را می‌پذیرد: `simple` برای sequential، `parallel` برای streaming و `pipeline` برای batch. مقادیر ناشناخته حالت `ModeSequential` را برمی‌گردانند.

#### محدودسازی نرخ (Rate limiting)

تنظیمات `min_probe_duration`، `probe_per_sec` و `probe_burst` در `GeneralConfig` به یک `rate.Limiter` واحد (Token Bucket) و یک تاخیر به ازای هر Probe تبدیل می‌شوند که به طور یکسان روی تمام Stageها در هر دو حالت اسکن تکی و زنجیره‌ای اعمال می‌گردد (این تنظیمات به ازای هر Stage نیستند). متد `Wait` محدودکننده، هر Worker را قبل از اجرای `probe.Run` متوقف می‌کند؛ اگر یک Probe سریع‌تر از `MinProbeDuration` تمام شود، Worker به اندازه باقیمانده زمان می‌خوابد. `probe_per_sec` نرخ شارژ Tokenها (سقف پایدار اسکن) و `probe_burst` ظرفیت Bucket (حداکثر اسکن پشت‌سرهم) است. تعداد `Workers` هم‌روندی را کنترل می‌کند نه نرخ اسکن را — محدودکننده مستقلاً حجم اسکن را کنترل می‌کند.

#### کنترل توقف (Pause control)

تایپ `PauseController` اینترفیسی است که توسط `NewPauseController()` پیاده‌سازی شده است. Workerها قبل از هر Probe به Checkpoint می‌رسند و در صورت متوقف بودن اسکن، آنجا منتظر می‌مانند. متد `PausedDuration()` زمان توقف را جمع می‌زند تا نرخ پیشرفت دقیق بماند.

## اینترفیس Probe

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip netip.Addr) (result.Result, error)
    Schema() result.ResultSchema
    Close() error
}
```

متد `Init` یک‌بار قبل از اجرای هر `Run` و متد `Close` در پایان اجرا می‌شود. متد `Run` یک `netip.Addr` می‌گیرد (نه String) که باعث می‌شود IPv4 و IPv6 از یک مسیر کد یکسان استفاده کنند. متد `Schema` ارتباط Probe با چیدمان نتیجه و دایرکتوری خروجی را مشخص می‌کند.

هر Probe زیرپکیج اختصاصی خود را در `scanner/probe` دارد:

| پکیج | Probe | سازنده (Constructor) |
| --- | --- | --- |
| `icmpprobe` | ICMP echo برای IPv4 و IPv6 | `NewICMPProbe(Options)` |
| `tcpprobe` | اتصال TCP | `NewTCPProbe(port, timeout, tries)` |
| `httpprobe` | HTTP/1.1 و HTTP/2 روی ALPN | `NewHTTPProbe(req, acceptedCodes)` |
| `httpprobe` | HTTP/3 روی QUIC | `NewHTTP3Probe(req, acceptedCodes)` |
| `resolveprobe` | ریزالور DNS همراه با بررسی DPI | `NewResolverProbe(*DNSRequest)` |
| `dnsttprobe` | اعتبارسنجی تانل DNSTT | `NewDNSTTProbe(config, portMgr)` |
| `slipstreamprobe` | اعتبارسنجی تانل SlipStream | `NewSlipstreamProbe(workers, config, portMgr)` |
| `xrayprobe` | اتصال و پهنای باند Xray | `NewXrayProbe(cfg, template, portMgr)` |

Probeهایی که فایل اجرایی کلاینت را اجرا می‌کنند یک `portmgr.Manager` می‌گیرند و به ازای هر Probe یک پورت محلی اجاره می‌کنند تا Workerهای هم‌روند با هم تداخل پیدا نکنند.

## سیستم نتیجه‌ها (Result System)

هر Probe یک ساختار داده متناسب با اینترفیس `result.Result` و یک `result.ResultSchema` برای توصیف آن تعریف می‌کند.

```go
type Result interface {
    Key() string
    KeyType() KeyType
    ToRecord() []string
    Equal(other Result) bool
    Score() float64
}

type ResultSchema struct {
    Name      string
    Directory string
    Columns   []ColumnDef
    Parser    ResultParser
}
```

ترتیب خروجی `ToRecord` باید با ترتیب `Columns` یکسان باشد و `Parser` جهت بارگذاری معکوس فایل استفاده می‌شود. متد `Key` برای یکسان‌سازی و حذف داده‌های تکراری، `KeyType` برای تفکیک نتایج بر اساس IP یا Domain و `Score` برای مرتب‌سازی نتایج کاربرد دارد.

#### ثبت (Registration)

تابع `core.Init()` در `internal/core/core.go` قبل از اجرای هر بخش دیگری، تمام Schemaهای داخلی را در `result.DefaultRegistry` ثبت می‌کند:

```go
result.DefaultRegistry.Register(icmpprobe.Schema)
result.DefaultRegistry.Register(tcpprobe.Schema)
// http, resolve, dnstt, slipstream, xray
```

این Registry نام دایرکتوری را به Schema نگاشت می‌دهد؛ بدین ترتیب مرورگر فایل نتایج متوجه می‌شود که چگونه یک فایل روی دیسک را پارس کرده و نمایش دهد.

برای افزودن یک نوع اسکن جدید کافیست Struct نتیجه، Schema، Probe و Stage Builder را بنویسید و Schema را در `core.Init()` ثبت کنید. نیازی به تغییر در Writer، Registry یا جدول نتایج وجود ندارد.

#### Writer

```go
type Writer interface {
    Start() error
    Stop() error
    Write(r Result)
    GetResultPath() string
}

type WriterOptions struct {
    ResultPrefix string
    Schema       ResultSchema
    Config       config.WriterConfig
}
```

تابع `NewWriter` تنظیمات Writer و Schema را اعتبارسنجی کرده، سپس مسیر خروجی را از روی `result_directory`، دایرکتوری Schema، Prefix و برچسب زمانی محاسبه می‌کند. متد `Start` دایرکتوری را ساخته و فایل‌های قدیمی را پاک می‌سازد. متد `Write` داده‌ها را در کانالی با اندازه `chan_size` صف‌بندی می‌کند و در صورت لغو Context، نتیجه را نادیده می‌گیرد.

عملیات Flush زمانی رخ می‌دهد که تعداد به `batch_size` برسد، زمان `merge_flush_interval` سر برسد، یا هنگام خروج برنامه. در صورت لغو اجرا، Writer ابتدا کانال را خالی می‌کند تا داده‌هایی که قبل از `Stop` دریافت شده‌اند روی دیسک بنشینند.

#### ادغام (Merge)

فایل `merger.go` دسته داده‌ها را بر اساس Score مرتب کرده، با فایل موجود به صورت Merge-Sort ترکیب می‌کند، در `<path>.tmp` می‌نویسد، `Sync` را فراخوانی کرده و روی فایل اصلی Rename می‌کند. داده‌های تکراری بر اساس کلید جایگزین می‌شوند و هیچ‌کدام از فایل‌ها به طور کامل وارد حافظه RAM نمی‌شوند.

#### Registry و Loader

- فایل `registry.go` شامل `DefaultRegistry` است که نام دایرکتوری را به Schema نگاشت می‌دهد. توابع `FindResultFiles` و `GetResultFiles` دایرکتوری نتایج را پیمایش کرده و Schemaی مربوطه را به هر فایل متصل می‌کنند.
- فایل `loader.go` نتایج را از طریق `LoadResult` به صورت استریم یا با `LoadAll` به صورت بخش‌های محدود برمی‌گرداند که هر دو از Parser موجود در Schema استفاده می‌کنند.
- فایل `count.go` رکوردها را بدون بارگذاری کامل فایل شمارش می‌کند.

## تنظیمات (Config)

هیچ Singleton یا دسترسی‌دهنده سراسری (Package-level) برای تنظیمات وجود ندارد. متد `main` یک `Store` می‌سازد، مقدار `ScannerConfig` را بارگذاری کرده و اشاره‌گر آن را به UI و Scanner پاس می‌دهد.

```go
type ScannerConfig struct {
    General GeneralConfig
    Writer  WriterConfig
    ICMP    ICMPConfig
    TCP     TCPConfig
    HTTP    HTTPConfig
    Xray    XrayConfig
    DNS     DNSConfig
}

store := config.NewStore()          // به صورت پیش‌فرض دایرکتوری "settings"
cfg, err := store.Load()
```

استفاده از `WithSettingsDir(dir)` مسیر `Store` را تغییر می‌دهد که برای تست‌ها در دایرکتوری موقت کاربرد دارد.

متد `Load` به ازای هر بخش یک فایل TOML می‌خواند. اگر فایلی وجود نداشته باشد، از روی مقادیر پیش‌فرض ساخته می‌شود. اگر فایلی وجود داشته باشد اما پارس نشود، به جای بازگشت خاموش به پیش‌فرض، خطا برمی‌گرداند.

ذخیره‌سازی به صورت بخش‌بندی شده انجام می‌شود: `SaveGeneral`، `SaveWriter`، `SaveICMP`، `SaveTCP`، `SaveHTTP`، `SaveXray` و `SaveDNS`. Inspector ساختار داخل حافظه را ادیت کرده و متد مربوطه را فراخوانی می‌کند.

رابط کاربری هر دو را در `ui.AppState` نگه می‌دارد:

```go
type AppState struct {
    Layout *layout.Layout
    Config *config.ScannerConfig
    Store  *config.Store
}
```

#### اعتبارسنجی (Validation)

پکیج `config/validate` به ازای هر بخش یک فایل دارد. توابع `ValidateXxx` نگاشتی از خطاهای فیلدها را برمی‌گردانند و چیزی را تغییر نمی‌دهند. توابع `NormalizeXxx` فیلدهای نامعتبر را به مقادیر پیش‌فرض تغییر داده و برای هر تصحیح یک `Warning` برمی‌گردانند. فایل `aggregate.go` این‌ها را در `ValidateAll` و `NormalizeAll` ترکیب می‌کند.

در زمان Startup، تابع `NormalizeAll` فراخوانی شده، تمام اصلاحات چاپ می‌شوند و سپس بخش‌های اصلاح‌شده روی دیسک نوشته می‌شوند تا TOML روی دیسک همواره با آنچه Scanner استفاده می‌کند مطابقت داشته باشد.

## فهرست‌های IP (IP Lists)

- فایل `parser.go` تابع `StreamActiveIPs(ctx, path, limit, shuffled, out)` (تابعی که Engine از آن برای تغذیه Workerها استفاده می‌کند) و `StreamCIDR` را برای باز کردن Prefix ارائه می‌دهد.
- فایل‌های `csv.go` و `parser.go` فرمت دو ستونی `<ip_or_cidr>,<enabled>` را می‌خوانند.
- فایل `loader.go` وارد کردن لیست‌های خارجی به همراه `CountIPs` و `CountActiveIPs` را برای مجموع پیشرفت مدیریت می‌کند.
- فایل `registry.go` فایل‌های زیر `ips/` را لیست می‌کند و `shuffle.go` ترتیب آن‌ها را تصادفی می‌سازد.

ورودی‌های غیرفعال (`enabled=false`) هنگام استریم نادیده گرفته می‌شوند. هر دو Prefix مربوط به IPv4 و IPv6 پشتیبانی می‌شوند؛ بنابراین شمارش یک Prefix بزرگ IPv6 به جای یک مقدار دقیق، یک مقدار اشباع‌شده برمی‌گرداند.

## زیرسیستم DNS

- فایل `query.go` کوئری‌ها را ساخته و روی UDP، TCP و DoT ارسال می‌کند.
- فایل `type.go` انواع Transport و Rcodeها را پارس می‌کند. حالت DoH با موفقیت پارس می‌شود اما به DoT تبدیل می‌گردد زیرا اسکنر ریزالورها را با IP هدف قرار می‌دهد.
- فایل‌های `dnstt.go` و `slipstream.go` برنامه‌های اجرایی کلاینت خارجی را Wrapper کرده و انتخاب Transport را به Flagهای کلاینت تبدیل می‌کنند.
- فایل `socks5.go` یک کلاینت ساده SOCKS5 است که برای اعتبارسنجی تانل پس از بالا آمدن استفاده می‌شود.

## یکپارچه‌سازی Xray

- فایل‌های `xray.go` و `command.go` پروسس را اجرا و کنترل می‌کنند.
- فایل‌های `inbound.go` و `outbound.go` JSON کانفیگ را تولید می‌کنند.
- فایل `link.go` لینک‌های اشتراکی را پارس می‌کند.
- فایل `speedtest.go` میزان سرعت (Throughput) را اندازه‌گیری می‌کند.

پکیج `xrayprobe` کانفیگ Outbound انتخابی را می‌سازد، یک پورت اجاره می‌کند، Xray را اجرا کرده، Latency را از طریق پروکسی محلی می‌سنجد، در صورت تنظیم تست سرعت را اجرا کرده و در نهایت همه چیز را پاکسازی می‌کند.

## مدیریت پروسس‌ها (Process Management)

فایل `process.go` اینترفیس را تعریف می‌کند و فایل‌های `process_unix.go` و `process_windows.go` رفتار اختصاصی هر سیستم‌عامل را پیاده‌سازی می‌کنند. تمام Probeهایی که یک فایل اجرایی اجرا می‌کنند از این بخش استفاده می‌نمایند.

## لاگر (Logger)

| لاگر | فایل | پوشش |
| --- | --- | --- |
| Core | `logs/core.log` | Engine، Probeها، Config و I/O نتایج |
| UI | `logs/ui.log` | چرخه حیات کامپوننت‌ها، عملیات فایل و خطاهای UI |
| Debug | `logs/debug.log` | State Dumpها و Traceهای دقیق |

هر لاگر در فایل می‌نویسد و داده‌ها را برای هر کانال نمایشگر فعال ارسال می‌کند. سیستم چرخش (Rotation) از نوع lumberjack است: سقف ۵۰ مگابایت، ۳ فایل پشتیبان، نگهداری ۷ روز و به صورت فشرده (compressed).

## راه‌اندازی (Startup)

تابع `startup.RunHealthChecks(&cfg, &store)` بعد از بارگذاری کانفیگ و قبل از شروع TUI اجرا می‌شود:

```text
1. checkLoggerHealth()      باز کردن فایل‌های لاگ و شروع سیستم چرخش
2. theme.Init()             اعمال پالت رنگی
3. checkConfigHealth()      اجرای NormalizeAll و ذخیره بخش‌های اصلاح‌شده
4. checkXrayHealth()        پیدا کردن فایل اجرایی Xray و قالب‌ها
5. checkDNSTTHealth()       پیدا کردن dnstt-client
6. checkSlipstreamHealth()  پیدا کردن slipstream-client
```

بررسی‌ها بین فایل‌های `checks_logger.go`، `checks_config.go`، `checks_xray.go` و `checks_dns.go` تقسیم شده‌اند. عدم وجود یک فایل اجرایی اختیاری، فقط یک هشدار چاپ کرده و همان نوع اسکن را غیرفعال می‌کند. این مرحله با پیام فشردن کلید Enter پایان می‌یابد.

توجه داشته باشید که خطاهای کانفیگ زودتر رخ می‌دهند: فراخوانی `store.Load()` در `main` در صورت نامعتبر بودن TOML بلافاصله اجرای برنامه را متوقف می‌کند (قبل از اینکه بررسی‌های سلامت اجرا شوند).

می‌توانید ثابت `fastboot` را در فایل `health.go` برابر `true` قرار دهید تا مکث‌های ۵۰۰ میلی‌ثانیه‌ای بین بررسی‌ها هنگام دیباگ حذف شوند.

## صفحات مرتبط

- [معماری (Architecture)](../architecture/) — چیدمان پروژه و لایه‌بندی
- [رابط کاربری (UI)](../ui/) — مدل کامپوننت‌ها و پوسته (Theme)
