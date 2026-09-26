package uisettings

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/toggleinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

const (
	groupAppearance = "Appearance"
)

const (
	descTheme           = "Color theme for the interface. Auto follows the terminal background."
	descSolidBackground = "Paint the theme background color behind the app. Disabled keeps the terminal background."
	descMaxRow          = "Maximum number of rows kept for display in result tables."
	descLogLevel        = "Minimum severity written to log files. Debug logs every probe error with the full target IP."
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

func New(state *ui.AppState, name string) *Model {
	cfg := state.Settings

	// A hand-edited settings file may name a theme that no longer exists.
	// Fall back to Auto in that case instead of showing a blank select.
	themes := theme.List()
	current := cfg.ThemeName
	if current != "" && !slices.Contains(themes, current) {
		current = ""
	}

	themeOptions := make([]huh.Option[string], 0, len(themes)+1)
	themeOptions = append(themeOptions, huh.NewOption("Auto (terminal)", ""))
	for _, t := range themes {
		themeOptions = append(themeOptions, huh.NewOption(t, t))
	}

	themeSelect := selectinput.New(state.Deps(), "Select Color Theme",
		selectinput.WithValue(current),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(themeOptions...),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			next, err := theme.Get(v)
			if err != nil {
				return notice.NewNoticeCmd(state.Deps(), "Invalid theme", err.Error(), notice.NOTICE_ERROR)
			}
			state.Settings.ThemeName = next.Name
			state.Settings.Theme = next
			ui.SaveSettings(state.Settings)
			return ui.ThemeChangedCmd(next, v)
		}),
	)

	solidBackground := toggleinput.New(state.Deps(), "Solid Background",
		toggleinput.WithValue(cfg.SolidBackground),
		toggleinput.WithFocus(),
		toggleinput.WithLabels("Enabled", "Disabled"),
		toggleinput.WithOnSubmit(func(v bool) tea.Cmd {
			state.Settings.SolidBackground = v
			ui.SaveSettings(state.Settings)
			return nil
		}),
	)

	maxRow := textinput.New(state.Deps(), "Enter Max Rows",
		textinput.WithValue(strconv.FormatUint(uint64(cfg.MaxRow), 10)),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if n <= 0 {
				return errPositiveRows
			}
			if n > math.MaxUint32 {
				return errMaxRows
			}
			return nil
		}),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				return notice.NewNoticeCmd(state.Deps(), "Invalid Max Rows", "enter a positive number", notice.NOTICE_ERROR)
			}
			state.Settings.MaxRow = uint32(n)
			ui.SaveSettings(state.Settings)
			return nil
		}),
	)

	// A hand-edited settings file may name a level that no longer parses;
	// fall back to the default instead of showing a blank select.
	currentLevel := cfg.LogLevel
	if _, err := logger.ParseLevel(currentLevel); err != nil {
		currentLevel = ui.DefaultLogLevel
	}

	logLevel := selectinput.New(state.Deps(), "Select Log Level",
		selectinput.WithValue(currentLevel),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("Debug", "debug"),
			huh.NewOption("Info", "info"),
			huh.NewOption("Warning", "warn"),
			huh.NewOption("Error", "error"),
		),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			level, err := logger.ParseLevel(v)
			if err != nil {
				return notice.NewNoticeCmd(state.Deps(), "Invalid Log Level", err.Error(), notice.NOTICE_ERROR)
			}
			state.Settings.LogLevel = v
			ui.SaveSettings(state.Settings)
			state.Log.SetLevel(level)
			return nil
		}),
	)

	fields := []inspector.Field{
		{Name: "Theme", Description: descTheme, Group: groupAppearance, Input: inspector.Adapt(themeSelect), Visible: alwaysVisible, Format: inspector.FormatEmptyStringAuto},
		{Name: "Solid Background", Description: descSolidBackground, Group: groupAppearance, Input: inspector.Adapt(solidBackground), Visible: alwaysVisible, Format: inspector.FormatBool},
		{Name: "Max Rows", Description: descMaxRow, Group: groupAppearance, Input: inspector.Adapt(maxRow), Visible: alwaysVisible, Format: inspector.FormatInt},
		{Name: "Log Level", Description: descLogLevel, Group: groupAppearance, Input: inspector.Adapt(logLevel), Visible: alwaysVisible, Format: formatLogLevel},
	}

	return &Model{
		state:     state,
		name:      name,
		id:        ui.NewComponentID(),
		inspector: inspector.New(state.Deps(), "ui settings", fields),
	}
}

func alwaysVisible() bool { return true }

// formatLogLevel capitalizes the stored level name for display
// ("info" → "Info", "warn" → "Warn").
func formatLogLevel(v any) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
