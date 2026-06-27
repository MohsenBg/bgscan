package xray

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

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

	modified := replacePlaceholders(parsed, map[string]string{
		addressPlaceholder: ip,
	})

	return modified, nil
}

// ── 2. Template Saving Logic ─────────────────────────────────────────────────

// SaveOutboundFromFile validates and stores a new outbound template from a disk source file.
func SaveOutboundFromFile(src, name string) (*XrayOutboundsFile, error) {
	// Check source details
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
	dst := filepath.Join(templatePath, name)

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
		return nil, fmt.Errorf("outbound template missing required placeholder: \"address\": %q", addressPlaceholder)
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

	return &XrayOutboundsFile{
		Name:        fileutil.StripExt(name),
		Path:        dst,
		CreatedTime: time.Now(),
	}, nil
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
	dst := filepath.Join(templatePath, name)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("outbound template %q already exists", name)
	}

	var jsonValidationAny any
	if err := json.Unmarshal(data, &jsonValidationAny); err != nil {
		return nil, fmt.Errorf("serialized validation fallback failed: %w", err)
	}
	if !containsAddressPlaceholder(jsonValidationAny) {
		return nil, fmt.Errorf("link template missing required placeholder: \"address\": %q", addressPlaceholder)
	}
	logger.DebugDump("jsonValidationAny", jsonValidationAny)

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("failed to save outbound template from link: %w", err)
	}

	if err := ValidateOutbound(name); err != nil {
		if err := os.Remove(dst); err != nil {
			logger.CoreError("failed to remove  outbound file: %v", err)
		}

		return nil, fmt.Errorf("outbound validation failed: %w", err)
	}

	return &XrayOutboundsFile{
		Name:        fileutil.StripExt(name),
		Path:        dst,
		CreatedTime: time.Now(),
	}, nil
}

// ── 3. Template Operations & Retrieval ───────────────────────────────────────

// GetOutboundTemplateByName finds an outbound template by name, automatically handling extensions.
func GetOutboundTemplateByName(name string) (*XrayOutboundsFile, error) {
	name = normalizeTemplateName(name)
	path := filepath.Join(templatePath, name)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read outbound template %s: %w", path, err)
	}

	return &XrayOutboundsFile{
		Name:        fileutil.StripExt(name),
		CreatedTime: info.ModTime(),
		Path:        path,
	}, nil
}

// ListOutboundTemplates returns a list of all existing template metadata objects saved on disk.
func ListOutboundTemplates() ([]XrayOutboundsFile, error) {
	filter := func(name string, info os.FileInfo) bool {
		return !info.IsDir() && strings.HasSuffix(name, ".json")
	}

	files, err := fileutil.ListFiles(templatePath, filter)
	if err != nil {
		return nil, err
	}

	templates := make([]XrayOutboundsFile, 0, len(files))
	for _, f := range files {
		templates = append(templates, XrayOutboundsFile{
			Name:        fileutil.StripExt(f.Name),
			Path:        f.Path,
			CreatedTime: f.Info.ModTime(),
		})
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
	dst := filepath.Join(templatePath, newName)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("cannot rename: destination template %q already exists", newName)
	}

	if err := os.Rename(oldFile.Path, dst); err != nil {
		return nil, fmt.Errorf("failed to execute rename command: %w", err)
	}

	return &XrayOutboundsFile{
		Name:        fileutil.StripExt(newName),
		Path:        dst,
		CreatedTime: oldFile.CreatedTime,
	}, nil
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

	return ValidateConfig(configPath)
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
