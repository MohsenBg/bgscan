package dnstt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/form"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

const (
	maxFormWidth  = 55
	maxFormHeight = 35
	formPadding   = 2
)

const (
	groupConnection = "Connection"
	groupProxyAuth  = "Proxy & Auth"
)

const (
	descDomain       = "The target domain name for the DNSTT tunnel."
	descPubKey       = "The public key for encryption."
	descResolverType = "The DNS resolver type: UDP, TCP, or DOT."
	descResolverPort = "The DNS resolver port (typically 53)."
	descFingerprint  = "The TLS fingerprint for the resolver."
	descRPS          = "Requests per second (0 = unlimited)."
	descProxyType    = "The proxy type: SOCKS or SSH."
	descProxyPort    = "The proxy port number."
	descAuthMethod   = "The authentication method: none, password, or key."
	descUsername     = "The username for authentication."
	descPassword     = "The password for password authentication."
	descPrivateKey   = "The SSH private key used for authentication."
)

type Model struct {
	id           ui.ComponentID
	layout       *layout.Layout
	state        *ui.AppState
	form         *form.Model
	inspector    *inspector.Model
	cfg          *dns.DNSTTConfig
	name         string
	originalName string
	width        int
	height       int
}

func New(
	l *layout.Layout,
	state *ui.AppState,
	original *dns.DNSTunConfigFile,
) (*Model, error) {
	cfg := dns.DefaultDNSTTConfig()
	name := ""
	originalName := ""

	if original != nil {
		name = original.Name
		originalName = original.Name
		if c, ok := original.Config.(dns.DNSTTConfig); ok {
			cfg = c
		}
	}

	m := &Model{
		id:           ui.NewComponentID(),
		layout:       l,
		state:        state,
		cfg:          &cfg,
		name:         name,
		originalName: originalName,
	}

	m.calculateSize()
	m.buildForm(original)

	return m, nil
}

func (m *Model) calculateSize() {
	w := m.layout.BodyContentWidth() - formPadding
	h := m.layout.BodyContentHeight() - formPadding

	m.width = min(w, maxFormWidth)
	m.height = min(h, maxFormHeight)
}

func (m *Model) buildForm(original *dns.DNSTunConfigFile) {
	m.inspector = m.buildInspector()

	title := "New DNSTT Config"
	if original != nil {
		title = fmt.Sprintf("Edit: %s", m.name)
	}

	m.form = form.New(
		m.layout,
		m.inspector,
		form.WithName(title),
		form.WithWidth(m.width),
		form.WithHeight(m.height),
		form.WithValidation(func(fm *form.Model) error {
			errs := m.cfg.Validate()
			if len(errs) == 0 {
				return nil
			}
			return fmt.Errorf("%s", form.FormatValidationErrors(errs))
		}),
		form.WithSave(confirm.ConfirmCmd(
			m.layout,
			"Save configuration?",
			m.saveConfig,
			true,
		)),
		form.WithCancel(m.cancel),
	)
}

