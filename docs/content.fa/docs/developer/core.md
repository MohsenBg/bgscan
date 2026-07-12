---
title: "هسته اصلی"
weight: 3
---

# هسته اصلی (Core)

پکیج `internal/core` حاوی تمام منطق‌های غیر UI برنامه است: پیکربندی و تنظیمات، موتور اصلی اسکنر، پیاده‌سازی پروب‌ها، مدیریت لیست‌های آی‌پی، ثبت نتایج، دی‌ان‌اس، یکپارچه‌سازی Xray و مدیریت پروسس‌ها.

---

## نقشه پکیج‌ها (Package Map)

| پکیج | وظیفه و مسئولیت |
|---|---|
| `config` | پکیج سینگلتون و ایمن در برابر چندنخی که تمام تنظیمات TOML را نگه می‌دارد؛ همراه با ابزارهای کمکی بارگذاری/ذخیره و نقاط ورود اعتبارسنجی. |
| `config/validate` | اعتبارسنج‌های اختصاصی هر پروتکل (ICMP، TCP، HTTP، DNS، Xray، نویسنده نتایج، تنظیمات عمومی). |
| `scanner` | هماهنگ‌کننده ساختار `Scanner` و سازنده‌های مراحل اسکن (Stage Builders). |
| `scanner/engine` | اجرای چرخه پردازش اسکن: اسکن تک‌مرحله‌ای، زنجیره ترتیبی، خط‌لوله جریانی (Streaming) و خط‌لوله دسته‌ای (Batch). |
| `scanner/probe` | اینترفیس `Probe` و تمام پروب‌های واقعی (ICMP، TCP، HTTP، HTTP/3، DNS، DNSTT، SlipStream، Resolve، Xray). |
| `scanner/portmgr` | استخر پورت‌های زودگذر (Ephemeral Port Pool) برای پروب‌هایی که نیاز به بایند کردن پورت‌های محلی دارند (Xray، DNSTT، SlipStream). |
| `scanner/netutil` | ابزارهای کمکی مشترک شبکه. |
| `result` | ساختار `Writer` (ثبت ناهمگام دسته‌ای + ادغام)، فرمت فایل‌های CSV، بارگذاری‌کننده، رجیستری و مرتب‌سازی. |
| `iplist` | بارگذاری‌کننده فایل‌های CSV لیست آی‌پی، پارسر، رجیستری، برهم‌زننده تصادفی (Shuffle) و استریم داده‌ها. |
| `ip` | پارس کردن IPv4 و باز کردن رنج‌های CIDR. |
| `dns` | ابزارهای کمکی کوئری‌های DNS، پارس کردن لایه انتقال، کلاینت SOCKS5 ابزارهای DNSTT/SlipStream. |
| `xray` | اجراکننده باینری Xray، تنظیمات اینباند/آوتباند، پارس کردن لینک‌های اتصال و تست سرعت. |
| `process` | ایجاد/خاتمه پروسس‌ها به صورت چندسکویی (تفکیک مجزا برای سیستم‌عامل‌های ویندوز و یونیکس). |
| `fileutil` | ابزارهای کمکی برای فایل‌های CSV، JSON، TOML، متن، فایل‌های موقت و مسیرها. |
| `logger` | ثبت لاگ‌های سطح‌بندی‌شده (هسته، رابط کاربری، دباگ) همراه با قابلیت جابه‌جایی و چرخش خودکار فایل‌ها (Lumberjack). |

---

## اسکنر (Scanner)

فایل `internal/core/scanner/scanner.go` ساختار `Scanner` را تعریف می‌کند — این بخش همان API عمومی است که رابط کاربری (UI) برای اجرای یک اسکن آن را فراخوانی می‌کند.

```go
type Scanner struct {
    ctx    context.Context
    cancel context.CancelFunc
    pause  *engine.PauseController
    input  string
    pm     *portmgr.PortManager
    stages []StageConfig
}

```

ساختار `StageConfig` ویژگی‌هایی چون حالت اسکن (`ScanMode`)، تعداد ورکرها، `probe.Probe`، نویسنده نتایج (`result.Writer`)، محدودیت نرخ ارسال پکت (Rate Limit) و هوک‌های اختیاری اسکن (`ScanHooks`) را در خود بسته‌بندی می‌کند.

**چرخه حیات (Lifecycle):**

