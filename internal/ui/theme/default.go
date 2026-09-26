package theme

import "charm.land/lipgloss/v2"

var BGScanDark = Theme{
	Name: "bgscan-dark",

	Primary:      lipgloss.Color("#D75FD7"),
	Secondary:    lipgloss.Color("#8A8A8A"),
	Border:       lipgloss.Color("#585858"),
	BorderActive: lipgloss.Color("#5F5FD7"),
	Text:         lipgloss.Color("#D0D0D0"),
	Muted:        lipgloss.Color("#626262"),
	Timestamp:    lipgloss.Color("#CE9178"),

	Info:    lipgloss.Color("#00AFFF"),
	Error:   lipgloss.Color("#FF0000"),
	Success: lipgloss.Color("#00D787"),

	Orange: lipgloss.Color("#FF8700"),
	Yellow: lipgloss.Color("#FFD700"),
	Purple: lipgloss.Color("#5F5FD7"),

	Selected:   lipgloss.Color("#303030"),
	Warning:    lipgloss.Color("#FFD700"),
	Background: lipgloss.Color("#000000"),

	ProgressStart: lipgloss.Color("#A78BFA"),
	ProgressEnd:   lipgloss.Color("#7DD3FC"),
}

var BGScanLight = Theme{
	Name: "bgscan-light",

	Primary:      lipgloss.Color("#D75FD7"),
	Secondary:    lipgloss.Color("#A8A8A8"),
	Border:       lipgloss.Color("#808080"),
	BorderActive: lipgloss.Color("#5F5FD7"),
	Text:         lipgloss.Color("#1C1C1C"),
	Muted:        lipgloss.Color("#949494"),
	Timestamp:    lipgloss.Color("#1E3A8A"),

	Info:    lipgloss.Color("#005FFF"),
	Error:   lipgloss.Color("#D70000"),
	Success: lipgloss.Color("#008700"),

	Orange: lipgloss.Color("#FF8700"),
	Yellow: lipgloss.Color("#FFD700"),
	Purple: lipgloss.Color("#5F5FD7"),

	Selected:   lipgloss.Color("#E5E5E5"),
	Warning:    lipgloss.Color("#AF8700"),
	Background: lipgloss.Color("#FFFFFF"),

	ProgressStart: lipgloss.Color("#6D28D9"),
	ProgressEnd:   lipgloss.Color("#0369A1"),
}

var TokyoNight = Theme{
	Name: "tokyo-night",

	Primary:      lipgloss.Color("#7AA2F7"),
	Secondary:    lipgloss.Color("#565F89"),
	Border:       lipgloss.Color("#3B4261"),
	BorderActive: lipgloss.Color("#7AA2F7"),
	Text:         lipgloss.Color("#C0CAF5"),
	Muted:        lipgloss.Color("#565F89"),
	Timestamp:    lipgloss.Color("#BB9AF7"),

	Info:    lipgloss.Color("#7DCFFF"),
	Error:   lipgloss.Color("#F7768E"),
	Success: lipgloss.Color("#9ECE6A"),

	Orange: lipgloss.Color("#FF9E64"),
	Yellow: lipgloss.Color("#E0AF68"),
	Purple: lipgloss.Color("#BB9AF7"),

	Selected:   lipgloss.Color("#292E42"),
	Warning:    lipgloss.Color("#E0AF68"),
	Background: lipgloss.Color("#1A1B26"),

	ProgressStart: lipgloss.Color("#BB9AF7"),
	ProgressEnd:   lipgloss.Color("#7AA2F7"),
}

// GruvboxDark is the popular Gruvbox dark palette.
var GruvboxDark = Theme{
	Name:          "gruvbox-dark",
	Primary:       lipgloss.Color("#D3869B"),
	Secondary:     lipgloss.Color("#A89984"),
	Border:        lipgloss.Color("#3C3836"),
	BorderActive:  lipgloss.Color("#D3869B"),
	Text:          lipgloss.Color("#EBDBB2"),
	Muted:         lipgloss.Color("#928374"),
	Timestamp:     lipgloss.Color("#FE8019"),
	Info:          lipgloss.Color("#83A598"),
	Error:         lipgloss.Color("#FB4934"),
	Success:       lipgloss.Color("#B8BB26"),
	Orange:        lipgloss.Color("#FE8019"),
	Yellow:        lipgloss.Color("#FABD2F"),
	Purple:        lipgloss.Color("#D3869B"),
	Selected:      lipgloss.Color("#504945"),
	Warning:       lipgloss.Color("#FABD2F"),
	Background:    lipgloss.Color("#282828"),
	ProgressStart: lipgloss.Color("#83A598"),
	ProgressEnd:   lipgloss.Color("#B8BB26"),
}

