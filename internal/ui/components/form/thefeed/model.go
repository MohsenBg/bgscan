package thefeed

import (
	"fmt"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/components/form/formkit"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Model is the TheFeed tunnel configuration form.
type Model struct {
	formkit.Base
	cfg *dns.TheFeedConfig
}

// New creates the TheFeed config form, either for a new config or to edit an
// existing one described by original.
func New(
	l *layout.Layout,
	state *ui.AppState,
	original *dns.DNSTunConfigFile,
) (*Model, error) {
	cfg := dns.DefaultTheFeedConfig()
	name := ""
	originalName := ""

	if original != nil {
		name = original.Name
		originalName = original.Name
		if c, ok := original.Config.(dns.TheFeedConfig); ok {
			cfg = c
		}
	}

	m := &Model{cfg: &cfg}
	m.Base = formkit.NewBase(l, state, name, originalName, formkit.WithMaxHeight(35))
	m.buildForm(original)

	return m, nil
}

func (m *Model) buildForm(original *dns.DNSTunConfigFile) {
	title := "New TheFeed Config"
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
		func(c dns.TheFeedConfig) string { return c.Domain },
		func(c *dns.TheFeedConfig, v string) { c.Domain = strings.TrimSpace(v) },
	)
	passphrase := formkit.StringField(
		l, cfg, "Enter passphrase", "passphrase",
		func(c dns.TheFeedConfig) string { return c.Passphrase },
		func(c *dns.TheFeedConfig, v string) { c.Passphrase = strings.TrimSpace(v) },
	)

	resType, resPort := formkit.ResolverTypePort(
		l, cfg, m.Refresh,
		func(c dns.TheFeedConfig) dns.ResolverType { return c.ResolverType },
		func(c *dns.TheFeedConfig, v dns.ResolverType) { c.ResolverType = v },
		func(c dns.TheFeedConfig) uint16 { return c.ResolverPort },
		func(c *dns.TheFeedConfig, v uint16) { c.ResolverPort = v },
	)

	queryMode := selectinput.New(
		l, "Select query mode",
		selectinput.WithValue(cfg.QueryMode),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(
			huh.NewOption("Single", "single"),
			huh.NewOption("Double", "double"),
		),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			tmp.QueryMode = v
			if e, ok := tmp.Validate()["query_mode"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			cfg.QueryMode = v
			return nil
		}),
	)

	fields := []inspector.Field{
		{Name: "Config Name", Description: formkit.DescConfigName, Group: formkit.GroupConnection, Input: inspector.Adapt(configName), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Domain", Description: formkit.DescDomain, Group: formkit.GroupConnection, Input: inspector.Adapt(domain), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Passphrase", Description: formkit.DescPassphrase, Group: formkit.GroupConnection, Input: inspector.Adapt(passphrase), Visible: formkit.AlwaysVisible, Format: inspector.FormatEmptyString},
		{Name: "Resolver Type", Description: formkit.DescResolverType, Group: formkit.GroupConnection, Input: inspector.Adapt(resType), Visible: formkit.AlwaysVisible},
		{Name: "Resolver Port", Description: formkit.DescResolverPort, Group: formkit.GroupConnection, Input: inspector.Adapt(resPort), Visible: formkit.AlwaysVisible},
		{Name: "Query Mode", Description: formkit.DescQueryMode, Group: formkit.GroupConnection, Input: inspector.Adapt(queryMode), Visible: formkit.AlwaysVisible},
	}

	return inspector.New(l, "thefeed config", fields)
}

func (m *Model) saveConfig() tea.Msg {
	return formkit.Save(&m.Base, "TheFeed", dns.NewTheFeedService(), m.cfg)
}
