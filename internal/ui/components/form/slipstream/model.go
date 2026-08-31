package slipstream

import (
	"fmt"
	"strconv"
	"strings"

	"bgscan/internal/core/dns"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/components/basic/form"
	"bgscan/internal/ui/components/basic/input/selectinput"
	"bgscan/internal/ui/components/basic/input/textarea"
	"bgscan/internal/ui/components/basic/input/textinput"
	"bgscan/internal/ui/components/basic/inspector"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/shared/validation"

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
	descDomain       = "The target domain name for the Slipstream DNS tunnel."
	descResolverPort = "The DNS resolver port (typically 53)."
	descCertPath     = "Optional path to a custom TLS certificate file."
	descProxyType    = "The proxy type used for the resolver connection."
	descProxyPort    = "The proxy port number."
	descAuthMethod   = "The authentication method: none, password, or key."
	descUsername     = "The username for authentication."
	descPassword     = "The password for password authentication."
	descPrivateKey   = "The SSH private key used for authentication."
)

type Model struct {
	id            ui.ComponentID
	layout        *layout.Layout
	state         *ui.AppState
	form          *form.Model
	inspector     *inspector.Model
	cfg           *dns.SlipstreamConfig
	slipstreamSrv *dns.SlipstreamService
	name          string
	originalName  string
	width         int
	height        int
}

func New(
	l *layout.Layout,
	state *ui.AppState,
	original *dns.DNSTunConfigFile,
) (*Model, error) {
	cfg := dns.DefaultSlipstreamConfig()
	name := ""
	originalName := ""

	if original != nil {
		name = original.Name
		originalName = original.Name
		if c, ok := original.Config.(dns.SlipstreamConfig); ok {
			cfg = c
		}
	}

	srv, err := dns.NewSlipstreamService()
	if err != nil {
		return nil, err
	}

	m := &Model{
		id:            ui.NewComponentID(),
		layout:        l,
		state:         state,
		cfg:           &cfg,
		name:          name,
		originalName:  originalName,
		slipstreamSrv: &srv,
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

	title := "New Slipstream Config"
	if original != nil {
		title = fmt.Sprintf("Edit %s", m.name)
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
			if e, ok := errs["dns_port"]; ok {
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

	certPath := textinput.New(
		l, "Enter certificate path",
		textinput.WithValue(cfg.CertPath),
		textinput.WithFocus(),
		textinput.WithPlaceholder("(optional)"),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.CertPath = strings.TrimSpace(v)
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
		textinput.WithValue(cfg.Username),
		textinput.WithFocus(),
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
			cfg.PrivateKey = v
			return nil
		}),
	)

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
		{Name: "Resolver Port", Description: descResolverPort, Group: groupConnection, Input: inspector.Adapt(resolverPort), Visible: alwaysVisible},
		{Name: "Cert Path", Description: descCertPath, Group: groupConnection, Input: inspector.Adapt(certPath), Visible: alwaysVisible, Format: inspector.FormatEmptyString},

		{Name: "Proxy Type", Description: descProxyType, Group: groupProxyAuth, Input: inspector.Adapt(proxyType), Visible: alwaysVisible},
		{Name: "Proxy Port", Description: descProxyPort, Group: groupProxyAuth, Input: inspector.Adapt(proxyPort), Visible: alwaysVisible},

		{Name: "Auth Method", Description: descAuthMethod, Group: groupProxyAuth, Input: inspector.Adapt(authMethod), Visible: alwaysVisible},
		{Name: "Username", Description: descUsername, Group: groupProxyAuth, Input: inspector.Adapt(username), Visible: authVisible, Format: inspector.FormatEmptyString},
		{Name: "Password", Description: descPassword, Group: groupProxyAuth, Input: inspector.Adapt(password), Visible: passwordVisible, Format: inspector.FormatEmptyString},
		{Name: "Private Key", Description: descPrivateKey, Group: groupProxyAuth, Input: inspector.Adapt(privateKey), Visible: keyVisible, Format: inspector.FormatPrivateKey},
	}

	return inspector.New(l, "slipstream config", fields)
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
		if err := (*m.slipstreamSrv).EditConfig(*m.cfg, m.originalName); err != nil {
			logger.UIError("Failed to edit Slipstream config: %v", err)
			return notice.NewNoticeCmd(
				m.layout,
				"Edit Failed",
				err.Error(),
				notice.NOTICE_ERROR,
			)()
		}

		if m.originalName != name {
			if err := (*m.slipstreamSrv).RenameConfig(m.originalName, name); err != nil {
				logger.UIError("Failed to rename Slipstream config: %v", err)
				return notice.NewNoticeCmd(
					m.layout,
					"Rename Failed",
					err.Error(),
					notice.NOTICE_ERROR,
				)()
			}
		}
	} else if err := (*m.slipstreamSrv).SaveConfig(*m.cfg, name); err != nil {
		logger.UIError("Failed to save Slipstream config: %v", err)
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
			"Slipstream config saved",
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