// GruvboxLight is the popular Gruvbox light palette.
var GruvboxLight = Theme{
	Name:          "gruvbox-light",
	Primary:       lipgloss.Color("#8F3F71"),
	Secondary:     lipgloss.Color("#7C6F64"),
	Border:        lipgloss.Color("#D5C4A1"),
	BorderActive:  lipgloss.Color("#8F3F71"),
	Text:          lipgloss.Color("#3C3836"),
	Muted:         lipgloss.Color("#928374"),
	Timestamp:     lipgloss.Color("#AF3A03"),
	Info:          lipgloss.Color("#076678"),
	Error:         lipgloss.Color("#9D0006"),
	Success:       lipgloss.Color("#79740E"),
	Orange:        lipgloss.Color("#AF3A03"),
	Yellow:        lipgloss.Color("#B57614"),
	Purple:        lipgloss.Color("#8F3F71"),
	Selected:      lipgloss.Color("#EBDBB2"),
	Warning:       lipgloss.Color("#B57614"),
	Background:    lipgloss.Color("#FBF1C7"),
	ProgressStart: lipgloss.Color("#076678"),
	ProgressEnd:   lipgloss.Color("#79740E"),
}

// Dracula is the classic Dracula palette.
var Dracula = Theme{
	Name:          "dracula",
	Primary:       lipgloss.Color("#BD93F9"),
	Secondary:     lipgloss.Color("#6272A4"),
	Border:        lipgloss.Color("#44475A"),
	BorderActive:  lipgloss.Color("#BD93F9"),
	Text:          lipgloss.Color("#F8F8F2"),
	Muted:         lipgloss.Color("#6272A4"),
	Timestamp:     lipgloss.Color("#FFB86C"),
	Info:          lipgloss.Color("#8BE9FD"),
	Error:         lipgloss.Color("#FF5555"),
	Success:       lipgloss.Color("#50FA7B"),
	Orange:        lipgloss.Color("#FFB86C"),
	Yellow:        lipgloss.Color("#F1FA8C"),
	Purple:        lipgloss.Color("#BD93F9"),
	Selected:      lipgloss.Color("#44475A"),
	Warning:       lipgloss.Color("#F1FA8C"),
	Background:    lipgloss.Color("#282A36"),
	ProgressStart: lipgloss.Color("#BD93F9"),
	ProgressEnd:   lipgloss.Color("#8BE9FD"),
}

// SolarizedDark is the classic Solarized dark palette.
var SolarizedDark = Theme{
	Name:          "solarized-dark",
	Primary:       lipgloss.Color("#268BD2"),
	Secondary:     lipgloss.Color("#839496"),
	Border:        lipgloss.Color("#073642"),
	BorderActive:  lipgloss.Color("#268BD2"),
	Text:          lipgloss.Color("#EEE8D5"),
	Muted:         lipgloss.Color("#586E75"),
	Timestamp:     lipgloss.Color("#CB4B16"),
	Info:          lipgloss.Color("#2AA198"),
	Error:         lipgloss.Color("#DC322F"),
	Success:       lipgloss.Color("#859900"),
	Orange:        lipgloss.Color("#CB4B16"),
	Yellow:        lipgloss.Color("#B58900"),
	Purple:        lipgloss.Color("#6C71C4"),
	Selected:      lipgloss.Color("#073642"),
	Warning:       lipgloss.Color("#B58900"),
	Background:    lipgloss.Color("#002B36"),
	ProgressStart: lipgloss.Color("#268BD2"),
	ProgressEnd:   lipgloss.Color("#2AA198"),
}

