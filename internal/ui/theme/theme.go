package theme

import (
	"image/color"
	"os"
)

type Theme struct {
	Name string `toml:"-"`

	Primary   color.Color
	Secondary color.Color

	Border       color.Color
	BorderActive color.Color

	Text      color.Color
	Muted     color.Color
	Timestamp color.Color

	Info    color.Color
	Error   color.Color
	Success color.Color

	Orange color.Color
	Yellow color.Color
	Purple color.Color

	Selected color.Color
	Warning  color.Color

	Background color.Color

	ProgressStart color.Color
	ProgressEnd   color.Color
}

var (
	registry = map[string]Theme{}
	order    []string
)

func Register(themes ...Theme) {
	for _, t := range themes {
		if t.Name == "" {
			continue
		}

		if _, ok := registry[t.Name]; !ok {
			order = append(order, t.Name)
		}

		registry[t.Name] = t
	}
}

// Init registers the built-in themes in List order.
func Init() {
	Register(
		BGScanDark,
		TokyoNight,
		GruvboxDark,
		OneDark,
		Dracula,
		Monokai,
		SolarizedDark,
		RosePineDark,
		Kanagawa,
		GithubDark,
		MaterialDark,
		EverforestDark,
		AyuDark,

		BGScanLight,
		GruvboxLight,
		OneLight,
		SolarizedLight,
		RosePineDawn,
		EverforestLight,
		AyuLight,
		GithubLight,
	)
}

// List returns registered theme names in cycle order.
func List() []string {
	return append([]string(nil), order...)
}

// Get returns the named theme. An empty name auto-detects from the
// terminal; unknown names return an UnknownThemeError.
func Get(name string) (*Theme, error) {
	if name == "" {
		if terminalLooksDark() {
			name = BGScanDark.Name
		} else {
			name = BGScanLight.Name
		}
	}

	t, ok := registry[name]
	if !ok {
		return nil, &UnknownThemeError{Name: name}
	}
	return &t, nil
}

// UnknownThemeError reports an attempt to activate a theme that is not
// registered.
type UnknownThemeError struct {
	Name string
}

func (e *UnknownThemeError) Error() string {
	return "unknown theme " + e.Name
}

// terminalLooksDark detects terminal background darkness from COLORFGBG.
func terminalLooksDark() bool {
	bg := os.Getenv("COLORFGBG")

	if bg == "" {
		return true
	}

	for i := len(bg) - 1; i >= 0; i-- {
		if bg[i] == ';' {
			bg = bg[i+1:]
			break
		}
	}

	switch bg {
	case "0", "1", "2", "3", "4", "5", "6", "7":
		return true
	default:
		return false
	}
}
