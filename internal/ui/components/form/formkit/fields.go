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
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func ConfigNameField(env *ui.Env, name string, set func(string)) input.Input[string] {
	return textinput.New(
		env, "Enter config name",
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

// StringField binds a text input to a string config field, validating
// through a copy of the whole config keyed by errKey.
func StringField[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	title string, errKey dns.Field,
	get func(C) string,
	set func(*C, string),
) input.Input[string] {
	return buildTextInput(
		env, title, get(*cfg),
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

// SecretField binds a textarea to a secret config field (keys), validating
// through a copy.
func SecretField[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	title string, errKey dns.Field,
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
	return textarea.New(env, title, append(o, opts...)...)
}

// Uint8Field binds a numeric input to a uint8 field, parsing with bit size
// 8 so overflow fails validation instead of wrapping.
func Uint8Field[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	title string, errKey dns.Field,
	get func(C) uint8,
	set func(*C, uint8),
	opts ...textinput.Option,
) input.Input[string] {
	return buildUintInput(env, title, 8, uint64(get(*cfg)), func(n uint64) error {
		tmp := *cfg
		set(&tmp, uint8(n))
		if e, ok := tmp.Validate()[errKey]; ok {
			return e
		}
		return nil
	}, func(n uint64) { set(cfg, uint8(n)) }, opts...)
}

// Uint16Field binds a numeric input to a uint16 field, parsing with bit
// size 16 so overflow fails validation instead of wrapping.
func Uint16Field[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	title string, errKey dns.Field,
	get func(C) uint16,
	set func(*C, uint16),
	opts ...textinput.Option,
) input.Input[string] {
	return buildUintInput(env, title, 16, uint64(get(*cfg)), func(n uint64) error {
		tmp := *cfg
		set(&tmp, uint16(n))
		if e, ok := tmp.Validate()[errKey]; ok {
			return e
		}
		return nil
	}, func(n uint64) { set(cfg, uint16(n)) }, opts...)
}

func FloatField[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	title string, errKey dns.Field,
	get func(C) float64,
	set func(*C, float64),
	opts ...textinput.Option,
) input.Input[string] {
	return buildTextInput(
		env, title, strconv.FormatFloat(get(*cfg), 'f', -1, 64),
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

// FingerprintField is the uTLS fingerprint select shared by vaydns/dnstt.
func FingerprintField[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	get func(C) string,
	set func(*C, string),
) input.Input[string] {
	opts := make([]huh.Option[string], 0, len(config.FingerprintLabels()))
	for _, label := range config.FingerprintLabels() {
		opts = append(opts, huh.NewOption(label, label))
	}

	return selectinput.New(
		env, "Select TLS fingerprint",
		selectinput.WithValue(get(*cfg)),
		selectinput.WithFocus[string](),
		selectinput.WithOptions(opts...),
		selectinput.WithValidation(func(v string) error {
			tmp := *cfg
			set(&tmp, v)
			if e, ok := tmp.Validate()[dns.FieldFingerprint]; ok {
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

// EncMethodField is the encryption-method select shared by masterdns/stormdns.
func EncMethodField[C TunnelConfig](
	env *ui.Env,
	cfg *C,
	title string,
	get func(C) dns.EncMethod,
	set func(*C, dns.EncMethod),
) input.Input[dns.EncMethod] {
	return selectinput.New(
		env, title,
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
			if e, ok := tmp.Validate()[dns.FieldEncMethod]; ok {
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

// buildTextInput is the shared single-line constructor; validate/submit
// receive the raw string and handle parsing, config validation and set.
func buildTextInput(
	env *ui.Env,
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
	return textinput.New(env, title, append(o, opts...)...)
}

// buildUintInput is the shared numeric constructor, parsing with the given
// bit size so overflow fails validation instead of wrapping.
func buildUintInput(
	env *ui.Env,
	title string,
	bitSize int,
	value uint64,
	validate func(uint64) error,
	submit func(uint64),
	opts ...textinput.Option,
) input.Input[string] {
	return buildTextInput(
		env, title, strconv.FormatUint(value, 10),
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
