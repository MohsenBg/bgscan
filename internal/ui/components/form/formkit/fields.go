package formkit

import (
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/selectinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textarea"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// ConfigNameField builds the config-name input bound to the form name.
func ConfigNameField(l *layout.Layout, name string, set func(string)) input.Input[string] {
	return textinput.New(
		l, "Enter config name",
		textinput.WithValue(name),
		textinput.WithFocus(),
		textinput.WithValidation(func(v string) error {
			return validation.ValidateFilename(v)
		}),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			set(strings.TrimSpace(v))
			return nil
		}),
	)
}

// StringField builds a single-line text input bound to a string config field,
// validating through a copy of the whole config keyed by errKey.
func StringField[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	title, errKey string,
	get func(C) string,
	set func(*C, string),
) input.Input[string] {
	return buildTextInput(
		l, title, get(*cfg),
		func(v string) error {
			tmp := *cfg
			set(&tmp, v)
			if e, ok := tmp.Validate()[errKey]; ok {
				return e
			}
			return nil
		},
		func(v string) { set(cfg, v) },
	)
}

// SecretField builds a textarea bound to a sensitive/secret config field
// (e.g. public key, private key, encryption key), validating through a copy.
func SecretField[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	title, errKey string,
	get func(C) string,
	set func(*C, string),
	opts ...textarea.Option,
) input.Input[string] {
	o := []textarea.Option{
		textarea.WithValue(get(*cfg)),
		textarea.WithFocus(),
		textarea.WithValidation(func(v string) error {
			tmp := *cfg
			set(&tmp, v)
			if e, ok := tmp.Validate()[errKey]; ok {
				return e
			}
			return nil
		}),
		textarea.WithOnSubmit(func(v string) tea.Cmd {
			set(cfg, v)
			return nil
		}),
	}
	return textarea.New(l, title, append(o, opts...)...)
}

// Uint8Field builds a numeric input bound to a uint8 config field, using
// overflow-safe parsing (bit size 8) and config-level validation.
func Uint8Field[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	title, errKey string,
	get func(C) uint8,
	set func(*C, uint8),
	opts ...textinput.Option,
) input.Input[string] {
	return buildUintInput(l, title, 8, uint64(get(*cfg)), func(n uint64) error {
		tmp := *cfg
		set(&tmp, uint8(n))
		if e, ok := tmp.Validate()[errKey]; ok {
			return e
		}
		return nil
	}, func(n uint64) { set(cfg, uint8(n)) }, opts...)
}

// Uint16Field builds a numeric input bound to a uint16 config field, using
// overflow-safe parsing (bit size 16) and config-level validation.
func Uint16Field[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	title, errKey string,
	get func(C) uint16,
	set func(*C, uint16),
	opts ...textinput.Option,
) input.Input[string] {
	return buildUintInput(l, title, 16, uint64(get(*cfg)), func(n uint64) error {
		tmp := *cfg
		set(&tmp, uint16(n))
		if e, ok := tmp.Validate()[errKey]; ok {
			return e
		}
		return nil
	}, func(n uint64) { set(cfg, uint16(n)) }, opts...)
}

// FloatField builds a numeric input bound to a float64 config field with
// config-level validation.
func FloatField[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	title, errKey string,
	get func(C) float64,
	set func(*C, float64),
	opts ...textinput.Option,
) input.Input[string] {
	return buildTextInput(
		l, title, strconv.FormatFloat(get(*cfg), 'f', -1, 64),
		func(v string) error {
			n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return err
			}
			tmp := *cfg
			set(&tmp, n)
			if e, ok := tmp.Validate()[errKey]; ok {
				return e
			}
			return nil
		},
		func(v string) {
			n, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
			set(cfg, n)
		},
		opts...,
	)
}

// FingerprintField builds the uTLS fingerprint select shared by the
// vaydns/dnstt forms.
func FingerprintField[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	get func(C) string,
	set func(*C, string),
) input.Input[string] {
	opts := make([]huh.Option[string], 0, len(config.FingerprintLabels()))
	for _, label := range config.FingerprintLabels() {
		opts = append(opts, huh.NewOption(label, label))
	}

	return selectinput.New(
		l, "Select TLS fingerprint",
		selectinput.WithValue(get(*cfg)),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(opts...),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			set(&tmp, v)
			if e, ok := tmp.Validate()["fingerprint"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v string) tea.Cmd {
			set(cfg, v)
			return nil
		}),
	)
}

// EncMethodField builds the data-encryption-method select shared by the
// masterdns/stormdns forms.
func EncMethodField[C TunnelConfig](
	l *layout.Layout,
	cfg *C,
	title string,
	get func(C) dns.EncMethod,
	set func(*C, dns.EncMethod),
) input.Input[dns.EncMethod] {
	return selectinput.New(
		l, title,
		selectinput.WithValue(get(*cfg)),
		selectinput.WithFocus[dns.EncMethod](),
		selectinput.WithOptions[dns.EncMethod](
			huh.NewOption(dns.EncNone.String(), dns.EncNone),
			huh.NewOption(dns.EncXOR.String(), dns.EncXOR),
			huh.NewOption(dns.EncChaCha20.String(), dns.EncChaCha20),
			huh.NewOption(dns.EncAES128GCM.String(), dns.EncAES128GCM),
			huh.NewOption(dns.EncAES192GCM.String(), dns.EncAES192GCM),
			huh.NewOption(dns.EncAES256GCM.String(), dns.EncAES256GCM),
		),
		selectinput.WithValidation(func(v dns.EncMethod) error {
			tmp := *cfg
			set(&tmp, v)
			if e, ok := tmp.Validate()["enc_method"]; ok {
				return e
			}
			return nil
		}),
		selectinput.WithOnSubmit(func(v dns.EncMethod) tea.Cmd {
			set(cfg, v)
			return nil
		}),
	)
}

// buildTextInput is the common single-line input constructor. The validation
// and submit closures receive the raw string and are responsible for parse,
// config validation and set.
func buildTextInput(
	l *layout.Layout,
	title, value string,
	validate func(string) error,
	submit func(string),
	opts ...textinput.Option,
) input.Input[string] {
	o := []textinput.Option{
		textinput.WithValue(value),
		textinput.WithFocus(),
		textinput.WithValidation(validate),
		textinput.WithOnSubmit(func(v string) tea.Cmd {
			submit(v)
			return nil
		}),
	}
	return textinput.New(l, title, append(o, opts...)...)
}

// buildUintInput is the common numeric input constructor with overflow-safe
// parsing of the given bit size.
func buildUintInput(
	l *layout.Layout,
	title string,
	bitSize int,
	value uint64,
	validate func(uint64) error,
	submit func(uint64),
	opts ...textinput.Option,
) input.Input[string] {
	return buildTextInput(
		l, title, strconv.FormatUint(value, 10),
		func(v string) error {
			n, err := strconv.ParseUint(strings.TrimSpace(v), 10, bitSize)
			if err != nil {
				return err
			}
			return validate(n)
		},
		func(v string) {
			n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, bitSize)
			submit(n)
		},
		opts...,
	)
}
