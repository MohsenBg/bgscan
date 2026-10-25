---
title: "هسته"
weight: 3
---

# Core

پوشه `internal/core` بخش‌های اصلی BGScan به‌جز رابط کاربری TUI رو در خودش نگه می‌داره، از تنظیمات و موتور اسکن گرفته تا Probeها، فهرست‌های IP، نتایج، DNS، Xray و ابزار قدیمی مدیریت پروسه‌ها.

## نقشه پکیج‌ها

| پکیج | مسئولیت |
|---|---|
| `config` | تعریف `ScannerConfig`، پیش‌فرض‌های کامپایل‌شده و `Store` برای خواندن و نوشتن فایل‌های `settings/*.toml`. |
| `config/validate` | اعتبارسنجی و نرمال‌سازی تنظیمات هر پروتکل، این بخش‌ها در `aggregate.go` با هم ترکیب می‌شن. |
| `scanner` | اینترفیس `Scanner`، سازنده‌های Stage و سرهم‌بندی مراحل اسکن. |
| `scanner/engine` | اجرای اسکن به‌صورت Sequential، Streaming یا Batch. حالت Sequential برای اسکن تک‌مرحله‌ای هم استفاده می‌شه. |
| `scanner/probe` | اینترفیس `Probe` و زیربسته‌های مربوط به هر Probe. |
| `scanner/portmgr` | مدیریت Pool پورت‌های محلی موقت برای Probeهای تونل و Xray. |
| `result` | `Writer`، ادغام فایل‌های CSV، Registry مربوط به Schemaها، Loader و شمارنده‌های نتایج. |
| `iplist` | خواندن فهرست IPها از CSV، Parse، Registry، Shuffle و پردازش جریانی ورودی‌ها. |
| `dns` | ابزارهای کوئری DNS، Parse ترنسپورت، مدیریت تنظیمات تونل‌های DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS و TheFeed، و مدیریت خود تونل‌ها. |
| `socks` | کلاینت SOCKS5 برای اعتبارسنجی تونل. |
| `ssh` | کلاینت SSH برای اتصال‌های تونل‌شده. |
| `xray` | سرویس Xray داخل پروسه، تنظیمات ورودی و خروجی، Parse لینک‌ها و تست سرعت. |
| `speedtest` | اندازه‌گیری تأخیر، سرعت دانلود و آپلود، مورد استفاده Probe مربوط به Xray. |
| `netutil` | نرمال‌سازی Host، Parse نسخه TLS و استخراج SNI برای Probeهای HTTP. |
| `process` | ابزار قدیمی ردیابی پروسه که اسکنر و TUI فعلی ازش استفاده نمی‌کنن. |
| `ffi` | اتصال با purego به کتابخانه کلاینت Slipstream. |
| `fileutil` | ابزارهای کمکی برای CSV، JSON، TOML، متن، فایل‌های موقت و مسیرها. |

## Scanner

`scanner.Scanner` یک اینترفیسه، نه یک Struct. رابط کاربری نمونه‌ای از اون رو نگه می‌داره، Stageها رو بهش اضافه می‌کنه و اسکن رو اجرا می‌کنه.

```go
type Scanner interface {
    Run() error
    Close() error

    GetStages() []StageConfig
    AddStage(StageConfig) error
    UpdateStageHooks(index int, hooks engine.ScanHooks) error

    Pause()
    Resume()
    IsPaused() bool
    PausedDuration() time.Duration

    BuildICMPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildTCPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildHTTPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildXrayStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildResolveStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
    BuildDNSTTStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildSlipStreamStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildVayDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildMasterDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildStormDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
    BuildTheFeedStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
}
```

متدهای `BuildXrayStage`، `BuildDNSTTStage`، `BuildSlipStreamStage`، `BuildVayDNSStage`، `BuildMasterDNSStage`، `BuildStormDNSStage` و `BuildTheFeedStage` نام فایل Config رو می‌گیرن و یک Slice از Stageها برمی‌گردونن. این Slice می‌تونه شامل یک Stage پیش‌اسکن اختیاری و Stage اصلی باشه.