// SolarizedLight is the classic Solarized light palette.
var SolarizedLight = Theme{
	Name:          "solarized-light",
	Primary:       lipgloss.Color("#268BD2"),
	Secondary:     lipgloss.Color("#657B83"),
	Border:        lipgloss.Color("#EEE8D5"),
	BorderActive:  lipgloss.Color("#268BD2"),
	Text:          lipgloss.Color("#073642"),
	Muted:         lipgloss.Color("#93A1A1"),
	Timestamp:     lipgloss.Color("#CB4B16"),
	Info:          lipgloss.Color("#2AA198"),
	Error:         lipgloss.Color("#DC322F"),
	Success:       lipgloss.Color("#859900"),
	Orange:        lipgloss.Color("#CB4B16"),
	Yellow:        lipgloss.Color("#B58900"),
	Purple:        lipgloss.Color("#6C71C4"),
	Selected:      lipgloss.Color("#EEE8D5"),
	Warning:       lipgloss.Color("#B58900"),
	Background:    lipgloss.Color("#FDF6E3"),
	ProgressStart: lipgloss.Color("#268BD2"),
	ProgressEnd:   lipgloss.Color("#2AA198"),
}

// OneDark is the popular Atom One Dark palette.
var OneDark = Theme{
	Name:          "one-dark",
	Primary:       lipgloss.Color("#61AFEF"),
	Secondary:     lipgloss.Color("#ABB2BF"),
	Border:        lipgloss.Color("#3E4451"),
	BorderActive:  lipgloss.Color("#61AFEF"),
	Text:          lipgloss.Color("#ABB2BF"),
	Muted:         lipgloss.Color("#5C6370"),
	Timestamp:     lipgloss.Color("#D19A66"),
	Info:          lipgloss.Color("#56B6C2"),
	Error:         lipgloss.Color("#E06C75"),
	Success:       lipgloss.Color("#98C379"),
	Orange:        lipgloss.Color("#D19A66"),
	Yellow:        lipgloss.Color("#E5C07B"),
	Purple:        lipgloss.Color("#C678DD"),
	Selected:      lipgloss.Color("#3E4451"),
	Warning:       lipgloss.Color("#E5C07B"),
	Background:    lipgloss.Color("#282C34"),
	ProgressStart: lipgloss.Color("#61AFEF"),
	ProgressEnd:   lipgloss.Color("#56B6C2"),
}

// OneLight is the popular Atom One Light palette.
var OneLight = Theme{
	Name:          "one-light",
	Primary:       lipgloss.Color("#4078F2"),
	Secondary:     lipgloss.Color("#A0A1A7"),
	Border:        lipgloss.Color("#D0D0D0"),
	BorderActive:  lipgloss.Color("#4078F2"),
	Text:          lipgloss.Color("#383A42"),
	Muted:         lipgloss.Color("#A0A1A7"),
	Timestamp:     lipgloss.Color("#C18401"),
	Info:          lipgloss.Color("#0184BC"),
	Error:         lipgloss.Color("#E45649"),
	Success:       lipgloss.Color("#50A14F"),
	Orange:        lipgloss.Color("#C18401"),
	Yellow:        lipgloss.Color("#C18401"),
	Purple:        lipgloss.Color("#A626A4"),
	Selected:      lipgloss.Color("#E5E5E6"),
	Warning:       lipgloss.Color("#C18401"),
	Background:    lipgloss.Color("#FAFAFA"),
	ProgressStart: lipgloss.Color("#4078F2"),
	ProgressEnd:   lipgloss.Color("#0184BC"),
}

// Monokai is the classic Monokai palette.
var Monokai = Theme{
	Name:          "monokai",
	Primary:       lipgloss.Color("#F92672"),
	Secondary:     lipgloss.Color("#75715E"),
	Border:        lipgloss.Color("#3E3D32"),
	BorderActive:  lipgloss.Color("#F92672"),
	Text:          lipgloss.Color("#F8F8F2"),
	Muted:         lipgloss.Color("#75715E"),
	Timestamp:     lipgloss.Color("#FD971F"),
	Info:          lipgloss.Color("#66D9EF"),
	Error:         lipgloss.Color("#F92672"),
	Success:       lipgloss.Color("#A6E22E"),
	Orange:        lipgloss.Color("#FD971F"),
	Yellow:        lipgloss.Color("#E6DB74"),
	Purple:        lipgloss.Color("#AE81FF"),
	Selected:      lipgloss.Color("#49483E"),
	Warning:       lipgloss.Color("#E6DB74"),
	Background:    lipgloss.Color("#272822"),
	ProgressStart: lipgloss.Color("#66D9EF"),
	ProgressEnd:   lipgloss.Color("#A6E22E"),
}

