package slipstream

import (
	"fmt"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/formkit"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

const (
	descCertPath = "Optional path to a custom TLS certificate file."
)

// Model is the Slipstream tunnel configuration form.
type Model struct {
	formkit.Base
	cfg           *dns.SlipstreamConfig
	slipstreamSrv dns.SlipstreamService
}

// New creates the Slipstream config form, either for a new config or to edit
// an existing one described by original.
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
		cfg:           &cfg,
		slipstreamSrv: srv,
	}
	m.Base = formkit.NewBase(l, state, name, originalName)
	m.buildForm(original)

	return m, nil
}

func (m *Model) buildForm(original *dns.DNSTunConfigFile) {
	title := "New Slipstream Config"
	if original != nil {
		title = fmt.Sprintf("Edit %s", m.Name())
	}

	m.BuildForm(
		title,
		m.buildInspector(),
		func() map[string]error { return m.cfg.Validate() },
		m.saveConfig,
	)
}

func (m *Model) buildInspector() *inspector.Model {
	cfg := m.cfg
	l := m.Layout()

	configName := formkit.ConfigNameField(l, m.Name(), m.SetName)
	domain := formkit.StringField(
		l, cfg, "Enter domain", "domain",
		func(c dns.SlipstreamConfig) string { return c.Domain },
		func(c *dns.SlipstreamConfig, v string) { c.Domain = strings.TrimSpace(v) },
	)
	resPort := formkit.Uint16Field(
		l, cfg, "Enter resolver port", "resolver_port",
		func(c dns.SlipstreamConfig) uint16 { return c.ResolverPort },
		func(c *dns.SlipstreamConfig, v uint16) { c.ResolverPort = v },
	)
	certPath := formkit.StringField(
		l, cfg, "Enter certificate path", "cert_path",
		func(c dns.SlipstreamConfig) string { return c.CertPath },
		func(c *dns.SlipstreamConfig, v string) { c.CertPath = strings.TrimSpace(v) },
	)

	proxy, vis := formkit.BuildProxy(
		l, cfg, m.Refresh,
		func(c dns.SlipstreamConfig) dns.ResolverProxyType { return c.ProxyType },
		func(c *dns.SlipstreamConfig, v dns.ResolverProxyType) { c.ProxyType = v },
		func(c dns.SlipstreamConfig) uint16 { return c.ProxyPort },
		func(c *dns.SlipstreamConfig, v uint16) { c.ProxyPort = v },
		func(c dns.SlipstreamConfig) dns.AuthMethod { return c.AuthMethod },
		func(c *dns.SlipstreamConfig, v dns.AuthMethod) { c.AuthMethod = v },
		func(c dns.SlipstreamConfig) string { return c.Username },
		func(c *dns.SlipstreamConfig, v string) { c.Username = v },
		func(c dns.SlipstreamConfig) string { return c.Password },
		func(c *dns.SlipstreamConfig, v string) { c.Password = v },
		func(c dns.SlipstreamConfig) string { return c.PrivateKey },
		func(c *dns.SlipstreamConfig, v string) { c.PrivateKey = v },
	)

	fields := []inspector.Field{
		{Name: "Config Name", Description: formkit.DescConfigName, Group: formkit.GroupConnection, Input: inspector.Adapt(configName), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Domain", Description: formkit.DescDomain, Group: formkit.GroupConnection, Input: inspector.Adapt(domain), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Resolver Port", Description: formkit.DescResolverPort, Group: formkit.GroupConnection, Input: inspector.Adapt(resPort), Visible: formkit.AlwaysVisible},
		{Name: "Cert Path", Description: descCertPath, Group: formkit.GroupConnection, Input: inspector.Adapt(certPath), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},

		{Name: "Proxy Type", Description: formkit.DescProxyType, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Type), Visible: formkit.AlwaysVisible},
		{Name: "Proxy Port", Description: formkit.DescProxyPort, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Port), Visible: vis.Proxy},

		{Name: "Auth Method", Description: formkit.DescAuthMethod, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Auth), Visible: formkit.AlwaysVisible},
		{Name: "Username", Description: formkit.DescUsername, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Username), Visible: vis.Auth, Format: inspector.FormatEmptyString},
		{Name: "Password", Description: formkit.DescPassword, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Password), Visible: vis.Password, Format: inspector.FormatEmptyString},
		{Name: "Private Key", Description: formkit.DescPrivateKey, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.PrivateKey), Visible: vis.Key, Format: inspector.FormatPrivateKey},
	}

	return inspector.New(l, "slipstream config", fields)
}

func (m *Model) saveConfig() tea.Msg {
	return formkit.Save(&m.Base, "Slipstream", m.slipstreamSrv, m.cfg)
}
