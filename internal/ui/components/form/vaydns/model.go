package vaydns

import (
	"fmt"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/formkit"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

const (
	descClientIDSize = "Client ID size in bytes (1-8)."
	descMaxQnameLen  = "Maximum QNAME length (0-253)."
	descMaxNumLabels = "Maximum number of labels (0-4, 0=unlimited)."
	descMTU          = "Maximum Transmission Unit (0-1452)."
	descRecordType   = "DNS record type for queries."
)

// Model is the VayDNS tunnel configuration form.
type Model struct {
	formkit.Base
	cfg *dns.VayDNSConfig
}

// New creates the VayDNS config form, either for a new config or to edit an
// existing one described by original.
func New(
	l *layout.Layout,
	state *ui.AppState,
	original *dns.DNSTunConfigFile,
) (*Model, error) {
	cfg := dns.DefaultVayDNSConfig()
	name := ""
	originalName := ""

	if original != nil {
		name = original.Name
		originalName = original.Name
		if c, ok := original.Config.(dns.VayDNSConfig); ok {
			cfg = c
		}
	}

	m := &Model{cfg: &cfg}
	m.Base = formkit.NewBase(l, state, name, originalName)
	m.buildForm(original)

	return m, nil
}

func (m *Model) buildForm(original *dns.DNSTunConfigFile) {
	title := "New VayDNS Config"
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
		func(c dns.VayDNSConfig) string { return c.Domain },
		func(c *dns.VayDNSConfig, v string) { c.Domain = strings.TrimSpace(v) },
	)
	pubKey := formkit.SecretField(
		l, cfg, "Enter public key", "pub_key",
		func(c dns.VayDNSConfig) string { return c.PubKey },
		func(c *dns.VayDNSConfig, v string) { c.PubKey = v },
		textarea.WithNewlines(false),
	)

	recordType := selectinput.New(
		l, "Select record type",
		selectinput.WithValue(string(cfg.RecordType)),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("A", "A"),
			huh.NewOption("AAAA", "AAAA"),
			huh.NewOption("CNAME", "CNAME"),
			huh.NewOption("NS", "NS"),
			huh.NewOption("MX", "MX"),
			huh.NewOption("TXT", "TXT"),
			huh.NewOption("SRV", "SRV"),
			huh.NewOption("NULL", "NULL"),
			huh.NewOption("CAA", "CAA"),
		),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.RecordType = dns.RecordType(v)
			if e, ok := tmp.Validate()["record_type"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.RecordType = dns.RecordType(v)
			return nil
		}),
	)

	resType, resPort := formkit.ResolverTypePort(
		l, cfg, m.Refresh,
		func(c dns.VayDNSConfig) dns.ResolverType { return c.ResolverType },
		func(c *dns.VayDNSConfig, v dns.ResolverType) { c.ResolverType = v },
		func(c dns.VayDNSConfig) uint16 { return c.ResolverPort },
		func(c *dns.VayDNSConfig, v uint16) { c.ResolverPort = v },
	)
	fingerprint := formkit.FingerprintField(
		l, cfg,
		func(c dns.VayDNSConfig) string { return c.Fingerprint },
		func(c *dns.VayDNSConfig, v string) { c.Fingerprint = v },
	)

	clientIDSize := formkit.Uint16Field(
		l, cfg, "Enter client ID size", "client_id_size",
		func(c dns.VayDNSConfig) uint16 { return c.ClientIDSize },
		func(c *dns.VayDNSConfig, v uint16) { c.ClientIDSize = v },
		textinput.WithPlaceholder("1-8"),
	)
	maxQnameLen := formkit.Uint8Field(
		l, cfg, "Enter max QNAME length", "max_qname_len",
		func(c dns.VayDNSConfig) uint8 { return c.MaxQnameLen },
		func(c *dns.VayDNSConfig, v uint8) { c.MaxQnameLen = v },
		textinput.WithPlaceholder("0-253"),
	)
	maxNumLabels := formkit.Uint8Field(
		l, cfg, "Enter max labels", "max_num_labels",
		func(c dns.VayDNSConfig) uint8 { return c.MaxNumLabels },
		func(c *dns.VayDNSConfig, v uint8) { c.MaxNumLabels = v },
		textinput.WithPlaceholder("0-4, 0=auto"),
	)
	mtu := formkit.Uint16Field(
		l, cfg, "Enter MTU", "mtu",
		func(c dns.VayDNSConfig) uint16 { return c.MTU },
		func(c *dns.VayDNSConfig, v uint16) { c.MTU = v },
		textinput.WithPlaceholder("0-1452 (0 - auto)"),
	)
	rps := formkit.FloatField(
		l, cfg, "Enter requests per second", "rps",
		func(c dns.VayDNSConfig) float64 { return c.RPS },
		func(c *dns.VayDNSConfig, v float64) { c.RPS = v },
		textinput.WithPlaceholder("0 = unlimited"),
	)

	proxy, vis := formkit.BuildProxy(
		l, cfg, m.Refresh,
		func(c dns.VayDNSConfig) dns.ResolverProxyType { return c.ProxyType },
		func(c *dns.VayDNSConfig, v dns.ResolverProxyType) { c.ProxyType = v },
		func(c dns.VayDNSConfig) uint16 { return c.ProxyPort },
		func(c *dns.VayDNSConfig, v uint16) { c.ProxyPort = v },
		func(c dns.VayDNSConfig) dns.AuthMethod { return c.AuthMethod },
		func(c *dns.VayDNSConfig, v dns.AuthMethod) { c.AuthMethod = v },
		func(c dns.VayDNSConfig) string { return c.Username },
		func(c *dns.VayDNSConfig, v string) { c.Username = v },
		func(c dns.VayDNSConfig) string { return c.Password },
		func(c *dns.VayDNSConfig, v string) { c.Password = v },
		func(c dns.VayDNSConfig) string { return c.PrivateKey },
		func(c *dns.VayDNSConfig, v string) { c.PrivateKey = v },
	)

	fields := []inspector.Field{
		{Name: "Config Name", Description: formkit.DescConfigName, Group: formkit.GroupConnection, Input: inspector.Adapt(configName), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Domain", Description: formkit.DescDomain, Group: formkit.GroupConnection, Input: inspector.Adapt(domain), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Public Key", Description: formkit.DescPubKey, Group: formkit.GroupConnection, Input: inspector.Adapt(pubKey), Visible: formkit.AlwaysVisible, Format: inspector.FormatPublicKey},
		{Name: "Resolver Type", Description: formkit.DescResolverType, Group: formkit.GroupConnection, Input: inspector.Adapt(resType), Visible: formkit.AlwaysVisible},
		{Name: "Resolver Port", Description: formkit.DescResolverPort, Group: formkit.GroupConnection, Input: inspector.Adapt(resPort), Visible: formkit.AlwaysVisible},
		{Name: "TLS Fingerprint", Description: formkit.DescFingerprint, Group: formkit.GroupConnection, Input: inspector.Adapt(fingerprint), Visible: formkit.AlwaysVisible},
		{Name: "Record Type", Description: descRecordType, Group: formkit.GroupConnection, Input: inspector.Adapt(recordType), Visible: formkit.AlwaysVisible},

		{Name: "Client ID Size", Description: descClientIDSize, Group: formkit.GroupAdvanced, Input: inspector.Adapt(clientIDSize), Visible: formkit.AlwaysVisible},
		{Name: "Max QNAME Len", Description: descMaxQnameLen, Group: formkit.GroupAdvanced, Input: inspector.Adapt(maxQnameLen), Visible: formkit.AlwaysVisible, Format: inspector.FormatZeroAsAuto},
		{Name: "Max Labels", Description: descMaxNumLabels, Group: formkit.GroupAdvanced, Input: inspector.Adapt(maxNumLabels), Visible: formkit.AlwaysVisible, Format: inspector.FormatZeroAsAuto},
		{Name: "MTU", Description: descMTU, Group: formkit.GroupAdvanced, Input: inspector.Adapt(mtu), Visible: formkit.AlwaysVisible, Format: inspector.FormatZeroAsAuto},
		{Name: "RPS", Description: formkit.DescRPS, Group: formkit.GroupAdvanced, Input: inspector.Adapt(rps), Visible: formkit.AlwaysVisible, Format: inspector.FormatZeroAsAuto},

		{Name: "Proxy Type", Description: formkit.DescProxyType, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Type), Visible: formkit.AlwaysVisible},
		{Name: "Proxy Port", Description: formkit.DescProxyPort, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Port), Visible: vis.Proxy},

		{Name: "Auth Method", Description: formkit.DescAuthMethod, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Auth), Visible: formkit.AlwaysVisible},
		{Name: "Username", Description: formkit.DescUsername, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Username), Visible: vis.Auth, Format: inspector.FormatEmptyString},
		{Name: "Password", Description: formkit.DescPassword, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.Password), Visible: vis.Password, Format: inspector.FormatEmptyString},
		{Name: "Private Key", Description: formkit.DescPrivateKey, Group: formkit.GroupProxyAuth, Input: inspector.Adapt(proxy.PrivateKey), Visible: vis.Key, Format: inspector.FormatPrivateKey},
	}

	return inspector.New(l, "vaydns config", fields)
}

func (m *Model) saveConfig() tea.Msg {
	return formkit.Save(&m.Base, "VayDNS", dns.NewVayDNSService(), m.cfg)
}
