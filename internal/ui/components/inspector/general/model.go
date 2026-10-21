package general

import (
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/config/validate"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/toggleinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

const (
	groupGeneral = "General"
	groupWriter  = "Writer"
)

type pipelineMode = string

const (
	pipelineSequential pipelineMode = "sequential"
	pipelineStreaming  pipelineMode = "streaming"
	pipelineBatch      pipelineMode = "batch"
)

const (
	descStatusInterval     = "The interval in milliseconds for pushing status updates to the UI."
	descMaxSuccessfulIPs   = "The maximum number of successful results before halting the scan. Set to 0 to scan all targets."
	descMaxIPsToTest       = "The maximum number of IPs to read from the input source. Set to 0 to read all available IPs."
	descShuffled           = "Randomizes the target IP order before scanning to prevent subnet slamming and reduce firewall alerts."
	descMinProbeDuration   = "The minimum duration between probes to control scan speed and reduce target/network overload."
	descProbePerSec        = "Maximum number of probes allowed per second globally. Controls scan rate to prevent device overload."
	descProbeBurst         = "Maximum burst size allowed before rate limiting. Higher values allow short traffic spikes."
	descPipelineMode       = "The execution mode for multi-stage scanning: 'sequential' (disk-based), 'streaming' (channel-based), or 'batch' (hybrid)."
	descMaxIPsPerStage     = "The maximum number of IPs a pipeline stage can hold in memory. Exceeding this limit blocks the previous stage."
	descBatchSize          = "The number of IPs processed per batch when using the 'batch' pipeline mode."
	descMergeFlushInterval = "The interval in milliseconds for merging delta results into the main result file."
	descChanSize           = "The capacity of the internal channel used by scanner workers to send IP scan results to the writer goroutine."
	descWriterBatchSize    = "The initial capacity of the in-memory batch used to accumulate IP scan results before flushing to disk."
	descResultDirectory    = "The directory name used to store scan results. The directory is created inside the project root (for example, 'result' stores results in './result')."
)

type Model struct {
	session   *ui.Session
	name      string
	id        ui.ComponentID
	inspector ui.Component
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Init() tea.Cmd      { return m.inspector.Init() }
func (m *Model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

func saveGeneral(session *ui.Session) tea.Cmd {
	if err := session.Store.SaveGeneral(session.Config.General); err != nil {
		session.Env().Log.Error("Failed to save General settings: %v", err)
		return notice.NewNoticeCmd(session.Env(), "Failed to save General settings", err.Error(), notice.NOTICE_ERROR)
	}
	return nil
}

func saveWriter(session *ui.Session) tea.Cmd {
	if err := session.Store.SaveWriter(session.Config.Writer); err != nil {
		session.Env().Log.Error("Failed to save Writer settings: %v", err)
		return notice.NewNoticeCmd(session.Env(), "Failed to save Writer settings", err.Error(), notice.NOTICE_ERROR)
	}
	return nil
}

func intInput(session *ui.Session, title string, value int, validate func(string) error, set func(int), save func() tea.Cmd) input.Input[string] {
	return textinput.New(session.Env(), title,
		textinput.WithValue(strconv.Itoa(value)),
		textinput.WithValidation(validate),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, err := strconv.Atoi(v)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Invalid "+title, err.Error(), notice.NOTICE_ERROR)
			}
			set(n)
			return save()
		}),
	)
}

func durationMSInput(session *ui.Session, title string, value time.Duration, validate func(string) error, set func(time.Duration), save func() tea.Cmd) input.Input[string] {
	return textinput.New(session.Env(), title,
		textinput.WithValue(strconv.FormatInt(value.Milliseconds(), 10)),
		textinput.WithValidation(validate),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, err := strconv.Atoi(v)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Invalid "+title, err.Error(), notice.NOTICE_ERROR)
			}
			set(time.Duration(n) * time.Millisecond)
			return save()
		}),
	)
}