func (m *Model) buildInspector() *inspector.Model {
	cfg := m.cfg
	l := m.layout

	configName := textinput.New(
		l, "Enter config name",
		textinput.WithValue(m.name),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			if err := validation.ValidateFilename(v); err != nil {
				return err
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			m.name = strings.TrimSpace(v)
			return nil
		}),
	)

	domain := textinput.New(
		l, "Enter domain",
		textinput.WithValue(cfg.Domain),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.Domain = v
			errs := tmp.Validate()
			if e, ok := errs["domain"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.Domain = strings.TrimSpace(v)
			return nil
		}),
	)

	pubKey := textarea.New(
		l, "Enter public key",
		textarea.WithValue(cfg.PubKey),
		textarea.WithFocus(),
		textarea.WithNewlines(false),
		textarea.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.PubKey = v
			errs := tmp.Validate()
			if e, ok := errs["pub_key"]; ok {
				return e
			}
			return nil
		}),
		textarea.WithOnSubmit(func(v string) tea.Cmd {
			cfg.PubKey = strings.TrimSpace(v)
			return nil
		}),
	)

	resolverPort := textinput.New(
		l, "Enter resolver port",
		textinput.WithValue(strconv.Itoa(int(cfg.ResolverPort))),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return err
			}
			tmp := *cfg
			tmp.ResolverPort = uint16(n)
			errs := tmp.Validate()
			if e, ok := errs["resolver_port"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, _ := strconv.ParseUint(v, 10, 16)
			cfg.ResolverPort = uint16(n)
			return nil
		}),
	)

	resolverType := selectinput.New(
		l, "Select resolver type",
		selectinput.WithValue(string(cfg.ResolverType)),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("UDP", "udp"),
			huh.NewOption("TCP", "tcp"),
			huh.NewOption("DOT", "dot"),
		),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			rt := dns.ResolverType(v)
			if rt == cfg.ResolverType {
				return nil
			}
			cfg.ResolverType = rt
			if rt == dns.ResolverTypeDOT {
				cfg.ResolverPort = 853
				resolverPort.SetValue("853")
			} else {
				cfg.ResolverPort = 53
				resolverPort.SetValue("53")
			}
			return m.refresh()
		}),
	)

	fingerprintOptions := make([]huh.Option[string], 0, len(config.FingerprintLabels()))
	for _, label := range config.FingerprintLabels() {
		fingerprintOptions = append(fingerprintOptions, huh.NewOption(label, label))
	}

	fingerprint := selectinput.New(
		l, "Select TLS fingerprint",
		selectinput.WithValue(cfg.Fingerprint),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(fingerprintOptions...),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.Fingerprint = v
			errs := tmp.Validate()
			if e, ok := errs["fingerprint"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.Fingerprint = v
			return nil
		}),
	)

	rps := textinput.New(
		l, "Enter requests per second",
		textinput.WithValue(strconv.FormatFloat(cfg.RPS, 'f', -1, 64)),
		textinput.WithFocus(),
		textinput.WithPlaceholder("0 = unlimited"),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			tmp := *cfg
			tmp.RPS = n
			errs := tmp.Validate()
			if e, ok := errs["rps"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, _ := strconv.ParseFloat(v, 64)
			cfg.RPS = n
			return nil
		}),
	)

	proxyPort := textinput.New(
		l, "Enter proxy port",
		textinput.WithValue(strconv.Itoa(int(cfg.ProxyPort))),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			n, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return err
			}
			tmp := *cfg
			tmp.ProxyPort = uint16(n)
			errs := tmp.Validate()
			if e, ok := errs["proxy_port"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			n, _ := strconv.ParseUint(v, 10, 16)
			cfg.ProxyPort = uint16(n)
			return nil
		}),
	)

	proxyType := selectinput.New(
		l, "Select proxy type",
		selectinput.WithValue(string(cfg.ProxyType)),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("SOCKS", "socks"),
			huh.NewOption("SSH", "ssh"),
		),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			pt := dns.ResolverProxyType(v)
			if pt == cfg.ProxyType {
				return nil
			}
			cfg.ProxyType = pt
			if cfg.ProxyType == dns.ResolverProxySOCKS {
				cfg.ProxyPort = 1080
				proxyPort.SetValue("1080")
			} else {
				cfg.ProxyPort = 22
				proxyPort.SetValue("22")
			}

			return m.refresh()
		}),
	)

	authMethod := selectinput.New(
		l, "Select authentication method",
		selectinput.WithValue(string(cfg.AuthMethod)),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("None", "none"),
			huh.NewOption("Password", "password"),
			huh.NewOption("Key", "key"),
		),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.AuthMethod = dns.AuthMethod(v)
			return nil
		}),
	)

	username := textinput.New(
		l, "Enter username",
		textinput.WithFocus(),
		textinput.WithValue(cfg.Username),
		textinput.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.Username = v
			errs := tmp.Validate()
			if e, ok := errs["username"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.Username = v
			return nil
		}),
	)

	password := textinput.New(
		l, "Enter password",
		textinput.WithValue(cfg.Password),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.Password = v
			errs := tmp.Validate()
			if e, ok := errs["password"]; ok {
				return e
			}
			return nil
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.Password = v
			return nil
		}),
	)

	privateKey := textarea.New(
		l, "Enter private key",
		textarea.WithValue(cfg.PrivateKey),
		textarea.WithFocus(),
		textarea.WithHeight(6),
		textarea.WithPlaceholder("-----BEGIN OPENSSH PRIVATE KEY----- ..."),
		textarea.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.PrivateKey = v
			errs := tmp.Validate()
			if e, ok := errs["private_key"]; ok {
				return e
			}
			return nil
		}),
		textarea.WithOnSubmit(func(v string) tea.Cmd {
			cfg.PrivateKey = strings.TrimSpace(v)
			return nil
		}),
	)

	proxyVisible := func() bool {
		return cfg.ProxyType != ""
	}

	authVisible := func() bool {
		return cfg.AuthMethod != dns.AuthNone
	}

	passwordVisible := func() bool {
		return cfg.AuthMethod == dns.AuthPassword
	}

	keyVisible := func() bool {
		return cfg.AuthMethod == dns.AuthKey
	}

	fields := []inspector.Field{
		{Name: "Config Name", Description: "The name of the configuration file.", Group: groupConnection, Input: inspector.Adapt(configName), Visible: alwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Domain", Description: descDomain, Group: groupConnection, Input: inspector.Adapt(domain), Visible: alwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Public Key", Description: descPubKey, Group: groupConnection, Input: inspector.Adapt(pubKey), Visible: alwaysVisible, Format: inspector.FormatPublicKey},
		{Name: "Resolver Type", Description: descResolverType, Group: groupConnection, Input: inspector.Adapt(resolverType), Visible: alwaysVisible},
		{Name: "Resolver Port", Description: descResolverPort, Group: groupConnection, Input: inspector.Adapt(resolverPort), Visible: alwaysVisible},
		{Name: "TLS Fingerprint", Description: descFingerprint, Group: groupConnection, Input: inspector.Adapt(fingerprint), Visible: alwaysVisible},
		{Name: "RPS", Description: descRPS, Group: groupConnection, Input: inspector.Adapt(rps), Visible: alwaysVisible, Format: inspector.FormatZeroAsAuto},

		{Name: "Proxy Type", Description: descProxyType, Group: groupProxyAuth, Input: inspector.Adapt(proxyType), Visible: alwaysVisible},
		{Name: "Proxy Port", Description: descProxyPort, Group: groupProxyAuth, Input: inspector.Adapt(proxyPort), Visible: proxyVisible},

		{Name: "Auth Method", Description: descAuthMethod, Group: groupProxyAuth, Input: inspector.Adapt(authMethod), Visible: alwaysVisible},
		{Name: "Username", Description: descUsername, Group: groupProxyAuth, Input: inspector.Adapt(username), Visible: authVisible, Format: inspector.FormatEmptyString},
		{Name: "Password", Description: descPassword, Group: groupProxyAuth, Input: inspector.Adapt(password), Visible: passwordVisible, Format: inspector.FormatEmptyString},
		{Name: "Private Key", Description: descPrivateKey, Group: groupProxyAuth, Input: inspector.Adapt(privateKey), Visible: keyVisible, Format: inspector.FormatPrivateKey},
	}

	return inspector.New(l, "dnstt config", fields)
}