```text
NewScanner(ctx, input)
  ├─ AddStage(BuildICMPStage(ctx))
  ├─ AddStage(BuildTCPStage(ctx))
  └─ Run()
      ├─ تک‌مرحله‌ای ← engine.RunScan(ctx, input, maxIPs, cfg, shuffled, pause)
      └─ چندمرحله‌ای  ← engine.RunScanWithChain(ctx, input, maxIPs, chainCfg)

```

ساختار `Scanner` همچنین متدهای `Pause()` ،`Resume()` ،`IsPaused()` ،`PausedDuration()` و `Close()` را برای کنترل فرآیند اسکن از طریق رابط کاربری در اختیار ما می‌گذارد.

**سازنده‌های مراحل اسکن** (`BuildICMPStage` ،`BuildTCPStage` ،`BuildHTTPStage` ،`BuildXrayStage` ،`BuildResolveStage` ،`BuildDNSTTStage` ،`BuildSlipStreamStage`) هرکدام وظایف زیر را انجام می‌دهند:

1. تنظیمات پروتکل را از طریق متدهای `()config.GetXxx` می‌خوانند.
2. مسیر فایل نتایج را ساخته و یک `result.Writer` ایجاد می‌کنند.
3. پروب مناسب (`probe.Probe`) را متولد می‌کنند.
4. یک لایه تنظیمات `StageConfig` را همراه با ورکرها، نرخ ارسال و نویسنده نتایج برمی‌گردانند.

---

## موتور اسکن (Engine)

پکیج `internal/core/scanner/engine` قلب تپنده اجرای اسکن است. این بخش از جزئیات و کارکرد داخلی پروب‌ها بی‌خبر است — وظیفه آن صرفاً جابه‌جایی آی‌پی‌ها و انتقال نتایج است.

#### اسکن تک‌مرحله‌ای: `RunScan`

فایل `engine/scan.go` یک مرحله اسکن مستقل را هماهنگ می‌کند:

```text
خواننده داده‌ها (iplist.StreamActiveIPs)
   │
   ▼
 کانال آی‌پی‌ها (ips channel) ──► استخر ورکرها (شامل N گوروتین)
                             │
                             ├─ محدودکننده نرخ ارسال (rateCh)
                             ├─ probe.Run(ctx, ip)
                             ├─ در صورت موفقیت: کانال نتایج + هوک OnSuccess
                             └─ در صورت خطا: ثبت لاگ + هوک OnError
                             │
                             ▼
                    کانال نتایج ──► گوروتین نویسنده ──► دیسک (ادغام CSV)

گوروتین گزارش پیشرفت → هوک OnProgress (در بازه‌های زمانی مشخص شده در status_interval)

```

متد `RunScan` تا زمانی که خواننده داده‌ها کارش تمام شود، ورکرها کاملاً تخلیه شوند، نویسنده نتایج داده‌ها را روی دیسک بنویسد و آخرین گزارش پیشرفت ارسال شود، عملیات را به صورت مسدودکننده (Blocking) نگه می‌دارد.

#### اسکن زنجیره‌ای (چندمرحله‌ای): `RunScanWithChain`

فایل `engine/chain.go` وظیفه هدایت و توزیع کار را بر اساس حالت چرخه پردازش (`PipelineMode`) بر عهده دارد:

| حالت چرخه پردازش | تابع مربوطه | نحوه اتصال مراحل به یکدیگر |
| --- | --- | --- |
| `sequential` | `executeSequentialChain` | مرحله N نتایج را روی دیسک می‌نویسد؛ مرحله N+1 همان فایل را به عنوان ورودی خود می‌خواند. |
| `streaming` | `executeStreamingPipeline` | استفاده از کانال‌های بافر شده بین مراحل. تمام مراحل به صورت هم‌زمان اجرا می‌شوند. |
| `batch` | `executeBatchPipeline` | آی‌پی‌ها به صورت دسته‌های کوچک خرد می‌شوند؛ هر دسته قبل از شروع دسته بعدی، تمام مراحل را طی می‌کند. |

**حالت ترتیبی (Sequential)** — کمترین میزان مصرف حافظه رم، کمترین سرعت. هر مرحله پیش از شروع مرحله بعدی، باید کاملاً به پایان برسد. مسیر خروجی نویسنده نتایج، به فایل ورودی مرحله بعدی تبدیل می‌شود.

**حالت جریانی (Streaming)** — بیشترین بازدهی و سرعت. متد `createStageChannels` کانال‌های بافر شده از نوع `chan string` بین مراحل ایجاد می‌کند. اندازه کانال معادل با `max(workers, MaxIPsPerStage)` است. آی‌پی‌های موفق بلافاصله از طریق دستور `output <- ip` به مرحله بعدی سرازیر می‌شوند.

