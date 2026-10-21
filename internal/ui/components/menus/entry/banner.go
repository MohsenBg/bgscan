package entry

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// bannerHeight is the row budget the banner block
const bannerHeight = 8

var bannerStr = strings.TrimSpace(`
██████╗  ██████╗      ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██╔════╝      ██╔════╝██╔════╝██╔══██╗████╗  ██║
██████╔╝██║  ███╗     ███████╗██║     ███████║██╔██╗ ██║
██╔══██╗██║   ██║     ╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝╚██████╔╝     ███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝  ╚═════╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
`)

// bannerStyle paints the banner into the top of the body region
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
