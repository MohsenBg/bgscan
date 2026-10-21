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
	descAnimation       = "Toggles every TUI animation at once, dialog reveals, tab underlines etc..."
	descMaxRow          = "Maximum number of rows kept for display in result tables."
	descLogLevel        = "Minimum severity written to log files. Debug logs every probe error with the full target IP."
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

func New(session *ui.Session, name string) *Model {
	cfg := session.Env().Settings
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

	themeSelect := selectinput.New(session.Env(), "Select Color Theme",
		selectinput.WithValue(current),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(themeOptions...),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			cmd, err := session.SetTheme(v)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Invalid theme", err.Error(), notice.NOTICE_ERROR)
			}
			return cmd
		}),
	)

	solidBackground := toggleinput.New(session.Env(), "Solid Background",
		toggleinput.WithValue(cfg.SolidBackground),
		toggleinput.WithFocus(),
		toggleinput.WithLabels("Enabled", "Disabled"),
		toggleinput.WithOnSubmit(func(v bool) tea.Cmd {
			session.Env().Settings.SolidBackground = v
			err := ui.SaveSettings(session.Env().Settings)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Failed to save settings", err.Error(), notice.NOTICE_ERROR)
			}
			return nil
		}),
	)

	animation := toggleinput.New(session.Env(), "Animation",
		toggleinput.WithValue(cfg.Animated),
		toggleinput.WithFocus(),
		toggleinput.WithLabels("Enabled", "Disabled"),
		toggleinput.WithOnSubmit(func(v bool) tea.Cmd {
			session.Env().Settings.Animated = v
			err := ui.SaveSettings(session.Env().Settings)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Failed to save settings", err.Error(), notice.NOTICE_ERROR)
			}
			return nil
		}),
	)

	maxRow := textinput.New(session.Env(), "Enter Max Rows",
		textinput.WithValue(strconv.FormatUint(uint64(cfg.MaxRow), 10)),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			if n <= 0 {
				return errPositiveRows
			}
			if n > int64(math.MaxUint32) {
				return errMaxRows
			}
			return nil
		}),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, err := strconv.ParseUint(v, 10, 32)
			if err != nil || n == 0 {
				return notice.NewNoticeCmd(session.Env(), "Invalid Max Rows", "enter a positive number", notice.NOTICE_ERROR)
			}
			session.Env().Settings.MaxRow = uint32(n)
			err = ui.SaveSettings(session.Env().Settings)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Failed to save settings", err.Error(), notice.NOTICE_ERROR)
			}
			return nil
		}),
	)

	// A hand-edited settings file may name a level that no longer parses;
	// fall back to the default instead of showing a blank select.
	currentLevel := cfg.LogLevel
	if _, err := logger.ParseLevel(currentLevel); err != nil {
		currentLevel = ui.DefaultLogLevel
	}

	logLevel := selectinput.New(session.Env(), "Select Log Level",
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
				return notice.NewNoticeCmd(session.Env(), "Invalid Log Level", err.Error(), notice.NOTICE_ERROR)
			}
			session.Env().Settings.LogLevel = v
			err = ui.SaveSettings(session.Env().Settings)
			if err != nil {
				return notice.NewNoticeCmd(session.Env(), "Failed to save settings", err.Error(), notice.NOTICE_ERROR)
			}

			uiLog := session.Env().Log
			uiLog.SetLevel(level)
			coreLog := session.CoreLog()
			coreLog.SetLevel(level)
			return nil
		}),
	)

	fields := []inspector.Field{
		{Name: "Theme", Description: descTheme, Group: groupAppearance, Input: inspector.Adapt(themeSelect), Visible: alwaysVisible, Format: inspector.FormatEmptyStringAuto},
		{Name: "Solid Background", Description: descSolidBackground, Group: groupAppearance, Input: inspector.Adapt(solidBackground), Visible: alwaysVisible, Format: inspector.FormatBool},
		{Name: "Animation", Description: descAnimation, Group: groupAppearance, Input: inspector.Adapt(animation), Visible: alwaysVisible, Format: inspector.FormatBool},
		{Name: "Max Rows", Description: descMaxRow, Group: groupAppearance, Input: inspector.Adapt(maxRow), Visible: alwaysVisible, Format: inspector.FormatInt},
		{Name: "Log Level", Description: descLogLevel, Group: groupAppearance, Input: inspector.Adapt(logLevel), Visible: alwaysVisible, Format: formatLogLevel},
	}

	return &Model{
		session:   session,
		name:      name,
		id:        ui.NewComponentID(),
		inspector: inspector.New(session.Env(), "ui settings", fields),
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