**حالت دسته‌ای (Batch)** — حالتی ترکیبی. متد `streamIPsFromFile` آی‌پی‌ها را به آرایه‌هایی به اندازه `batchSize` خرد می‌کند. متد `processBatch` هر آرایه را به صورت ترتیبی از میان تمام لایه‌های `stageExecutor` عبور می‌دهد. تا زمانی که دسته فعلی تمام مراحل را به پایان نرساند، فرآیند دسته بعدی آغاز نخواهد شد.

#### انواع ساختارها (Types)

```go
type ChainConfig struct {
    Mode      PipelineMode    // sequential | streaming | batch
    MaxBuffer int             // اندازه بافر کانال بین مراحل جریانی
    Stages    []ScanConfig
    Pause     *PauseController
    Shuffled  bool
}

type ScanConfig struct {
    Workers int
    Rate    int
    Probe   probe.Probe
    Writer  *result.Writer
    Hooks   ScanHooks
}

type ScanHooks struct {
    OnProgress func(Progress)
    OnSuccess  func(result.IPScanResult)
    OnScanEnd  func()
    OnError    func(error)
}

```

تمام هوک‌ها اختیاری هستند (مقدار nil به معنای غیرفعال بودن است). موتور اسکن متدهای `callOnSuccess` ،`callOnError` و `callOnScanEnd` را به صورت ایمن فراخوانی می‌کند.

#### کنترل وضعیت توقف (Pause Control)

فایل `engine/pause.go` ساختار `PauseController` را ارائه می‌دهد — مکانیزمی غیرمسدودکننده (Non-blocking) برای متوقف کردن و از سر گیری اسکن. ورکرها متد `pause.IsPaused()` را بررسی کرده و در زمان توقف، کار جدیدی برنمی‌دارند. متد `PausedDuration()` مجموع زمان‌های توقف را محاسبه می‌کند تا گزارش‌های پیشرفت اسکن کاملاً دقیق باقی بمانند.

---

## اینترفیس پروب (Probe Interface)

فایل `scanner/probe/probe.go` تنها قراردادی را تعریف می‌کند که تمام پروب‌ها ملزم به پیاده‌سازی آن هستند:

```go
type Probe interface {
    Init(ctx context.Context) error
    Run(ctx context.Context, ip string) (*result.IPScanResult, error)
    Close() error
}

```

* **`Init`** — یک‌بار در زمان راه‌اندازی فراخوانی می‌شود. برای تخصیص سوکت‌ها، متولد کردن گوروتین‌ها و باز کردن کش‌ها.
* **`Run`** — به ازای هر آی‌پی فراخوانی می‌شود. این متد باید از پارامتر `ctx` برای قابلیت لغو عملیات تبعیت کند. در صورت موفقیت، ساختار `IPScanResult` و در صورت خطا یا اتمام مهلت زمانی، یک خطای `error` برمی‌گرداند.
* **`Close`** — یک‌بار در زمان خاموش شدن سیستم فراخوانی می‌شود. برای آزادسازی سوکت‌ها، گوروتین‌ها و توصیف‌گرهای فایل (File Descriptors).

#### پروب‌های موجود

| فایل پیاده‌سازی | نوع پروب | تابع سازنده |
| --- | --- | --- |
| `icmp.go` | پکت‌های تست ICMP echo | `NewICMPProbe(timeout, tries)` |
| `tcp.go` | برقراری اتصال TCP connect | `NewTCPProbe(port, timeout, tries)` |
| `http.go` | پروتکل‌های HTTP/1.1 و HTTP/2 (ALPN) | `NewHTTPProbe(reqCfg, acceptedCodes)` |
| `http3.go` | پروتکل HTTP/3 بر بستر QUIC | `NewHTTP3Probe(reqCfg, acceptedCodes)` |
| `httpshare.go` | سازنده مشترک تنظیمات درخواست HTTP | `NewHTTPRequestFromConfig` / `NewHTTP3RequestFromConfig` |
| `resolve.go` | ریزالور DNS (بررسی رکوردهای A/AAAA و DPI) | `NewResolverProbe(DnsRequest)` |
| `dnstt.go` | اعتبارسنجی تونل DNSTT | `NewDNSTTProbe(DNSTTConfig, portMgr)` |
| `slipstream.go` | اعتبارسنجی تونل SOCKS ابزار SlipStream | `NewSlipstreamProbe(workers, SlipstreamConfig, portMgr)` |
| `xray.go` | بررسی اتصال آوتباند و پهنای باند Xray | `NewXrayProbe(cfg, template, portMgr)` |
| `processes.go` | چرخه حیات پروسس‌های Xray/DNSTT/Slipstream | مورد استفاده پروب‌هایی که باینری‌های خارجی اجرا می‌کنند |

