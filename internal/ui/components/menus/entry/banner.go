package entry

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// bannerHeight is the number of rows reserved for the banner at the
// top of the body region.
const bannerHeight = 8

var bannerStr = strings.TrimSpace(`
██████╗  ██████╗      ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██╔════╝      ██╔════╝██╔════╝██╔══██╗████╗  ██║
██████╔╝██║  ███╗     ███████╗██║     ███████║██╔██╗ ██║
██╔══██╗██║   ██║     ╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝╚██████╔╝     ███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝  ╚═════╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
`)

// bannerStyle centers the banner text in the given size and colors it
// with the theme success color.
func bannerStyle(env *ui.Env, width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(width).
		Height(height).
		Foreground(env.Theme.Success).
		Bold(true)
}

func menuStyle(width, hight int) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Width(width).Height(hight)
}