// RosePineDark is the popular Rosé Pine palette.
var RosePineDark = Theme{
	Name:          "rose-pine",
	Primary:       lipgloss.Color("#C4A7E7"),
	Secondary:     lipgloss.Color("#908CAA"),
	Border:        lipgloss.Color("#26233A"),
	BorderActive:  lipgloss.Color("#C4A7E7"),
	Text:          lipgloss.Color("#E0DEF4"),
	Muted:         lipgloss.Color("#6E6A86"),
	Timestamp:     lipgloss.Color("#EA9A97"),
	Info:          lipgloss.Color("#9CCFD8"),
	Error:         lipgloss.Color("#EB6F92"),
	Success:       lipgloss.Color("#31748F"),
	Orange:        lipgloss.Color("#EA9A97"),
	Yellow:        lipgloss.Color("#F6C177"),
	Purple:        lipgloss.Color("#C4A7E7"),
	Selected:      lipgloss.Color("#403D52"),
	Warning:       lipgloss.Color("#F6C177"),
	Background:    lipgloss.Color("#191724"),
	ProgressStart: lipgloss.Color("#C4A7E7"),
	ProgressEnd:   lipgloss.Color("#9CCFD8"),
}

// RosePineDawn is the popular Rosé Pine Dawn (light) palette.
var RosePineDawn = Theme{
	Name:          "rose-pine-dawn",
	Primary:       lipgloss.Color("#907AA9"),
	Secondary:     lipgloss.Color("#797593"),
	Border:        lipgloss.Color("#DFDAD9"),
	BorderActive:  lipgloss.Color("#907AA9"),
	Text:          lipgloss.Color("#575279"),
	Muted:         lipgloss.Color("#9893A5"),
	Timestamp:     lipgloss.Color("#D7827E"),
	Info:          lipgloss.Color("#56949F"),
	Error:         lipgloss.Color("#B4637A"),
	Success:       lipgloss.Color("#286983"),
	Orange:        lipgloss.Color("#D7827E"),
	Yellow:        lipgloss.Color("#EA9D34"),
	Purple:        lipgloss.Color("#907AA9"),
	Selected:      lipgloss.Color("#F2E9E1"),
	Warning:       lipgloss.Color("#EA9D34"),
	Background:    lipgloss.Color("#FAF4ED"),
	ProgressStart: lipgloss.Color("#907AA9"),
	ProgressEnd:   lipgloss.Color("#56949F"),
}

// EverforestDark is the popular Everforest dark palette.
var EverforestDark = Theme{
	Name:          "everforest-dark",
	Primary:       lipgloss.Color("#A7C080"),
	Secondary:     lipgloss.Color("#859289"),
	Border:        lipgloss.Color("#4F5B58"),
	BorderActive:  lipgloss.Color("#A7C080"),
	Text:          lipgloss.Color("#D3C6AA"),
	Muted:         lipgloss.Color("#7A8478"),
	Timestamp:     lipgloss.Color("#E69875"),
	Info:          lipgloss.Color("#7FBBB3"),
	Error:         lipgloss.Color("#E67E80"),
	Success:       lipgloss.Color("#A7C080"),
	Orange:        lipgloss.Color("#E69875"),
	Yellow:        lipgloss.Color("#DBBC7F"),
	Purple:        lipgloss.Color("#D699B6"),
	Selected:      lipgloss.Color("#3D484D"),
	Warning:       lipgloss.Color("#DBBC7F"),
	Background:    lipgloss.Color("#2D353B"),
	ProgressStart: lipgloss.Color("#7FBBB3"),
	ProgressEnd:   lipgloss.Color("#A7C080"),
}