func alwaysVisible() bool { return true }

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) Init() tea.Cmd {
	if m.form == nil {
		return nil
	}
	return m.form.Init()
}

func (m *Model) Mode() env.Mode {
	return env.ManagedMode
}

func (m *Model) OnClose() tea.Cmd {
	if m.form == nil {
		return nil
	}
	return m.form.OnClose()
}

func (m *Model) saveConfig() tea.Msg {
	srv := dns.NewDNSTTService()

	name := strings.TrimSpace(m.name)
	if name == "" {
		return notice.NewNoticeCmd(
			m.layout,
			"Error",
			"config name is required",
			notice.NOTICE_ERROR,
		)()
	}

	if m.originalName != "" {
		if err := srv.EditConfig(*m.cfg, m.originalName); err != nil {
			logger.UIError("Failed to edit DNSTT config: %v", err)
			return notice.NewNoticeCmd(
				m.layout,
				"Edit Failed",
				err.Error(),
				notice.NOTICE_ERROR,
			)()
		}

		if m.originalName != name {
			if err := srv.RenameConfig(m.originalName, name); err != nil {
				logger.UIError("Failed to rename DNSTT config: %v", err)
				return notice.NewNoticeCmd(
					m.layout,
					"Rename Failed",
					err.Error(),
					notice.NOTICE_ERROR,
				)()
			}
		}
	} else if err := srv.SaveConfig(*m.cfg, name); err != nil {
		logger.UIError("Failed to save DNSTT config: %v", err)
		return notice.NewNoticeCmd(
			m.layout,
			"Save Failed",
			err.Error(),
			notice.NOTICE_ERROR,
		)()
	}

	return tea.Sequence(
		notice.NewNoticeCmd(
			m.layout,
			"Saved",
			"DNSTT config saved",
			notice.NOTICE_SUCCESS,
		),
		func() tea.Msg {
			return ui.CloseComponentMsg{ID: m.ID()}
		},
	)()
}

func (m *Model) cancel() tea.Msg {
	return confirm.ConfirmCmd(
		m.layout,
		"Discard unsaved changes?",
		func() tea.Msg {
			return ui.CloseComponentMsg{ID: m.ID()}
		},
		false,
	)()
}

func (m *Model) refresh() tea.Cmd {
	if m.inspector != nil {
		return m.inspector.Refresh()
	}
	return nil
}