پروب‌هایی که نیاز به بایند کردن پورت‌های محلی دارند (Xray، DNSTT، SlipStream) یک نمونه از `*portmgr.PortManager` را دریافت می‌کنند — این ساختار یک استخر پورت‌های زودگذر است که برای جلوگیری از تداخل پورت‌ها، پورت‌های جدیدی را تخصیص داده و پورت‌های قبلی را بازیافت می‌کند.

---

## تنظیمات (Config)

فایل `internal/core/config/config.go` یک پکیج سینگلتون ایمن در برابر چندنخی را در دسترس لایه‌های دیگر می‌گذارد:

```go
func Get() *ScannerConfig          // ساختار اصلی سینگلتون
func GetGeneral() *GeneralConfig   // دسترسی مجزا بر حسب پروتکل (دارای قفل‌های خواندن/نوشتن)
func GetTCP() *TCPConfig
// ... گزینه‌های مربوط به ICMP, HTTP, Xray, DNS, Writer

```

**چیدمان فایل‌ها** — هر پروتکل فایل TOML اختصاصی خود را در دایرکتوری `settings/` دارد:

```text
settings/
├── general_settings.toml
├── writer_settings.toml
├── icmp_settings.toml
├── tcp_settings.toml
├── http_settings.toml
├── xray_settings.toml
└── dns_settings.toml

```

**گردش کار بارگذاری (Load Flow):**

1. متد `config.Init()` تمام توابع `()LoadXxxConfig` را به صورت ترتیبی فراخوانی می‌کند.
2. هر لودر فایل TOML مربوطه را می‌خواند یا در صورت عدم وجود، به سراغ مقادیر پیش‌فرض (نسخه‌های با پسوند `.default`) می‌رود.
3. ساختار بارگذاری‌شده از طریق یک متد ثبت‌کننده خصوصی (که دارای قفل نوشتن یا Write-lock است)، جایگزین فیلد مربوطه در سینگلتون می‌شود.
4. بلافاصله پس از اجرای `()Init`، متد `()startup.checkConfigHealth` اعتبارسنج‌های موجود در مسیر `config/validate/` را اجرا می‌کند تا مقادیر را نرمال‌سازی کند.

**گردش کار ذخیره‌سازی (Save Flow):**

1. بخش بازرس رابط کاربری (UI Inspector) متد `SaveXxxConfig(cfg)` را صدا می‌زند.
2. متد `saveConfig` داده‌های ساختار TOML را از طریق ابزار کمکی `fileutil.WriteTOMLFile` روی دیسک می‌نویسد.
3. فیلد سینگلتون موجود در حافظه رم (In-memory) از طریق متد ثبت‌کننده به‌روزرسانی می‌شود.

#### اعتبارسنج‌ها (Validators)

پوشه `config/validate/` حاوی یک اعتبارسنج مجزا به ازای هر پروتکل است (`validate_icmp.go` ،`validate_tcp.go` و غیره) — که همگی از مسیر `validate/all.go` فراخوانی می‌شوند. آن‌ها مقادیر را در محدوده‌های امن قفل کرده و خطاهای پیکربندی را در زمان راه‌اندازی، یعنی قبل از اینکه اسکنر بخواهد شروع به کار کند، آشکار می‌سازند.

---

## چرخه پردازش نتایج (Result Pipeline)

پکیج `internal/core/result/` وظیفه نوشتن خروجی اسکن‌ها را روی دیسک بر عهده دارد.

#### نویسنده نتایج (Writer)

ساختار `result.Writer` یک نویسنده ناهمگام دسته‌ای همراه با قابلیت ادغام است:

```go
type Writer struct {
    config     Config
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    resultPath string
    input      chan IPScanResult
    batch      []IPScanResult
    batchSize  int
}

```

**محرک‌های تخلیه داده روی دیسک (Flush Triggers):**

* انباشته شدن نتایج به تعداد تعیین‌شده در `BatchSize`
* سررسیدن بازه زمانی تعیین‌شده در `MergeFlushInterval`
* فراخوانی متد `()Stop` (تخلیه نهایی در زمان خاموش شدن سیستم)