Xray از گزینه `pre_scan_type` استفاده می‌کنه که می‌تونه `tcp`، `icmp`، `http` یا `none` باشه. سازنده‌های تونل DNS هم وقتی `check_dns_resolver` فعال باشه، یک مرحله برای بررسی ریزالور اضافه می‌کنن. بقیه سازنده‌ها فقط یک Stage برمی‌گردونن.

همه سازنده‌ها می‌تونن `ScanHooks` رو به‌صورت آرگومان‌های Variadic دریافت کنن. بنابراین، لازم نیست برای استفاده از Hookها امضای جداگانه‌ای داشته باشن.

ساختار یک Stage به این شکله:

```go
type StageConfig struct {
    Workers          int
    ProgressInterval time.Duration
    Probe            probe.Probe
    Writer           result.Writer
    Hooks            engine.ScanHooks
}
```

هر سازنده، تنظیمات پروتکل مربوط به خودش رو می‌خونه، Probe رو می‌سازه، یک `result.Writer` متناسب با Schema همون Probe ایجاد می‌کنه و در نهایت Stage رو برمی‌گردونه. متد `AddHooks` هم برای وصل کردن Callbackهای رابط کاربری استفاده می‌شه.

جریان کلی استفاده از Scanner به این شکله:

```text
NewScanner(ctx, input, opts...)
 ├─ AddStage(BuildICMPStage(ctx))
 ├─ AddStage(BuildTCPStage(ctx))
 └─ Run()
     └─ engine.RunScanWithChain
          sequential | streaming | batch
```

پیاده‌سازی واقعی `scanner` چند Option می‌گیره. گزینه `WithConfig` اجازه می‌ده Config از بیرون به Scanner داده بشه، به‌جای اینکه خودش از دیسک بخونه. `WithPauseController` هم یک کنترلر Pause مشترک در اختیارش می‌ذاره.

تست‌ها از نقطه اتصال `scanRunner` استفاده می‌کنن تا مراحل اسکن رو بدون راه‌اندازی Workerها سرهم‌بندی کنن.

## Engine

پکیج `scanner/engine` مسئول اجرای مراحل اسکنه. این بخش فقط با IPها و نتایج سروکار داره و از جزئیات کاری که هر Probe انجام می‌ده خبر نداره.

### اسکن تک‌مرحله‌ای

اسکن تک‌مرحله‌ای هم از همون مسیر اجرای اسکن استفاده می‌کنه. فقط حالت Sequential اجرا می‌شه. برای اسکن تک‌مرحله‌ای، مسیر جداگانه‌ای در Engine وجود نداره.

```text
iplist.StreamActiveIPs
       │
       ▼
ips channel ──► worker pool (N goroutines)
                   ├─ rate limiter
                   ├─ probe.Run(ctx, addr)
                   ├─ success → writer + OnSuccess
                   └─ error   → log + OnError
                   │
                   ▼
             result.Writer ──► CSV merge ──► disk

progress goroutine → OnProgress every status_interval
```

اجرای اسکن وقتی تموم می‌شه که Reader به انتهای ورودی رسیده باشه، همه Workerها کارشون رو تموم کرده باشن، Writer داده‌ها رو Flush کرده باشه و گزارش نهایی پیشرفت هم ارسال شده باشه.

### اسکن زنجیره‌ای

تابع `RunScanWithChain` با توجه به مقدار `PipelineMode`، نحوه اجرای Stageها رو انتخاب می‌کنه.

| حالت | تابع | نحوه ارتباط Stageها |
|---|---|---|
| `sequential` | `runSequential` | هر Stage نتایجش رو در فایل می‌نویسه و Stage بعدی اون فایل رو می‌خونه. |
| `streaming` | `runStreaming` | همه Stageها هم‌زمان اجرا می‌شن و IPها از طریق یک `chan netip.Addr` بافرشده بینشون منتقل می‌شن. |
| `batch` | `runBatch` | داده‌ها به Batchهای با اندازه مشخص تقسیم می‌شن و هر Batch به‌ترتیب از Stageها عبور می‌کنه. |

