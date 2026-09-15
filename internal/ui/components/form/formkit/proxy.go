package formkit

import (
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// ProxyInputs holds the inputs composing the shared Proxy & Auth section.
type ProxyInputs struct {
	Type       input.Input[string]
	Port       input.Input[string]
	Auth       input.Input[string]
	Username   input.Input[string]
	Password   input.Input[string]
	PrivateKey input.Input[string]
}

// ProxyVisibility provides the field-visibility predicates derived from the
// current proxy/auth selection.
type ProxyVisibility struct {
	Proxy    func() bool
	Auth     func() bool
	Password func() bool
	Key      func() bool
}

// BuildProxy assembles the full Proxy & Auth section shared by the vaydns,
// dnstt and slipstream forms: proxy type with automatic port switching
// (SOCKS -> 1080, SSH -> 22), auth method and its dependent fields.
//
// The known_hosts_file config field is intentionally not edited here.
// TODO: expose a known_hosts_file field in the proxy forms.
func BuildProxy[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	refresh func() tea.Cmd,
	getType func(C) dns.ResolverProxyType,
	setType func(*C, dns.ResolverProxyType),
	getPort func(C) uint16,
	setPort func(*C, uint16),
	getAuth func(C) dns.AuthMethod,
	setAuth func(*C, dns.AuthMethod),
	getUser func(C) string,
	setUser func(*C, string),
	getPass func(C) string,
	setPass func(*C, string),
	getPriv func(C) string,
	setPriv func(*C, string),
) (*ProxyInputs, *ProxyVisibility) {
	port := textinput.New(
		l, "Enter proxy port",
		textinput.WithValue(strconv.Itoa(int(getPort(*cfg)))),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.ParseUint(strings.TrimSpace(v), 10, 16)
			if err != nil {
				return err
			}
			tmp := *cfg
			setPort(&tmp, uint16(n))
			if e, ok := tmp.Validate()["proxy_port"]; ok {
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

	proxyType := selectinput.New(
		l, "Select proxy type",
		selectinput.WithValue(string(getType(*cfg))),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("SOCKS", "socks"),
			huh.NewOption("SSH", "ssh"),
		),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			setType(&tmp, dns.ResolverProxyType(v))
			if e, ok := tmp.Validate()["proxy_type"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			pt := dns.ResolverProxyType(v)
			if pt == getType(*cfg) {
				return nil
			}
			setType(cfg, pt)
			if pt == dns.ResolverProxySOCKS {
				setPort(cfg, 1080)
				port.SetValue("1080")
			} else {
				setPort(cfg, 22)
				port.SetValue("22")
			}
			return refresh()
		}),
	)

	authMethod := selectinput.New(
		l, "Select authentication method",
		selectinput.WithValue(string(getAuth(*cfg))),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("None", "none"),
			huh.NewOption("Password", "password"),
			huh.NewOption("Key", "key"),
		),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			setAuth(&tmp, dns.AuthMethod(v))
			if e, ok := tmp.Validate()["auth_method"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			setAuth(cfg, dns.AuthMethod(v))
			return refresh()
		}),
	)

	username := textinput.New(
		l, "Enter username",
		textinput.WithValue(getUser(*cfg)),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			tmp := *cfg
			setUser(&tmp, v)
			if e, ok := tmp.Validate()["username"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			setUser(cfg, v)
			return nil
		}),
	)

	password := textinput.New(
		l, "Enter password",
		textinput.WithValue(getPass(*cfg)),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			tmp := *cfg
			setPass(&tmp, v)
			if e, ok := tmp.Validate()["password"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			setPass(cfg, v)
			return nil
		}),
	)

	privateKey := textarea.New(
		l, "Enter private key",
		textarea.WithValue(getPriv(*cfg)),
		textarea.WithFocus(),
		textarea.WithHeight(6),
		textarea.WithPlaceholder("-----BEGIN OPENSSH PRIVATE KEY----- ..."),
		textarea.WithValidation(func(v string) error {
			tmp := *cfg
			setPriv(&tmp, v)
			if e, ok := tmp.Validate()["private_key"]; ok {
				return e
			}
			return nil
		}),
		textarea.WithOnSubmit(func(v string) tea.Cmd {
			setPriv(cfg, v)
			return nil
		}),
	)

	return &ProxyInputs{
		Type:       proxyType,
		Port:       port,
		Auth:       authMethod,
		Username:   username,
		Password:   password,
		PrivateKey: privateKey,
	}, &ProxyVisibility{
		Proxy: func() bool { return getType(*cfg) != "" },
		Auth:  func() bool { return getAuth(*cfg) != dns.AuthNone },
		Password: func() bool {
			return getAuth(*cfg) == dns.AuthPassword
		},
		Key: func() bool { return getAuth(*cfg) == dns.AuthKey },
	}
}
