package dnstt

import (
	"fmt"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/formkit"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the DNSTT tunnel configuration form.
type Model struct {
	formkit.Base
	cfg *dns.DNSTTConfig
}

// New creates the DNSTT config form, either for a new config or to edit an
// existing one described by original.
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

	m := &Model{cfg: &cfg}
	m.Base = formkit.NewBase(l, state, name, originalName)
	m.buildForm(original)

	return m, nil
}

func (m *Model) buildForm(original *dns.DNSTunConfigFile) {
	title := "New DNSTT Config"
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
		func(c dns.DNSTTConfig) string { return c.Domain },
		func(c *dns.DNSTTConfig, v string) { c.Domain = strings.TrimSpace(v) },
	)
	pubKey := formkit.SecretField(
		l, cfg, "Enter public key", "pub_key",
		func(c dns.DNSTTConfig) string { return c.PubKey },
		func(c *dns.DNSTTConfig, v string) { c.PubKey = strings.TrimSpace(v) },
		textarea.WithNewlines(false),
	)
	resType, resPort := formkit.ResolverTypePort(
		l, cfg, m.Refresh,
		func(c dns.DNSTTConfig) dns.ResolverType { return c.ResolverType },
		func(c *dns.DNSTTConfig, v dns.ResolverType) { c.ResolverType = v },
		func(c dns.DNSTTConfig) uint16 { return c.ResolverPort },
		func(c *dns.DNSTTConfig, v uint16) { c.ResolverPort = v },
	)
	fingerprint := formkit.FingerprintField(
		l, cfg,
		func(c dns.DNSTTConfig) string { return c.Fingerprint },
		func(c *dns.DNSTTConfig, v string) { c.Fingerprint = v },
	)
	rps := formkit.FloatField(
		l, cfg, "Enter requests per second", "rps",
		func(c dns.DNSTTConfig) float64 { return c.RPS },
		func(c *dns.DNSTTConfig, v float64) { c.RPS = v },
		textinput.WithPlaceholder("0 = unlimited"),
	)

	proxy, vis := formkit.BuildProxy(
		l, cfg, m.Refresh,
		func(c dns.DNSTTConfig) dns.ResolverProxyType { return c.ProxyType },
		func(c *dns.DNSTTConfig, v dns.ResolverProxyType) { c.ProxyType = v },
		func(c dns.DNSTTConfig) uint16 { return c.ProxyPort },
		func(c *dns.DNSTTConfig, v uint16) { c.ProxyPort = v },
		func(c dns.DNSTTConfig) dns.AuthMethod { return c.AuthMethod },
		func(c *dns.DNSTTConfig, v dns.AuthMethod) { c.AuthMethod = v },
		func(c dns.DNSTTConfig) string { return c.Username },
		func(c *dns.DNSTTConfig, v string) { c.Username = v },
		func(c dns.DNSTTConfig) string { return c.Password },
		func(c *dns.DNSTTConfig, v string) { c.Password = v },
		func(c dns.DNSTTConfig) string { return c.PrivateKey },
		func(c *dns.DNSTTConfig, v string) { c.PrivateKey = strings.TrimSpace(v) },
	)

	fields := []inspector.Field{
		{Name: "Config Name", Description: formkit.DescConfigName, Group: formkit.GroupConnection, Input: inspector.Adapt(configName), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Domain", Description: formkit.DescDomain, Group: formkit.GroupConnection, Input: inspector.Adapt(domain), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Public Key", Description: formkit.DescPubKey, Group: formkit.GroupConnection, Input: inspector.Adapt(pubKey), Visible: formkit.AlwaysVisible, Format: inspector.FormatPublicKey},
		{Name: "Resolver Type", Description: formkit.DescResolverType, Group: formkit.GroupConnection, Input: inspector.Adapt(resType), Visible: formkit.AlwaysVisible},
		{Name: "Resolver Port", Description: formkit.DescResolverPort, Group: formkit.GroupConnection, Input: inspector.Adapt(resPort), Visible: formkit.AlwaysVisible},
		{Name: "TLS Fingerprint", Description: formkit.DescFingerprint, Group: formkit.GroupConnection, Input: inspector.Adapt(fingerprint), Visible: formkit.AlwaysVisible},
		{Name: "RPS", Description: formkit.DescRPS, Group: formkit.GroupConnection, Input: inspector.Adapt(rps), Visible: formkit.AlwaysVisible, Format: inspector.FormatZeroAsAuto},

		{Name: "Proxy Type", Description: formkit.DescProxyType, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Type), Visible: formkit.AlwaysVisible},
		{Name: "Proxy Port", Description: formkit.DescProxyPort, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Port), Visible: vis.Proxy},

		{Name: "Auth Method", Description: formkit.DescAuthMethod, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Auth), Visible: formkit.AlwaysVisible},
		{Name: "Username", Description: formkit.DescUsername, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Username), Visible: vis.Auth, Format: inspector.FormatEmptyString},
		{Name: "Password", Description: formkit.DescPassword, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Password), Visible: vis.Password, Format: inspector.FormatEmptyString},
		{Name: "Private Key", Description: formkit.DescPrivateKey, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.PrivateKey), Visible: vis.Key, Format: inspector.FormatPrivateKey},
	}

	return inspector.New(l, "dnstt config", fields)
}

func (m *Model) saveConfig() tea.Msg {
	return formkit.Save(&m.Base, "DNSTT", dns.NewDNSTTService(), m.cfg)
}