تابع `createStageChannels` اندازه بافر هر کانال رو از `MaxBuffer` می‌گیره. اگه این مقدار تنظیم نشده باشه، پیش‌فرض `10,000` هست. اگه تعداد Workerهای Stage بعدی بیشتر باشه، اندازه بافر متناسب با تعداد اون‌ها افزایش پیدا می‌کنه.

تابع `pipeline.batchSize` برای اسکن تک‌مرحله‌ای مقدار `BatchSize` رو برمی‌گردونه. در اسکن چندمرحله‌ای، مقدار بزرگ‌تر بین `BatchSize` و بیشترین تعداد Worker در Stageهای بعد از Stage اول انتخاب می‌شه. اگه `BatchSize` تنظیم نشده باشه، مقدار پیش‌فرض `1,000` هست.

### ساختارها و نوع‌ها

```go
type ChainConfig struct {
    Mode             PipelineMode
    MaxBuffer        int
    BatchSize        int
    MaxIPsToTest     uint64
    Stages           []StageConfig
    MinProbeDuration time.Duration
    Pause            PauseController
    Shuffled         bool
    RateLimiter      *rate.Limiter
    MaxSuccessfulIPs uint64
    Log              *logger.Logger
}

type StageConfig struct {
    Workers          int
    ProgressInterval time.Duration
    Probe            probe.Probe
    Writer           result.Writer
    Hooks            ScanHooks
}

type ScanHooks struct {
    OnProgress func(Progress)
    OnSuccess  func(result.Result)
    OnScanEnd  func()
    OnError    func(error)
}
```

استفاده از Hookها اختیاریه. Engine برای اجرای Callbackها از توابع `callOnSuccess`، `callOnError` و `callOnScanEnd` استفاده می‌کنه. اگه Callback مربوطه `nil` باشه، این توابع کاری انجام نمی‌دن.

تابع `ParsePipelineMode` فقط مقدارهای `sequential`، `streaming` و `batch` رو می‌شناسه. مقدارهایی مثل `simple`، `parallel` و `pipeline` به‌عنوان حالت ناشناخته در نظر گرفته می‌شن و نتیجه `ModeSequential` خواهد بود.

اعتبارسنج تنظیمات هنوز سه اسم قدیمی رو قبول می‌کنه، اما Engine اون‌ها رو به‌عنوان حالت‌های شناخته‌شده نمی‌شناسه. در نتیجه، این مقادیر از اعتبارسنجی عبور می‌کنن، ولی اجرا در نهایت به حالت Sequential برمی‌گرده.

### محدودیت نرخ

گزینه‌های `min_probe_duration`، `probe_per_sec` و `probe_burst` از `GeneralConfig` به یک `rate.Limiter` از نوع Token Bucket و یک تأخیر برای هر Probe تبدیل می‌شن. این محدودیت‌ها در اسکن تک‌مرحله‌ای و زنجیره‌ای، روی همه Stageها اعمال می‌شن و تنظیم جداگانه‌ای برای هر Stage ندارن.

- `probe_per_sec` نرخ پر شدن سطل توکن رو تعیین می‌کنه و روی نرخ پایدار اجرای Probeها اثر می‌ذاره.
- `probe_burst` ظرفیت سطل رو مشخص می‌کنه، یعنی چند Probe می‌تونن پشت‌سرهم اجرا بشن.
- `Workers` تعداد کارهای هم‌زمان رو کنترل می‌کنه، نه نرخ اجرای Probeها.
- Limiter نرخ رو مستقل از تعداد Workerها محدود می‌کنه.

هر Worker قبل از اجرای `probe.Run` منتظر `Wait` مربوط به Limiter می‌مونه. اگه Probe زودتر از مقدار `MinProbeDuration` تموم بشه، Worker باقی زمان رو منتظر می‌مونه.

### کنترل Pause

اینترفیس `PauseController` رفتار توقف و ادامه اسکن رو مشخص می‌کنه و تابع `NewPauseController()` پیاده‌سازی اون رو می‌سازه.

