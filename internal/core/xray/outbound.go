package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/logger"
)

// ── 1. Placeholder Replacement Logic ─────────────────────────────────────────

// replacePlaceholders recursively walks a decoded JSON structure and replaces
// any string whose value exactly matches a key from the replacements map.
func replacePlaceholders(data any, replacements map[string]string) any {
	switch v := data.(type) {
	case map[string]any:
		for key, val := range v {
			v[key] = replacePlaceholders(val, replacements)
		}
		return v

	case []any:
		for i, val := range v {
			v[i] = replacePlaceholders(val, replacements)
		}
		return v

	case string:
		if newVal, ok := replacements[v]; ok {
			return newVal
		}
		return v

	default:
		return v
	}
}

// applyOutboundTemplate loads an outbound template, swaps placeholders
// with runtime values, and returns the modified, decoded structure.
func applyOutboundTemplate(templatePath, ip string) (any, error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP target address: %s", ip)
	}

	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse outbound template JSON: %w", err)
	}

	return replacePlaceholders(parsed, map[string]string{
		addressPlaceholder: ip,
	}), nil
}

// ── 2. Template Saving Logic ─────────────────────────────────────────────────

// SaveOutboundFromFile validates and stores a new outbound template from a disk source file.
func SaveOutboundFromFile(src, name string) (*XrayOutboundsFile, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source outbound file does not exist: %s", src)
		}

		return nil, fmt.Errorf("cannot access source file %s: %w", src, err)
	}

	if srcInfo.IsDir() {
		return nil, fmt.Errorf("source path is a directory, expected file: %s", src)
	}

	name = normalizeTemplateName(name)
	dst := filepath.Join(templateDir(), name)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("outbound template %q already exists", name)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("cannot read source file %s: %w", src, err)
	}

	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON in outbound template: %w", err)
	}

	if !containsAddressPlaceholder(jsonData) {
		return nil, fmt.Errorf("outbound template missing required placeholder: %q", addressPlaceholder)
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("failed to save outbound template: %w", err)
	}

	if err := ValidateOutbound(name); err != nil {
		defer func() {
			if err := os.Remove(dst); err != nil {
				logger.CoreError("failed to remove config file: %v", err)
			}
		}()

		return nil, fmt.Errorf("outbound validation failed: %w", err)
	}

	return loadOutboundFileMetadata(dst)
}

// SaveOutboundFromLink parses an outbound sharing URL link, converts it to an
// address-templated JSON file, validates it, and saves it to disk.
func SaveOutboundFromLink(link, name string) (*XrayOutboundsFile, error) {
	parsed, err := ParseLink(link)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sharing link: %w", err)
	}

	data, err := json.MarshalIndent(parsed.Outbound, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format outbound link JSON: %w", err)
	}

	name = normalizeTemplateName(name)
	dst := filepath.Join(templateDir(), name)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("outbound template %q already exists", name)
	}

	var jsonValidationAny any
	if err := json.Unmarshal(data, &jsonValidationAny); err != nil {
		return nil, fmt.Errorf("serialized validation fallback failed: %w", err)
	}

	if !containsAddressPlaceholder(jsonValidationAny) {
		return nil, fmt.Errorf("link template missing required placeholder: %q", addressPlaceholder)
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("failed to save outbound template from link: %w", err)
	}

	if err := ValidateOutbound(name); err != nil {
		if err := os.Remove(dst); err != nil {
			logger.CoreError("failed to remove outbound file: %v", err)
		}

		return nil, fmt.Errorf("outbound validation failed: %w", err)
	}

	return loadOutboundFileMetadata(dst)
}

// ── 3. Template Operations & Retrieval ───────────────────────────────────────

// GetOutboundTemplateByName finds an outbound template by name, automatically handling extensions.
func GetOutboundTemplateByName(name string) (*XrayOutboundsFile, error) {
	name = normalizeTemplateName(name)

	path := filepath.Join(templateDir(), name)

	return loadOutboundFileMetadata(path)
}

// ListOutboundTemplates returns a list of all existing template metadata objects saved on disk.
func ListOutboundTemplates() ([]XrayOutboundsFile, error) {
	filter := func(name string, info os.FileInfo) bool {
		return !info.IsDir() && strings.HasSuffix(name, ".json")
	}

	files, err := fileutil.ListFiles(templateDir(), filter)
	if err != nil {
		return nil, err
	}

	templates := make([]XrayOutboundsFile, 0, len(files))

	for _, f := range files {
		meta, err := loadOutboundFileMetadata(f.Path)
		if err != nil {
			logger.CoreError("failed to parse outbound metadata: %v", err)
			continue
		}

		templates = append(templates, *meta)
	}

	return templates, nil
}

// RenameOutboundTemplate atomically updates an outbound template filename configuration.
func RenameOutboundTemplate(oldName, newName string) (*XrayOutboundsFile, error) {
	oldFile, err := GetOutboundTemplateByName(oldName)
	if err != nil {
		return nil, err
	}

	newName = normalizeTemplateName(newName)

	dst := filepath.Join(templateDir(), newName)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("cannot rename: destination template %q already exists", newName)
	}

	if err := os.Rename(oldFile.Path, dst); err != nil {
		return nil, fmt.Errorf("failed to execute rename command: %w", err)
	}

	return loadOutboundFileMetadata(dst)
}

// ── 4. Validation & Internal Helpers ──────────────────────────────────────────

// ValidateOutbound generates a validation testbed frame to run Xray logic testing.
func ValidateOutbound(outbound string) error {
	configPath, err := GenerateConfig(outbound, "127.0.0.1", 40443)
	if err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(configPath); err != nil {
			logger.CoreError("failed to remove config file: %v", err)
		}
	}()

	return ValidateConfig(context.Background(), configPath)
}

// loadOutboundFileMetadata reads an outbound template and extracts
// metadata used by bgscan.
//
// Missing fields are ignored and keep their Go zero values.
func loadOutboundFileMetadata(path string) (*XrayOutboundsFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	result := &XrayOutboundsFile{
		Name:        fileutil.StripExt(filepath.Base(path)),
		Path:        path,
		CreatedTime: info.ModTime(),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var outbound map[string]any
	if err := json.Unmarshal(data, &outbound); err != nil {
		return nil, err
	}

	if protocol, ok := outbound["protocol"].(string); ok {
		result.Protocol = protocol
	}

	streamSettings, ok := outbound["streamSettings"].(map[string]any)
	if !ok {
		return result, nil
	}

	if network, ok := streamSettings["network"].(string); ok {
		result.Network = network
	}

	if security, ok := streamSettings["security"].(string); ok {
		result.UseTLS = security == "tls"
	}

	return result, nil
}

// containsAddressPlaceholder recursively checks if the address placeholder string token exists.
func containsAddressPlaceholder(v any) bool {
	switch val := v.(type) {
	case map[string]any:
		for k, v2 := range val {
			if k == "address" {
				if s, ok := v2.(string); ok && s == addressPlaceholder {
					return true
				}
			}

			if containsAddressPlaceholder(v2) {
				return true
			}
		}

	case []any:
		return slices.ContainsFunc(val, containsAddressPlaceholder)
	}

	return false
}

// normalizeTemplateName trims trailing symbols and guarantees a lowercase json extension suffix.
func normalizeTemplateName(name string) string {
	if ext := filepath.Ext(name); ext != ".json" {
		name = strings.TrimSuffix(name, ext) + ".json"
	}

	return name
}
