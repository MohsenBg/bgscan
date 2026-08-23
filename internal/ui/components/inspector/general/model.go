package general

import (
	"strconv"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/config/validate"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/input/selectinput"
	"bgscan/internal/ui/components/basic/input/textinput"
	"bgscan/internal/ui/components/basic/input/toggleinput"
	"bgscan/internal/ui/components/basic/inspector"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

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
	descStopAfterFound     = "The maximum number of successful results before halting the scan. Set to 0 to scan all targets."
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
	state     *ui.AppState
	name      string
	id        ui.ComponentID
	inspector ui.Component
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) Mode() env.Mode     { return env.NormalMode }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }

func saveGeneral(state *ui.AppState) tea.Cmd {
	if err := state.Store.SaveGeneral(state.Config.General); err != nil {
		logger.UIError("Failed to save General settings: %v", err)
		return notice.NewNoticeCmd(state.Layout, "Failed to save General settings", err.Error(), notice.NOTICE_ERROR)
	}
	return nil
}

func saveWriter(state *ui.AppState) tea.Cmd {
	if err := state.Store.SaveWriter(state.Config.Writer); err != nil {
		logger.UIError("Failed to save Writer settings: %v", err)
		return notice.NewNoticeCmd(state.Layout, "Failed to save Writer settings", err.Error(), notice.NOTICE_ERROR)
	}
	return nil
}

func intInput(state *ui.AppState, title string, value int, validate func(string) error, set func(int), save func() tea.Cmd) input.Input[string] {
	return textinput.New(
		state.Layout, title,
		textinput.WithValue(strconv.Itoa(value)),
		textinput.WithValidation(validate),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, err := strconv.Atoi(v)
			if err != nil {
				return notice.NewNoticeCmd(state.Layout, "Invalid "+title, err.Error(), notice.NOTICE_ERROR)
			}
			set(n)
			return save()
		}),
	)
}

func durationMSInput(state *ui.AppState, title string, value time.Duration, validate func(string) error, set func(time.Duration), save func() tea.Cmd) input.Input[string] {
	return textinput.New(
		state.Layout, title,
		textinput.WithValue(strconv.FormatInt(value.Milliseconds(), 10)),
		textinput.WithValidation(validate),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, err := strconv.Atoi(v)
			if err != nil {
				return notice.NewNoticeCmd(state.Layout, "Invalid "+title, err.Error(), notice.NOTICE_ERROR)
			}
			set(time.Duration(n) * time.Millisecond)
			return save()
		}),
	)
}