نویسنده نتایج تضمین می‌کند که تک‌تک نتایجی که قبل از فراخوانی `()Stop` وارد کانال شده‌اند، حتماً روی دیسک نوشته شوند.

#### ادغام نتایج (Merge)

فایل `result/merger.go` یک سیستم ادغام جریانی و ایمن در برابر کرش (Crash-safe) را پیاده‌سازی می‌کند:

* دسته نتایج جدید را مرتب کرده و آن را با فایل موجود روی دیسک ادغام و مرتب (Merge-sort) می‌کند.
* داده‌ها را ابتدا در یک فایل موقت با پسوند `resultPath.tmp` می‌نویسد، دستور `fsync` را صادر می‌کند و سپس به صورت یکپارچه و اتمیک (Atomic Rename) نام آن را به مسیر نهایی تغییر می‌دهد.
* آی‌پي‌های تکراری با رکوردهای جدیدتر جایگزین می‌شوند.
* میزان مصرف حافظه رم در این روش کاملاً ثابت (Constant Memory) است؛ زیرا هیچ‌وقت هر دو فایل به صورت کامل درون حافظه بارگذاری نمی‌شوند.

#### فرمت نتایج

هر فایل نتایج یک فایل CSV است که ستون‌های آن دقیقاً با فیلدهای ساختار `IPScanResult` مطابقت دارد:

```go
type IPScanResult struct {
    IP        string
    Latency   time.Duration
    Download  time.Duration
    Upload    time.Duration
}

```

فایل‌ها بر اساس نوع اسکن در پوشه‌های مربوطه سازماندهی می‌شوند:

```text
results/
├── icmp/
├── tcp/
├── http/
├── xray/
├── dnstt/
├── slipstream/
└── resolve/

```

#### رجیستری و بارگذاری‌کننده (Registry and Loader)

* `registry.go` — فایل‌های نتایج را بر اساس نوع آن‌ها از درون دایرکتوری نتایج کشف و دسته‌بندی می‌کند.
* `loader.go` — نتایج را به صورت استریم از فایل‌های CSV بازمی‌خواند (این ویژگی زمان اسکن مجدد از روی لیست نتایج کاربرد دارد).
* `count.go` — بدون بارگذاری کامل فایل در حافظه رم، تعداد رکوردهای موجود در آن را شمارش می‌کند.

---

## لیست‌های آی‌پی (IP Lists)

پکیج `internal/core/iplist/` بارگذاری فایل‌های ورودی را مدیریت می‌کند:

* `loader.go` — تابع `(StreamActiveIPs(ctx, path, maxIP, shuffled, out` را تعریف می‌کند؛ این همان تابعی است که موتور اسکن برای تغذیه استخر ورکرها آن را فراخوانی می‌کند.
* `parser.go` / `csv.go` — فرمت داخلی فایل‌های CSV دو ستونی را پارس می‌کند (`<ip_or_cidr>,<enable>`).
* `registry.go` — فایل‌های آی‌پی موجود در پوشه `ips/` را لیست و مدیریت می‌کند.
* `shuffle.go` — ترتیب آی‌پی‌های هدف را به صورت تصادفی برهم می‌زند.
* `ip/expand.go` — رنج‌های CIDR را به آی‌پی‌های تکی بسط می‌دهد.

ورودی‌های غیرفعال (`enable=0`) در طول فرآیند استریم به صورت خودکار و بی‌صدا نادیده گرفته می‌شوند.

---

## زیرسیستم دی‌ان‌اس (DNS Subsystem)

پکیج `internal/core/dns/` ابزارهای کمکی مشترکی را ارائه می‌دهد که توسط پروب‌های ریزالور، DNSTT و SlipStream استفاده می‌شوند:

* `query.go` — ساخت و ارسال کوئری‌های DNS در سطوح پایین (پشتیبانی از UDP، TCP و DNS-over-TLS).
* `type.go` — ابزارهای کمکی برای پارس کردن انواع رکوردها و کدهای وضعیت دی‌ان‌اس (`ParseDNSRcode` ،`ParseTransport`).
* `dnstt.go` / `slipstream.go` — رپرهای کلاینت‌های تونل‌زنی.
* `socks5.go` — یک کلاینت مینیمال SOCKS5 جهت اعتبارسنجی و تست پایداری تونل‌ها.

پروب ریزالور DNS از بررسی‌های فایروال (DPI)، اندازه‌های بافر EDNS0، ساب‌دامین‌های تصادفی و انواع رکوردهای قابل تنظیم پشتیبانی می‌کند.

---

