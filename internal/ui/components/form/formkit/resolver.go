package formkit

import (
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// ResolverTypePort builds the coupled resolver transport + port inputs shared
// by the vaydns, dnstt and thefeed forms: selecting DOT switches the port to
// 853, anything else resets it to 53.
func ResolverTypePort[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	refresh func() tea.Cmd,
	getType func(C) dns.ResolverType,
	setType func(*C, dns.ResolverType),
	getPort func(C) uint16,
	setPort func(*C, uint16),
) (input.Input[string], input.Input[string]) {
	port := textinput.New(
		l, "Enter resolver port",
		textinput.WithValue(strconv.Itoa(int(getPort(*cfg)))),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.ParseUint(strings.TrimSpace(v), 10, 16)
			if err != nil {
				return err
			}
			tmp := *cfg
			setPort(&tmp, uint16(n))
			if e, ok := tmp.Validate()["resolver_port"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 16)
			setPort(cfg, uint16(n))
			return nil
		}),
	)

	resType := selectinput.New(
		l, "Select resolver type",
		selectinput.WithValue(string(getType(*cfg))),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("UDP", "udp"),
			huh.NewOption("TCP", "tcp"),
			huh.NewOption("DOT", "dot"),
		),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			setType(&tmp, dns.ResolverType(v))
			if e, ok := tmp.Validate()["resolver_type"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			rt := dns.ResolverType(v)
			if rt == getType(*cfg) {
				return nil
			}
			setType(cfg, rt)
			if rt == dns.ResolverTypeDOT {
				setPort(cfg, 853)
				port.SetValue("853")
			} else {
				setPort(cfg, 53)
				port.SetValue("53")
			}
			return refresh()
		}),
	)

	return resType, port
}