Workerها قبل از اجرای هر Probe به یک Checkpoint می‌رسن. اگه اسکن Pause شده باشه، همون‌جا منتظر می‌مونن تا اجرا دوباره ادامه پیدا کنه.

متد `PausedDuration()` مدت زمانی رو که اسکن در حالت Pause بوده محاسبه می‌کنه تا زمان توقف روی محاسبه نرخ پیشرفت واقعی تأثیر نذاره.

## اینترفیس Probe

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip netip.Addr) (result.Result, error)
    Schema() result.ResultSchema
    Close() error
}
```

متد `Init` یک بار، پیش از اجرای `Run`ها، فراخوانی می‌شه و `Close` هم در پایان کار اجرا می‌شه. متد `Run` به‌جای String یک `netip.Addr` دریافت می‌کنه. به همین دلیل IPv4 و IPv6 از یک مسیر کد استفاده می‌کنن.

متد `Schema` هم مشخص می‌کنه نتایج Probe با چه ساختاری ذخیره بشن و در کدوم پوشه قرار بگیرن.

هر Probe در زیربسته خودش زیر `scanner/probe` پیاده‌سازی شده:

| پکیج | نوع Probe | سازنده |
|---|---|---|
| `icmpprobe` | بررسی پاسخ ICMP در IPv4 و IPv6 | `NewICMPProbe(Options)` |
| `tcpprobe` | بررسی اتصال TCP | `NewTCPProbe(port, timeout, tries)` |
| `httpprobe` | بررسی HTTP/1.1 و HTTP/2 از طریق ALPN | `NewHTTPProbe(req, acceptedCodes)` |
| `httpprobe` | بررسی HTTP/3 روی QUIC | `NewHTTP3Probe(req, acceptedCodes)` |
| `resolveprobe` | بررسی ریزالور DNS با تشخیص DPI | `NewResolverProbe(*DNSRequest)` |
| `dnsttprobe` | اعتبارسنجی تونل DNSTT | `NewDNSTTProbe(config, timeout, ...)` |
| `vaydnsprobe` | اعتبارسنجی تونل VayDNS | `NewVayDNSProbe(config, timeout, ...)` |
| `slipstreamprobe` | اعتبارسنجی تونل Slipstream | `NewSlipstreamProbe(config, timeout, portMgr, ...)` |
| `masterdnsprobe` | اعتبارسنجی تونل MasterDNS | `NewMasterDNSProbe(config, timeout, portMgr, ...)` |
| `stormdnsprobe` | اعتبارسنجی تونل StormDNS | `NewStormDNSProbe(config, timeout, portMgr, ...)` |
| `thefeedprobe` | اعتبارسنجی تونل TheFeed | `NewTheFeedProbe(config, timeout, ...)` |
| `xrayprobe` | بررسی اتصال و پهنای باند Xray | `NewXrayProbe(cfg, template, portMgr)` |

فقط Probe مربوط به Slipstream یک کتابخانه Native داخلی به نام `libslipstream-client` بارگذاری می‌کنه.

Probeهای Slipstream، MasterDNS، StormDNS و Xray یک `portmgr.Manager` دریافت می‌کنن و برای هر Probe یک پورت محلی اجاره می‌کنن تا Workerهای هم‌زمان از یک پورت مشترک استفاده نکنن.

## سیستم نتیجه

هر Probe نوع نتیجه مخصوص خودش رو داره که اینترفیس `result.Result` رو پیاده‌سازی می‌کنه. علاوه بر اون، هر Probe یک `result.ResultSchema` داره که ساختار نتیجه رو مشخص می‌کنه.

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

ترتیب مقادیر خروجی `ToRecord` باید با ترتیب `Columns` یکسان باشه. `Parser` هم برای تبدیل رکوردهای ذخیره‌شده به نتیجه، هنگام خواندن دوباره فایل استفاده می‌شه.

- `Key` مشخص می‌کنه نتایج چطور برای تشخیص تکراری‌ها مقایسه بشن.
- `KeyType` نتایجی رو که کلیدشون IP هست از نتایجی که کلیدشون دامنه است جدا می‌کنه.
- `Score` برای مرتب‌سازی نتایج استفاده می‌شه.

### ثبت Schemaها

تابع `core.Init()` در فایل `internal/core/core.go`، نُه Schema داخلی رو در `result.DefaultRegistry` ثبت می‌کنه:

```go
result.DefaultRegistry.Register(icmpprobe.Schema)
result.DefaultRegistry.Register(tcpprobe.Schema)
result.DefaultRegistry.Register(httpprobe.Schema)
result.DefaultRegistry.Register(resolveprobe.Schema)
result.DefaultRegistry.Register(dnsttprobe.Schema)
result.DefaultRegistry.Register(vaydnsprobe.Schema)
result.DefaultRegistry.Register(slipstreamprobe.Schema)
result.DefaultRegistry.Register(thefeedprobe.Schema)
result.DefaultRegistry.Register(xrayprobe.Schema)
```

Schemaهای `masterdnsprobe` و `stormdnsprobe` هم تعریف شدن، اما در این فهرست ثبت نمی‌شن. در نتیجه، Registry نمی‌تونه پوشه‌های این دو پروتکل رو به Schema مربوط به خودشون وصل کنه و فایل‌هاشون در مرورگر Result Files نمایش داده نمی‌شن.

با این حال، نتایج این دو پروتکل همچنان به‌صورت CSV در مسیرهای `result/masterdns/` و `result/stormdns/` ذخیره می‌شن.

Registry نام پوشه رو به Schema مربوط به اون وصل می‌کنه. به این ترتیب، مرورگر فایل‌های نتیجه می‌تونه تشخیص بده هر فایل روی دیسک رو چطور بخونه و نمایش بده.

برای اضافه کردن یک نوع اسکن جدید، باید نوع نتیجه، Schema، Probe و سازنده Stage مربوط به اون رو پیاده‌سازی کنید و Schema رو در `core.Init()` ثبت کنید. برای این کار نیازی به تغییر Writer، Registry یا جدول نتایج نیست.

### Writer

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
    Log          *logger.Logger
}
```