## یکپارچه‌سازی Xray (Xray Integration)

پکیج `internal/core/xray/` تعامل با باینری Xray را مدیریت می‌کند:

* `xray.go` / `command.go` — اجرای پروسس Xray و کنترل آن.
* `inbound.go` / `outbound.go` — تولید فایل پیکربندی JSON برای اینباندها و آوتباندهای Xray.
* `link.go` — پارس کردن لینک‌های اشتراک‌گذاری Xray (مانند پروتکل‌های vless:// ،vmess:// و غیره).
* `speedtest.go` — سنجش و اندازه‌گیری پهنای باند آپلود و دانلود.

پروب Xray (`probe/xray.go`) تمام این قطعات را به یکدیگر متصل می‌کند: این پروب یک کانفیگ تولید کرده، پروسس Xray را اجرا می‌کند، وضعیت اتصال را تست کرده، در صورت تمایل تست سرعت می‌گیرد و در نهایت پروسس را متوقف کرده و ردپای آن را پاک می‌کند.

---

## مدیریت پروسس‌ها (Process Management)

پکیج `internal/core/process/` چرخه حیات پروسس‌ها را به صورت مستقل از پلتفرم پیاده‌سازی می‌کند:

* `process.go` — اینترفیس و رابط مشترک.
* `process_unix.go` — اجرای اختصاصی سیستم‌عامل‌های یونیکس (تنظیم تکه‌های مربوط به setsid و مدیریت سیگنال‌ها).
* `process_windows.go` — اجرای اختصاصی ویندوز (شامل پرچم CREATE_NEW_PROCESS_GROUP و ابزار taskkill).

این پکیج توسط پروب‌هایی که باینری‌های خارجی را متولد می‌کنند (Xray، DNSTT، SlipStream) استفاده می‌شود.

---

## لاگر (Logger)

پکیج `internal/logger/` سه جریان لاگ‌گیری سطح‌بندی‌شده را ارائه می‌دهد:

| نام لاگر | نام فایل | حوزه پوشش |
| --- | --- | --- |
| هسته (Core) | `logs/core.log` | موتور اسکنر، پروب‌ها، تنظیمات، ورودی/خروجی نتایج |
| رابط کاربری (UI) | `logs/ui.log` | چرخه حیات کامپوننت‌ها، عملیات فایل‌ها، خطاهای سطح UI |
| دباگ (Debug) | `logs/debug.log` | دامت‌های کامل وضعیت برنامه، وضعیت گوروتین‌ها، ردپای دقیق خطاها |

فایل‌های لاگ از طریق ابزار lumberjack به صورت خودکار جابه‌جا و بچرخند: سقف ۵۰ مگابایت، نگهداری تا ۳ نسخه پشتیبان، حفظ به مدت ۷ روز و فشرده‌سازی با فرمت gzip.

---

## راه‌اندازی (Startup)

فایل `internal/startup/health.go` تابع `()RunHealthChecks` را پیش از بالا آمدن TUI اجرا می‌کند:

```text
1. checkLoggerHealth()     → باز کردن فایل‌های لاگ، راه‌اندازی چرخه لاگر
2. theme.Init()            → تعیین پالت رنگی تاریک/روشن سیستم
3. checkConfigHealth()     → فراخوانی config.Init() + validate/all.go
4. checkXrayHealth()       → پیدا کردن محل باینری Xray + قالب‌های پیکربندی
5. checkDNSTTHealth()      → پیدا کردن محل باینری کلاینت DNSTT
6. checkSlipstreamHealth() → پیدا کردن محل باینری کلاینت SlipStream

```

اگر باینری‌های اختیاری (Xray/DNSTT/SlipStream) پیدا نشوند، برنامه یک اخطار (Warning) ثبت کرده و نوع اسکن وابسته به آن را غیرفعال می‌کند — اما خود برنامه همچنان بدون مشکل بالا می‌آید. وجود هرگونه اشکال در اعتبارسنجی تنظیمات (Config Validation) یک خطای مهلک (Fatal) محسوب شده و برنامه را متوقف می‌کند.

> 💡 **نکته:** برای حذف وقفه‌های ۵۰۰ میلی‌ثانیه‌ای بین مراحل بررسی در زمان دباگ، می‌توانید گزینه `fastboot = true` را در فایل `health.go` فعال کنید.

---

## صفحات مرتبط

* [معماری (Architecture)](../architecture/) — چیدمان کلی و سطح بالای پروژه
* [رابط کاربری (UI)](../ui/) — معماری سیستم TUI
