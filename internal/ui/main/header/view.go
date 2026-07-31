package header

import "strings"

// getBanner returns the ASCII banner string.
func getBanner() string {
	banner := `
██████╗  ██████╗      ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██╔════╝      ██╔════╝██╔════╝██╔══██╗████╗  ██║
██████╔╝██║  ███╗     ███████╗██║     ███████║██╔██╗ ██║
██╔══██╗██║   ██║     ╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝╚██████╔╝     ███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝  ╚═════╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
`
	banner = strings.Replace(banner, "\n", "", 1)
	return banner
}

// View renders the banner within the header region.
func (m Model) View() string {
	banner := bannerStyle(
		m.layout.Header.Width,
		m.layout.Header.Height,
	).Render(getBanner())
	return banner
}