تابع `NewWriter` تنظیمات Writer و Schema رو اعتبارسنجی می‌کنه. بعد مسیر خروجی رو با استفاده از `result_directory`، پوشه Schema، پیشوند فایل و یک مهر زمانی می‌سازه.

متد `Start` پوشه خروجی رو ایجاد می‌کنه و فایل‌های قدیمی رو پاک می‌کنه. متد `Write` نتیجه رو داخل کانالی با ظرفیت `chan_size` قرار می‌ده. اگه Context از قبل Cancel شده باشه، نتیجه کنار گذاشته می‌شه.

Writer در سه حالت داده‌ها رو روی دیسک می‌نویسه:

- وقتی تعداد رکوردهای Batch به `batch_size` برسه.
- وقتی زمان `merge_flush_interval` فرا برسه.
- هنگام خاموش شدن Writer.

هنگام Cancel شدن، Writer ابتدا داده‌های باقی‌مونده در کانال رو پردازش می‌کنه. به این ترتیب، داده‌هایی که پیش از فراخوانی `Stop` پذیرفته شده باشن، فرصت ذخیره شدن روی دیسک رو دارن.

### ادغام نتایج

فایل `merger.go` مسئول مرتب‌سازی و ادغام نتایجه. این بخش ابتدا Batch رو بر اساس امتیاز مرتب می‌کنه، بعد اون رو با فایل موجود ادغام می‌کنه.

نتیجه در فایل موقت `\<path>.tmp` نوشته می‌شه، سپس `Sync` اجرا می‌شه و در پایان فایل موقت با فایل اصلی جایگزین می‌شه.

رکوردهای تکراری بر اساس کلید با هم مقایسه می‌شن و فقط رکورد مناسب نگه داشته می‌شه. برای ادغام هم نیازی نیست محتوای کامل هر دو فایل به‌طور هم‌زمان در حافظه بارگذاری بشه.

### Registry و Loader