// EverforestLight is the popular Everforest light palette.
var EverforestLight = Theme{
	Name:          "everforest-light",
	Primary:       lipgloss.Color("#8DA101"),
	Secondary:     lipgloss.Color("#939F91"),
	Border:        lipgloss.Color("#E0DCC7"),
	BorderActive:  lipgloss.Color("#8DA101"),
	Text:          lipgloss.Color("#5C6A72"),
	Muted:         lipgloss.Color("#A6B0A0"),
	Timestamp:     lipgloss.Color("#F57D26"),
	Info:          lipgloss.Color("#3A94C5"),
	Error:         lipgloss.Color("#F85552"),
	Success:       lipgloss.Color("#8DA101"),
	Orange:        lipgloss.Color("#F57D26"),
	Yellow:        lipgloss.Color("#DFA000"),
	Purple:        lipgloss.Color("#DF69BA"),
	Selected:      lipgloss.Color("#EFEBD4"),
	Warning:       lipgloss.Color("#DFA000"),
	Background:    lipgloss.Color("#FDF6E3"),
	ProgressStart: lipgloss.Color("#3A94C5"),
	ProgressEnd:   lipgloss.Color("#8DA101"),
}

// AyuDark is the popular Ayu Dark palette.
var AyuDark = Theme{
	Name:          "ayu-dark",
	Primary:       lipgloss.Color("#E6B450"),
	Secondary:     lipgloss.Color("#565B66"),
	Border:        lipgloss.Color("#2D3640"),
	BorderActive:  lipgloss.Color("#E6B450"),
	Text:          lipgloss.Color("#B3B1AD"),
	Muted:         lipgloss.Color("#565B66"),
	Timestamp:     lipgloss.Color("#FF8F40"),
	Info:          lipgloss.Color("#39BAE6"),
	Error:         lipgloss.Color("#FF3333"),
	Success:       lipgloss.Color("#C2D94C"),
	Orange:        lipgloss.Color("#FF8F40"),
	Yellow:        lipgloss.Color("#E6B450"),
	Purple:        lipgloss.Color("#D2A6FF"),
	Selected:      lipgloss.Color("#273747"),
	Warning:       lipgloss.Color("#E6B450"),
	Background:    lipgloss.Color("#0A0E14"),
	ProgressStart: lipgloss.Color("#39BAE6"),
	ProgressEnd:   lipgloss.Color("#C2D94C"),
}

// AyuLight is the popular Ayu Light palette.
var AyuLight = Theme{
	Name:          "ayu-light",
	Primary:       lipgloss.Color("#FF9940"),
	Secondary:     lipgloss.Color("#828C99"),
	Border:        lipgloss.Color("#E7E8E9"),
	BorderActive:  lipgloss.Color("#FF9940"),
	Text:          lipgloss.Color("#5C6166"),
	Muted:         lipgloss.Color("#ABB0B6"),
	Timestamp:     lipgloss.Color("#FA8D3E"),
	Info:          lipgloss.Color("#399EE6"),
	Error:         lipgloss.Color("#F51818"),
	Success:       lipgloss.Color("#86B300"),
	Orange:        lipgloss.Color("#FA8D3E"),
	Yellow:        lipgloss.Color("#F2AE49"),
	Purple:        lipgloss.Color("#A37ACC"),
	Selected:      lipgloss.Color("#F0EEE4"),
	Warning:       lipgloss.Color("#F2AE49"),
	Background:    lipgloss.Color("#FAFAFA"),
	ProgressStart: lipgloss.Color("#399EE6"),
	ProgressEnd:   lipgloss.Color("#86B300"),
}

// Kanagawa is the popular Kanagawa palette.
var Kanagawa = Theme{
	Name:          "kanagawa",
	Primary:       lipgloss.Color("#7E9CD8"),
	Secondary:     lipgloss.Color("#727169"),
	Border:        lipgloss.Color("#2A2A37"),
	BorderActive:  lipgloss.Color("#7E9CD8"),
	Text:          lipgloss.Color("#DCD7BA"),
	Muted:         lipgloss.Color("#727169"),
	Timestamp:     lipgloss.Color("#FFA066"),
	Info:          lipgloss.Color("#7FB4CA"),
	Error:         lipgloss.Color("#E82424"),
	Success:       lipgloss.Color("#98BB6C"),
	Orange:        lipgloss.Color("#FFA066"),
	Yellow:        lipgloss.Color("#E6C384"),
	Purple:        lipgloss.Color("#957FB8"),
	Selected:      lipgloss.Color("#363646"),
	Warning:       lipgloss.Color("#E6C384"),
	Background:    lipgloss.Color("#1F1F28"),
	ProgressStart: lipgloss.Color("#7E9CD8"),
	ProgressEnd:   lipgloss.Color("#7FB4CA"),
}

