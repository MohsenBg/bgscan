package masterdns

import (
	"fmt"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/formkit"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the MasterDNS tunnel configuration form.
type Model struct {
	formkit.Base
	cfg *dns.MasterDNSConfig
}

// New creates the MasterDNS config form, either for a new config or to edit
// an existing one described by original.
func New(
	l *layout.Layout,
	state *ui.AppState,
	original *dns.DNSTunConfigFile,
) (*Model, error) {
	cfg := dns.DefaultMasterDNSConfig()
	name := ""
	originalName := ""

	if original != nil {
		name = original.Name
		originalName = original.Name
		if c, ok := original.Config.(dns.MasterDNSConfig); ok {
			cfg = c
		}
	}

	m := &Model{cfg: &cfg}
	m.Base = formkit.NewBase(l, state, name, originalName)
	m.buildForm(original)

	return m, nil
}

func (m *Model) buildForm(original *dns.DNSTunConfigFile) {
	title := "New MasterDNS Config"
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
		func(c dns.MasterDNSConfig) string { return c.Domain },
		func(c *dns.MasterDNSConfig, v string) { c.Domain = strings.TrimSpace(v) },
	)
	encryptionKey := formkit.SecretField(
		l, cfg, "Enter encryption key", "encryption_key",
		func(c dns.MasterDNSConfig) string { return c.EncryptionKey },
		func(c *dns.MasterDNSConfig, v string) { c.EncryptionKey = v },
		textarea.WithNewlines(false),
	)
	encMethod := formkit.EncMethodField(
		l, cfg, "Select encryption method",
		func(c dns.MasterDNSConfig) dns.EncMethod { return c.DataEncMethod },
		func(c *dns.MasterDNSConfig, v dns.EncMethod) { c.DataEncMethod = v },
	)
	resPort := formkit.Uint16Field(
		l, cfg, "Enter resolver port", "resolver_port",
		func(c dns.MasterDNSConfig) uint16 { return c.ResolverPort },
		func(c *dns.MasterDNSConfig, v uint16) { c.ResolverPort = v },
	)

	mtuTestTimeout := formkit.FloatField(
		l, cfg, "Enter MTU test timeout", "mtu_test_timeout_sec",
		func(c dns.MasterDNSConfig) float64 { return c.MTUTestTimeoutSec },
		func(c *dns.MasterDNSConfig, v float64) { c.MTUTestTimeoutSec = v },
	)
	mtuTestRetries := formkit.Uint8Field(
		l, cfg, "Enter MTU test retries", "mtu_test_retries",
		func(c dns.MasterDNSConfig) uint8 { return c.MTUTestRetries },
		func(c *dns.MasterDNSConfig, v uint8) { c.MTUTestRetries = v },
	)
	sessionInitRetryMax := formkit.FloatField(
		l, cfg, "Enter session init retry max", "session_init_retry_max_sec",
		func(c dns.MasterDNSConfig) float64 { return c.SessionInitRetryMaxSec },
		func(c *dns.MasterDNSConfig, v float64) { c.SessionInitRetryMaxSec = v },
	)
	sessionInitRacing := formkit.Uint8Field(
		l, cfg, "Enter session init racing count", "session_init_racing_count",
		func(c dns.MasterDNSConfig) uint8 { return c.SessionInitRacingCount },
		func(c *dns.MasterDNSConfig, v uint8) { c.SessionInitRacingCount = v },
	)
	minUploadMTU := formkit.Uint16Field(
		l, cfg, "Enter min upload MTU", "min_upload_mtu",
		func(c dns.MasterDNSConfig) uint16 { return c.MinUploadMTU },
		func(c *dns.MasterDNSConfig, v uint16) { c.MinUploadMTU = v },
	)
	maxUploadMTU := formkit.Uint16Field(
		l, cfg, "Enter max upload MTU", "max_upload_mtu",
		func(c dns.MasterDNSConfig) uint16 { return c.MaxUploadMTU },
		func(c *dns.MasterDNSConfig, v uint16) { c.MaxUploadMTU = v },
	)
	minDownloadMTU := formkit.Uint16Field(
		l, cfg, "Enter min download MTU", "min_download_mtu",
		func(c dns.MasterDNSConfig) uint16 { return c.MinDownloadMTU },
		func(c *dns.MasterDNSConfig, v uint16) { c.MinDownloadMTU = v },
	)
	maxDownloadMTU := formkit.Uint16Field(
		l, cfg, "Enter max download MTU", "max_download_mtu",
		func(c dns.MasterDNSConfig) uint16 { return c.MaxDownloadMTU },
		func(c *dns.MasterDNSConfig, v uint16) { c.MaxDownloadMTU = v },
	)
	mtuParallelism := formkit.Uint8Field(
		l, cfg, "Enter MTU parallelism", "mtu_parallelism",
		func(c dns.MasterDNSConfig) uint8 { return c.MTUParallelism },
		func(c *dns.MasterDNSConfig, v uint8) { c.MTUParallelism = v },
	)
	rxTxWorkers := formkit.Uint8Field(
		l, cfg, "Enter RX/TX workers", "rx_tx_workers",
		func(c dns.MasterDNSConfig) uint8 { return c.RxTxWorkers },
		func(c *dns.MasterDNSConfig, v uint8) { c.RxTxWorkers = v },
	)

	fields := []inspector.Field{
		{Name: "Config Name", Description: formkit.DescConfigName, Group: formkit.GroupConnection, Input: inspector.Adapt(configName), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Domain", Description: formkit.DescDomain, Group: formkit.GroupConnection, Input: inspector.Adapt(domain), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Encryption Key", Description: formkit.DescEncryptionKey, Group: formkit.GroupConnection, Input: inspector.Adapt(encryptionKey), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Encryption Method", Description: formkit.DescEncMethod, Group: formkit.GroupConnection, Input: inspector.Adapt(encMethod), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Resolver Port", Description: formkit.DescResolverPort, Group: formkit.GroupConnection, Input: inspector.Adapt(resPort), Visible: formkit.AlwaysVisible},

		{Name: "MTU Test Timeout", Description: formkit.DescMTUTestTimeout, Group: formkit.GroupAdvanced, Input: inspector.Adapt(mtuTestTimeout), Visible: formkit.AlwaysVisible, Format: inspector.FormatSeconds},
		{Name: "MTU Test Retries", Description: formkit.DescMTUTestRetries, Group: formkit.GroupAdvanced, Input: inspector.Adapt(mtuTestRetries), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "Session Init Retry Max", Description: formkit.DescSessionInitRetryMax, Group: formkit.GroupAdvanced, Input: inspector.Adapt(sessionInitRetryMax), Visible: formkit.AlwaysVisible, Format: inspector.FormatSeconds},
		{Name: "Session Init Racing", Description: formkit.DescSessionInitRacing, Group: formkit.GroupAdvanced, Input: inspector.Adapt(sessionInitRacing), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "Min Upload MTU", Description: formkit.DescMinUploadMTU, Group: formkit.GroupAdvanced, Input: inspector.Adapt(minUploadMTU), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "Max Upload MTU", Description: formkit.DescMaxUploadMTU, Group: formkit.GroupAdvanced, Input: inspector.Adapt(maxUploadMTU), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "Min Download MTU", Description: formkit.DescMinDownloadMTU, Group: formkit.GroupAdvanced, Input: inspector.Adapt(minDownloadMTU), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "Max Download MTU", Description: formkit.DescMaxDownloadMTU, Group: formkit.GroupAdvanced, Input: inspector.Adapt(maxDownloadMTU), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "MTU Parallelism", Description: formkit.DescMTUParallelism, Group: formkit.GroupAdvanced, Input: inspector.Adapt(mtuParallelism), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
		{Name: "RX/TX Workers", Description: formkit.DescRxTxWorkers, Group: formkit.GroupAdvanced, Input: inspector.Adapt(rxTxWorkers), Visible: formkit.AlwaysVisible, Format: inspector.FormatInt},
	}

	return inspector.New(l, "masterdns config", fields)
}

func (m *Model) saveConfig() tea.Msg {
	return formkit.Save(&m.Base, "MasterDNS", dns.NewMasterDNSService(), m.cfg)
}
