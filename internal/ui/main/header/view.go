package header

import "strings"

var bannerStr = strings.TrimSpace(`
██████╗  ██████╗      ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██╔════╝      ██╔════╝██╔════╝██╔══██╗████╗  ██║
██████╔╝██║  ███╗     ███████╗██║     ███████║██╔██╗ ██║
██╔══██╗██║   ██║     ╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝╚██████╔╝     ███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝  ╚═════╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
`)

func (m Model) View() string {
	return bannerStyle(m.layout.Header.Width, m.layout.Header.Height).Render(bannerStr)
}