- `registry.go`، مقدار `DefaultRegistry` رو نگه می‌داره که نام پوشه رو به Schema مربوط به اون وصل می‌کنه. توابع `FindResultFiles` و `GetResultFiles` پوشه نتایج رو می‌گردن و Schema مناسب رو به هر فایل نسبت می‌دن.
- `loader.go`، با `LoadResult` نتایج رو به‌صورت جریانی می‌خونه و با `LoadAll` یک Slice محدود از نتایج می‌سازه. هر دو از Parser مربوط به Schema استفاده می‌کنن.
- `count.go`، تعداد رکوردها رو بدون بارگذاری کامل نتایج در حافظه می‌شماره.

## تنظیمات

برای Config از Singleton یا Accessor سراسری در سطح پکیج استفاده نمی‌شه. بررسی‌های Startup یک `Store` می‌سازن، تنظیمات `ScannerConfig` رو می‌خونن و اون‌ها رو در اختیار Session رابط کاربری قرار می‌دن تا رابط و اسکنر از تنظیمات یکسانی استفاده کنن.

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

store := config.NewStore() // defaults to the "settings" directory
cfg, err := store.Load()
```

با `WithSettingsDir(dir)` می‌شه مسیر تنظیمات یک Store رو تغییر داد. تست‌ها از این قابلیت استفاده می‌کنن تا تنظیمات رو در یک پوشه موقت بخونن و بنویسن، نه در مسیر تنظیمات واقعی سیستم.

متد `Load` فایل TOML هر بخش رو می‌خونه. اگه فایل وجود نداشته باشه، با پیش‌فرض‌های کامپایل‌شده ساخته می‌شه. اما اگه فایل وجود داشته باشه و Parse نشه، تابع خطا برمی‌گردونه. در این حالت تنظیمات به‌صورت بی‌صدا با پیش‌فرض‌ها جایگزین نمی‌شن.

ذخیره تنظیمات هم برای هر بخش متد جداگانه داره:

`SaveGeneral`، `SaveWriter`، `SaveICMP`، `SaveTCP`، `SaveHTTP`، `SaveXray` و `SaveDNS`.

Inspector هر بخش، Struct تنظیمات رو در حافظه تغییر می‌ده و هنگام ذخیره، متد مربوط به همون بخش رو صدا می‌زنه.

رابط کاربری تنظیمات و Store رو از طریق `ui.Session` در اختیار بخش‌های مختلف قرار می‌ده:

```go
type Session struct {
    Config  *config.ScannerConfig
    Store   *config.Store
    Program ui.Program
    // env and core logger
}
```

متد `Session.Env()` هم اطلاعات محیط رابط کاربری، از جمله Layout، تم، کلیدها، لاگر رابط و `UISettings` رو برمی‌گردونه.

### اعتبارسنجی تنظیمات

پکیج `config/validate` برای هر بخش تنظیمات فایل جداگانه‌ای داره.

- توابع `ValidateXxx` یک Map از خطاهای مربوط به فیلدها برمی‌گردونن، بدون اینکه مقدارها رو تغییر بدن.
- توابع `NormalizeXxx` مقدارهای نامعتبر رو به پیش‌فرض برمی‌گردونن و برای هر اصلاح یک `Warning` تولید می‌کنن.
- فایل `aggregate.go` این توابع رو در `ValidateAll` و `NormalizeAll` ترکیب می‌کنه.

Startup، `NormalizeAll` رو اجرا می‌کنه و اصلاحات رو به کاربر نشون می‌ده. مقدارهای اصلاح‌شده در حافظه باقی می‌مونن و فقط وقتی روی دیسک ذخیره می‌شن که کاربر از طریق Inspector بخش مربوطه رو ذخیره کنه.

## فهرست‌های IP

- `parser.go`، تابع `StreamActiveIPs(ctx, path, limit, shuffled, out)` رو داره که Engine برای فرستادن IPها به Workerها ازش استفاده می‌کنه. تابع `StreamCIDR` هم برای پردازش یک پیشوند شبکه استفاده می‌شه.
- `csv.go` و `parser.go`، قالب CSV دو ستونی `<ip_or_cidr>,<enabled>` رو می‌خونن.
- `loader.go`، فهرست‌های خارجی رو بارگذاری می‌کنه و توابع `CountIPs` و `CountActiveIPs` رو برای محاسبه تعداد کل IPها در اختیار می‌ذاره.
- `registry.go`، فایل‌های موجود در `ips/` رو فهرست می‌کنه و `shuffle.go` ترتیب ورودی‌ها رو به‌هم می‌زنه.

ورودی‌های غیرفعال هنگام پردازش جریانی کنار گذاشته می‌شن. پیشوندهای IPv4 و IPv6 هر دو پشتیبانی می‌شن. بنابراین، شمارش IPهای یک پیشوند بزرگ IPv6 ممکنه به‌جای تعداد دقیق، به مقدار اشباع‌شده برسه.

## زیرسیستم DNS

- `query.go`، کوئری‌های DNS رو روی UDP، TCP و DoT می‌سازه و ارسال می‌کنه.
- `types.go`، ترنسپورت‌ها، نوع‌های Record، کدهای Rcode، روش‌های احراز هویت و نام پروتکل‌های تونل رو Parse می‌کنه.
- فایل‌های `dnstt.go`، `vaydns.go`، `slipstream.go`، `masterdns.go`، `stormdns.go` و `thefeed.go`، Structهای Config و اینترفیس‌های سرویس مربوط به هر پروتکل رو تعریف می‌کنن و فایل‌های تنظیمات زیر `assets/dns-tunneling/` رو مدیریت می‌کنن. اینترفیس‌ها شامل `DNSTTService`، `VayDNSService`، `SlipstreamService`، `MasterDNSService`، `StormDNSService` و `TheFeedService` هستن.
- `store.go`، ذخیره‌گاه عمومی فایل‌های Config رو فراهم می‌کنه.
- `registry.go`، تابع تجمیعی `GetAllDNSTunsFile` رو داره که تنظیمات هر شش سرویس رو کنار هم قرار می‌ده.
- `validate.go`، ابزارهای مشترک اعتبارسنجی رو نگه می‌داره.
- `embedded.go`، کلاینت‌های داخلی MasterDNS و StormDNS رو پوشش می‌ده.
- پکیج `socks/`، کلاینت SOCKS5 رو برای اعتبارسنجی تونل پس از برقراری اتصال فراهم می‌کنه.
- پکیج `ssh/`، کلاینت SSH رو برای اتصال‌های تونل‌شده از طریق پروکسی SSH فراهم می‌کنه.

## یکپارچگی Xray

- `service.go`، نمونه‌های Xray رو با استفاده از کتابخانه داخلی `xray-core` در همان پروسه اجرا و مدیریت می‌کنه. فایل‌های `inbound.go` و `outbound.go` تنظیمات JSON ورودی و خروجی رو می‌سازن.
- `link.go`، لینک‌های اشتراک رو Parse می‌کنه.
- فایل‌های `download.go`، `latency.go` و `upload.go` در پکیج `speedtest`، سرعت و تأخیر اتصال رو اندازه می‌گیرن.

`xrayprobe` برای Outbound انتخاب‌شده تنظیمات لازم رو می‌سازه، یک پورت اجاره می‌کنه، نمونه Xray رو داخل همون پروسه بالا می‌آره و تأخیر اتصال رو از طریق پروکسی محلی اندازه می‌گیره. در صورت نیاز، تست سرعت هم اجرا می‌شه و در پایان منابع استفاده‌شده آزاد می‌شن.

## مدیریت پروسه

پکیج `internal/core/process` شامل ابزار قدیمی ردیابی پروسه‌ها برای چند پلتفرمه، از جمله `process_unix.go` و `process_windows.go`.

اسکنر فعلی از این پکیج استفاده نمی‌کنه، اما کدش هنوز در پروژه وجود داره. تنها پروسه‌های فرزندی که اسکنر اجرا می‌کنه، دستورهای `termux-wake-lock` و `termux-wake-unlock` در اندروید هستن.

## Logger

| Logger | فایل | کاربرد |
|---|---|---|
| Core | `logs/core.log` | لاگ‌های Engine، Probeها، تنظیمات و عملیات ورودی و خروجی نتایج |
| UI | `logs/ui.log` | چرخه عمر کامپوننت‌ها، عملیات فایل و خطاهای رابط کاربری |

هر دو Logger مقدار `log_level` رو از `settings/ui_settings.toml` می‌خونن. علاوه بر نوشتن در فایل، لاگ‌ها رو به مشترک‌هایی که به‌صورت زنده اون‌ها رو دریافت می‌کنن هم می‌فرستن.

برای چرخش فایل‌های لاگ از lumberjack استفاده می‌شه. تنظیمات فعلی شامل سقف حجم ۵۰ مگابایت، نگهداری ۳ نسخه پشتیبان، حداکثر ۷ روز نگهداری و فشرده‌سازی فایل‌های قدیمیه.

## Startup

مرحله Startup درون TUI و در مسیر `internal/ui/main/startup` پیاده‌سازی شده. این مرحله بعد از انیمیشن Splash و قبل از نمایش Workspace اجرا می‌شه و وضعیت بررسی‌ها رو به‌صورت زنده در نوار وضعیت نشون می‌ده.

بررسی‌ها به‌صورت توابع خصوصی در `checklist.go` تعریف شدن. کامپوننت Startup با متدهای `Init` و `Update` این فرایند رو مدیریت می‌کنه و بررسی‌ها رو به‌ترتیب، داخل یک Goroutine اجرا می‌کنه.

ترتیب بررسی‌ها به این شکله:

```text
1. Logger       بررسی لاگرهای Core و UI. اجرای core.Init()
                برای ثبت Schemaهای نتایج Probeها