// GithubDark is GitHub's dark UI palette.
var GithubDark = Theme{
	Name:          "github-dark",
	Primary:       lipgloss.Color("#58A6FF"),
	Secondary:     lipgloss.Color("#8B949E"),
	Border:        lipgloss.Color("#30363D"),
	BorderActive:  lipgloss.Color("#58A6FF"),
	Text:          lipgloss.Color("#C9D1D9"),
	Muted:         lipgloss.Color("#8B949E"),
	Timestamp:     lipgloss.Color("#DB6D28"),
	Info:          lipgloss.Color("#79C0FF"),
	Error:         lipgloss.Color("#F85149"),
	Success:       lipgloss.Color("#3FB950"),
	Orange:        lipgloss.Color("#DB6D28"),
	Yellow:        lipgloss.Color("#D29922"),
	Purple:        lipgloss.Color("#BC8CFF"),
	Selected:      lipgloss.Color("#21262D"),
	Warning:       lipgloss.Color("#D29922"),
	Background:    lipgloss.Color("#0D1117"),
	ProgressStart: lipgloss.Color("#58A6FF"),
	ProgressEnd:   lipgloss.Color("#3FB950"),
}

// GithubLight is GitHub's light UI palette.
var GithubLight = Theme{
	Name:          "github-light",
	Primary:       lipgloss.Color("#0969DA"),
	Secondary:     lipgloss.Color("#57606A"),
	Border:        lipgloss.Color("#D0D7DE"),
	BorderActive:  lipgloss.Color("#0969DA"),
	Text:          lipgloss.Color("#1F2328"),
	Muted:         lipgloss.Color("#6E7781"),
	Timestamp:     lipgloss.Color("#BC4C00"),
	Info:          lipgloss.Color("#218BFF"),
	Error:         lipgloss.Color("#CF222E"),
	Success:       lipgloss.Color("#1A7F37"),
	Orange:        lipgloss.Color("#BC4C00"),
	Yellow:        lipgloss.Color("#9A6700"),
	Purple:        lipgloss.Color("#8250DF"),
	Selected:      lipgloss.Color("#DDF4FF"),
	Warning:       lipgloss.Color("#9A6700"),
	Background:    lipgloss.Color("#FFFFFF"),
	ProgressStart: lipgloss.Color("#0969DA"),
	ProgressEnd:   lipgloss.Color("#1A7F37"),
}

// MaterialDark is the popular Material Theme dark palette.
var MaterialDark = Theme{
	Name:          "material-dark",
	Primary:       lipgloss.Color("#82AAFF"),
	Secondary:     lipgloss.Color("#B2CCD6"),
	Border:        lipgloss.Color("#2E3440"),
	BorderActive:  lipgloss.Color("#82AAFF"),
	Text:          lipgloss.Color("#EEFFFF"),
	Muted:         lipgloss.Color("#546E7A"),
	Timestamp:     lipgloss.Color("#F78C6C"),
	Info:          lipgloss.Color("#89DDFF"),
	Error:         lipgloss.Color("#FF5370"),
	Success:       lipgloss.Color("#C3E88D"),
	Orange:        lipgloss.Color("#F78C6C"),
	Yellow:        lipgloss.Color("#FFCB6B"),
	Purple:        lipgloss.Color("#C792EA"),
	Selected:      lipgloss.Color("#2D3B4E"),
	Warning:       lipgloss.Color("#FFCB6B"),
	Background:    lipgloss.Color("#263238"),
	ProgressStart: lipgloss.Color("#82AAFF"),
	ProgressEnd:   lipgloss.Color("#89DDFF"),
}