func New(session *ui.Session, name string) *Model {
	cfg := session.Config

	saveGeneralCmd := func() tea.Cmd { return saveGeneral(session) }
	saveWriterCmd := func() tea.Cmd { return saveWriter(session) }

	statusInterval := durationMSInput(session, "Enter Status Interval", cfg.General.StatusInterval.Duration(),
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.StatusInterval = config.NewDurationMS(time.Duration(n) * time.Millisecond)
			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldStatusInterval)
		},
		func(d time.Duration) { cfg.General.StatusInterval = config.NewDurationMS(d) }, saveGeneralCmd)

	maxSuccessfulIPs := intInput(session, "Enter Max Successful IPs (0 = unlimited)", cfg.General.MaxSuccessfulIPs,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.MaxSuccessfulIPs = n
			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldMaxSuccessfulIPs)
		},
		func(n int) { cfg.General.MaxSuccessfulIPs = n }, saveGeneralCmd)

	maxIPsToTest := intInput(session, "Enter Max IPs To Test (0 = unlimited)", cfg.General.MaxIPsToTest,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.MaxIPsToTest = n
			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldMaxIPsToTest)
		},
		func(n int) { cfg.General.MaxIPsToTest = n }, saveGeneralCmd)

	shuffled := toggleinput.New(session.Env(), "Shuffle Input Order",
		toggleinput.WithValue(cfg.General.Shuffled),
		toggleinput.WithFocus(),
		toggleinput.WithLabels("Enabled", "Disabled"),
		toggleinput.WithOnSubmit(func(v bool) tea.Cmd {
			cfg.General.Shuffled = v
			return saveGeneralCmd()
		}),
	)

	minProbeDuration := durationMSInput(
		session,
		"Enter Minimum Probe Duration (ms)",
		cfg.General.MinProbeDuration.Duration(),
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			tmp := cfg.General
			tmp.MinProbeDuration = config.NewDurationMS(time.Duration(n) * time.Millisecond)

			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldMinProbeDuration)
		},
		func(d time.Duration) {
			cfg.General.MinProbeDuration = config.NewDurationMS(d)
		},
		saveGeneralCmd,
	)

	probePerSec := intInput(
		session,
		"Enter Probe Rate Per Second (ms)",
		cfg.General.ProbePerSec,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			tmp := cfg.General
			tmp.ProbePerSec = n

			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldProbePerSec)
		},
		func(n int) {
			cfg.General.ProbePerSec = n
		},
		saveGeneralCmd,
	)

	probeBurst := intInput(
		session,
		"Enter Probe Burst Size",
		cfg.General.ProbeBurst,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			tmp := cfg.General
			tmp.ProbeBurst = n

			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldProbeBurst)
		},
		func(n int) {
			cfg.General.ProbeBurst = n
		},
		saveGeneralCmd,
	)

	pipelineMode := selectinput.New(session.Env(), "Select Pipeline Mode",
		selectinput.WithValue(cfg.General.PipelineMode),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("Sequential", pipelineSequential),
			huh.NewOption("Streaming", pipelineStreaming),
			huh.NewOption("Batch", pipelineBatch),
		),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.General.PipelineMode = v
			return saveGeneralCmd()
		}),
	)

	maxIPsPerStage := intInput(session, "Enter Max IPs Per Stage", cfg.General.MaxIPsPerStage,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.MaxIPsPerStage = n
			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldMaxIPsPerStage)
		},
		func(n int) { cfg.General.MaxIPsPerStage = n }, saveGeneralCmd)

	batchSize := intInput(session, "Batch Size", cfg.General.BatchSize,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.BatchSize = n
			return fieldErr(validate.ValidateGeneral(tmp), validate.FieldBatchSize)
		},
		func(n int) { cfg.General.BatchSize = n }, saveGeneralCmd)

	// Writer settings
	mergeFlushInterval := durationMSInput(session, "Enter Merge Flush Interval", cfg.Writer.MergeFlushInterval.Duration(),
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.Writer
			tmp.MergeFlushInterval = config.NewDurationMS(time.Duration(n) * time.Millisecond)
			return fieldErr(validate.ValidateWriter(tmp), validate.FieldMergeFlushInterval)
		},
		func(d time.Duration) { cfg.Writer.MergeFlushInterval = config.NewDurationMS(d) }, saveWriterCmd)

	chanSize := intInput(session, "Enter Channel Size", cfg.Writer.ChanSize,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.Writer
			tmp.ChanSize = n
			return fieldErr(validate.ValidateWriter(tmp), validate.FieldChanSize)
		},
		func(n int) { cfg.Writer.ChanSize = n }, saveWriterCmd)

	writerBatchSize := intInput(session, "Enter Batch Size", cfg.Writer.BatchSize,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.Writer
			tmp.BatchSize = n
			return fieldErr(validate.ValidateWriter(tmp), validate.FieldBatchSize)
		},
		func(n int) { cfg.Writer.BatchSize = n }, saveWriterCmd)

	resultDirectory := textinput.New(session.Env(), "Enter Result Directory",
		textinput.WithValue(cfg.Writer.ResultBaseDir),
		textinput.WithValidation(func(v string) error {
			tmp := cfg.Writer
			tmp.ResultBaseDir = v
			return fieldErr(validate.ValidateWriter(tmp), validate.FieldResultBaseDir)
		}),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.Writer.ResultBaseDir = v
			return saveWriterCmd()
		}),
	)

	fields := []inspector.Field{
		{Name: "Status Interval", Description: descStatusInterval, Group: groupGeneral, Input: inspector.Adapt(statusInterval), Visible: alwaysVisible, Format: inspector.FormatDurationMS},
		{Name: "Max Successful IPs", Description: descMaxSuccessfulIPs, Group: groupGeneral, Input: inspector.Adapt(maxSuccessfulIPs), Visible: alwaysVisible, Format: inspector.FormatIntOrUnlimited},
		{Name: "Max IPs To Test", Description: descMaxIPsToTest, Group: groupGeneral, Input: inspector.Adapt(maxIPsToTest), Visible: alwaysVisible, Format: inspector.FormatIntOrUnlimited},
		{Name: "Shuffled", Description: descShuffled, Group: groupGeneral, Input: inspector.Adapt(shuffled), Visible: alwaysVisible, Format: inspector.FormatBool},
		{Name: "Minimum Probe Duration", Description: descMinProbeDuration, Group: groupGeneral, Input: inspector.Adapt(minProbeDuration), Visible: alwaysVisible, Format: inspector.FormatDurationMS},
		{Name: "Probes Per Second", Description: descProbePerSec, Group: groupGeneral, Input: inspector.Adapt(probePerSec), Visible: alwaysVisible, Format: inspector.FormatInt},
		{Name: "Probe Burst", Description: descProbeBurst, Group: groupGeneral, Input: inspector.Adapt(probeBurst), Visible: alwaysVisible, Format: inspector.FormatInt},
		{Name: "Pipeline Mode", Description: descPipelineMode, Group: groupGeneral, Input: inspector.Adapt(pipelineMode), Visible: alwaysVisible},
		{Name: "Max IPs Per Stage", Description: descMaxIPsPerStage, Group: groupGeneral, Input: inspector.Adapt(maxIPsPerStage), Visible: visibleWhenMode(&cfg.General, pipelineStreaming), Format: inspector.FormatInt},
		{Name: "Batch Size", Description: descBatchSize, Group: groupGeneral, Input: inspector.Adapt(batchSize), Visible: visibleWhenMode(&cfg.General, pipelineBatch), Format: inspector.FormatInt},

		{Name: "Result Directory", Description: descResultDirectory, Group: groupWriter, Input: inspector.Adapt(resultDirectory), Visible: alwaysVisible},
		{Name: "Merge Flush Interval", Description: descMergeFlushInterval, Group: groupWriter, Input: inspector.Adapt(mergeFlushInterval), Visible: alwaysVisible, Format: inspector.FormatDurationMS},
		{Name: "Channel Size", Description: descChanSize, Group: groupWriter, Input: inspector.Adapt(chanSize), Visible: alwaysVisible, Format: inspector.FormatInt},
		{Name: "Batch Size", Description: descWriterBatchSize, Group: groupWriter, Input: inspector.Adapt(writerBatchSize), Visible: alwaysVisible, Format: inspector.FormatInt},
	}

	return &Model{
		session:   session,
		name:      name,
		id:        ui.NewComponentID(),
		inspector: inspector.New(session.Env(), "general settings", fields),
	}
}

func alwaysVisible() bool { return true }

func visibleWhenMode(cfg *config.GeneralConfig, want pipelineMode) func() bool {
	return func() bool { return cfg.PipelineMode == want }
}

func fieldErr(errs map[validate.Field]error, field validate.Field) error {
	return errs[field]
}