func New(state *ui.AppState, name string) *Model {
	cfg := state.Config

	saveGeneralCmd := func() tea.Cmd { return saveGeneral(state) }
	saveWriterCmd := func() tea.Cmd { return saveWriter(state) }

	statusInterval := durationMSInput(state, "Enter Status Interval", cfg.General.StatusInterval.Duration(),
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.StatusInterval = config.NewDurationMS(time.Duration(n) * time.Millisecond)
			return fieldErr(validate.ValidateGeneral(tmp), "StatusInterval")
		},
		func(d time.Duration) { cfg.General.StatusInterval = config.NewDurationMS(d) }, saveGeneralCmd)

	stopAfterFound := intInput(state, "Enter Stop After Found (0 = unlimited)", cfg.General.StopAfterFound,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.StopAfterFound = n
			return fieldErr(validate.ValidateGeneral(tmp), "StopAfterFound")
		},
		func(n int) { cfg.General.StopAfterFound = n }, saveGeneralCmd)

	maxIPsToTest := intInput(state, "Enter Max IPs To Test (0 = unlimited)", cfg.General.MaxIPsToTest,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.MaxIPsToTest = n
			return fieldErr(validate.ValidateGeneral(tmp), "MaxIPsToTest")
		},
		func(n int) { cfg.General.MaxIPsToTest = n }, saveGeneralCmd)

	shuffled := toggleinput.New(
		state.Layout, "Shuffle",
		toggleinput.WithValue(cfg.General.Shuffled),
		toggleinput.WithFocus(),
		toggleinput.WithLabels("Enabled", "Disabled"),
		toggleinput.WithOnSubmit(func(v bool) tea.Cmd {
			cfg.General.Shuffled = v
			return saveGeneralCmd()
		}),
	)

	minProbeDuration := durationMSInput(
		state,
		"Enter Minimum Probe Duration",
		cfg.General.MinProbeDuration.Duration(),
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			tmp := cfg.General
			tmp.MinProbeDuration = config.NewDurationMS(time.Duration(n) * time.Millisecond)

			return fieldErr(validate.ValidateGeneral(tmp), "MinProbeDuration")
		},
		func(d time.Duration) {
			cfg.General.MinProbeDuration = config.NewDurationMS(d)
		},
		saveGeneralCmd,
	)

	probePerSec := intInput(
		state,
		"Enter Probe Rate Per Second",
		cfg.General.ProbePerSec,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			tmp := cfg.General
			tmp.ProbePerSec = n

			return fieldErr(validate.ValidateGeneral(tmp), "ProbePerSec")
		},
		func(n int) {
			cfg.General.ProbePerSec = n
		},
		saveGeneralCmd,
	)

	probeBurst := intInput(
		state,
		"Enter Probe Burst Size",
		cfg.General.ProbeBurst,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}

			tmp := cfg.General
			tmp.ProbeBurst = n

			return fieldErr(validate.ValidateGeneral(tmp), "ProbeBurst")
		},
		func(n int) {
			cfg.General.ProbeBurst = n
		},
		saveGeneralCmd,
	)

	pipelineMode := selectinput.New(
		state.Layout, "Select Pipeline Mode",
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

	maxIPsPerStage := intInput(state, "Enter Max IPs Per Stage", cfg.General.MaxIPsPerStage,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.MaxIPsPerStage = n
			return fieldErr(validate.ValidateGeneral(tmp), "MaxIPsPerStage")
		},
		func(n int) { cfg.General.MaxIPsPerStage = n }, saveGeneralCmd)

	batchSize := intInput(state, "Batch Size", cfg.General.BatchSize,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.General
			tmp.BatchSize = n
			return fieldErr(validate.ValidateGeneral(tmp), "BatchSize")
		},
		func(n int) { cfg.General.BatchSize = n }, saveGeneralCmd)

	// Writer settings
	mergeFlushInterval := durationMSInput(state, "Enter Merge Flush Interval", cfg.Writer.MergeFlushInterval.Duration(),
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.Writer
			tmp.MergeFlushInterval = config.NewDurationMS(time.Duration(n) * time.Millisecond)
			return fieldErr(validate.ValidateWriter(tmp), "MergeFlushInterval")
		},
		func(d time.Duration) { cfg.Writer.MergeFlushInterval = config.NewDurationMS(d) }, saveWriterCmd)

	chanSize := intInput(state, "Enter Channel Size", cfg.Writer.ChanSize,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.Writer
			tmp.ChanSize = n
			return fieldErr(validate.ValidateWriter(tmp), "ChanSize")
		},
		func(n int) { cfg.Writer.ChanSize = n }, saveWriterCmd)

	writerBatchSize := intInput(state, "Enter Batch Size", cfg.Writer.BatchSize,
		func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			tmp := cfg.Writer
			tmp.BatchSize = n
			return fieldErr(validate.ValidateWriter(tmp), "BatchSize")
		},
		func(n int) { cfg.Writer.BatchSize = n }, saveWriterCmd)

	resultDirectory := textinput.New(
		state.Layout, "Enter Result Directory",
		textinput.WithValue(cfg.Writer.ResultBaseDir),
		textinput.WithValidation(func(v string) error {
			tmp := cfg.Writer
			tmp.ResultBaseDir = v
			return fieldErr(validate.ValidateWriter(tmp), "ResultDirectory")
		}),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.Writer.ResultBaseDir = v
			return saveWriterCmd()
		}),
	)

	fields := []inspector.Field{
		{Name: "Status Interval", Description: descStatusInterval, Group: groupGeneral, Input: inspector.Adapt(statusInterval), Visible: alwaysVisible, Format: inspector.FormatDurationMS},
		{Name: "Stop After Found", Description: descStopAfterFound, Group: groupGeneral, Input: inspector.Adapt(stopAfterFound), Visible: alwaysVisible, Format: inspector.FormatIntOrUnlimited},
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
		state:     state,
		name:      name,
		id:        ui.NewComponentID(),
		inspector: inspector.New(state.Layout, "general settings", fields),
	}
}

func alwaysVisible() bool { return true }

func visibleWhenMode(cfg *config.GeneralConfig, want pipelineMode) func() bool {
	return func() bool { return cfg.PipelineMode == want }
}

func fieldErr(errs map[string]error, field string) error {
	return errs[field]
}