2. Config       اجرای store.Load(). در صورت نامعتبر بودن TOML،
                گزارش خطا. اجرای validate.NormalizeAll برای
                اصلاح مقدارهای خارج از محدوده

3. Xray         نمایش نسخه هسته داخلی و بررسی Outboundها

4. DNSTT        بررسی فایل‌های تنظیمات DNSTT

5. Slipstream   پیدا کردن و بارگذاری کتابخانه کلاینت
                و بررسی فایل‌های تنظیمات

6. VayDNS       بررسی فایل‌های تنظیمات VayDNS

7. MasterDNS    بررسی فایل‌های تنظیمات MasterDNS

8. StormDNS     بررسی فایل‌های تنظیمات StormDNS

9. TheFeed      بررسی فایل‌های تنظیمات TheFeed

10. App         انتظار برای فشردن Enter توسط کاربر
```

هر بررسی وضعیت خودش رو از طریق یک `reporter` گزارش می‌ده. پیام‌ها می‌تونن یکی از این سطح‌ها رو داشته باشن: `[INFO]`، `[SUCCESS]`، `[WARN]` یا `[ERROR]`.

خطاهای بحرانی باعث می‌شن بررسی‌های بعدی اجرا نشن. نبود کتابخانه Slipstream خطا رو گزارش می‌کنه، اما جلوی ادامه Startup رو نمی‌گیره. در این حالت فقط اسکن Slipstream قابل اجرا نیست.

بعد از پایان بررسی‌ها، کاربر با فشردن Enter وارد Workspace می‌شه.

در ساختار فعلی، `main.go` تنظیمات رو بارگذاری نمی‌کنه و بررسی‌های Startup رو هم انجام نمی‌ده. وظیفه اون ساخت لاگرها، ایجاد اپلیکیشنی که تم و تنظیمات رابط کاربری رو راه‌اندازی می‌کنه، و شروع برنامه Bubble Tea هست.

خطاهای خواندن تنظیمات، ثبت Schema، بارگذاری کتابخانه و اعتبارسنجی Config در خود رابط Startup نمایش داده می‌شن، نه به‌شکل خطاهای خاموش پیش از بالا آمدن TUI. مقدارهای اصلاح‌شده هم در حافظه باقی می‌مونن تا کاربر در صورت نیاز، اون بخش رو از طریق Inspector ذخیره کنه.

## صفحه‌های مرتبط

- [معماری](./architecture/): ساختار کلی پروژه و لایه‌ها
- [رابط کاربری](./ui/): معماری کامپوننت‌ها و تم‌ها

