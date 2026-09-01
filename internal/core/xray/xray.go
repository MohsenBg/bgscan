package xray

import (
	"fmt"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// GenerateConfig builds a complete Xray configuration from a template
// and writes it to disk.
//
// The function performs the following steps:
//
//  1. Loads the specified outbound template.
//  2. Replaces template placeholders (e.g. $ADDRESS) with the target IP.
//  3. Injects a scanner-generated inbound proxy.
//  4. Writes the final configuration file to disk.
//
// Each generated config contains:
//
//   - a local SOCKS inbound used by the scanner probes
//   - a single outbound derived from the selected template
//
// The returned value is the path to the generated configuration file,
// which can then be passed to an Xray process.
func GenerateConfig(outboundName string, ip netip.Addr, port uint16) (string, error) {
	if !ip.IsValid() {
		return "", fmt.Errorf("invalid IP: %s", ip)
	}

	// Get template file path
	template, err := GetOutboundTemplateByName(outboundName)
	if err != nil {
		return "", err
	}

	outbound, err := applyOutboundTemplate(template.Path, ip)
	if err != nil {
		return "", err
	}

	// Build full config
	config := XrayConfig{
		Inbounds:  []Inbound{getInbound(port)},
		Outbounds: []any{outbound},
	}

	// Generate output path
	outputPath := getNewXrayConfigName(ip)

	// Write config file
	if err := fileutil.WriteJSONFile(outputPath, config); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return outputPath, nil
}

// getNewXrayConfigName returns the file path for a generated
// Xray configuration associated with the given IP address.

func getNewXrayConfigName(ip netip.Addr) string {
	filename := fmt.Sprintf("%s.json", strings.ReplaceAll(ip.String(), ":", "_"))
	return filepath.Join(configDir(), filename)
}

func getAssetsPath(parts ...string) string {
	base, err := fileutil.BasePath()
	if err != nil {
		return filepath.Join(parts...)
	}

	return filepath.Join(append([]string{base}, parts...)...)
}

func configDir() string {
	return getAssetsPath("assets", "xray", "configs")
}

func templateDir() string {
	return getAssetsPath("assets", "xray", "outbounds")
}

func RemoveTmpCfg() error {
	dir := configDir()
	err := fileutil.EnsureDir(dir)
	if err != nil {
		return err
	}

	_, err = fileutil.ListFiles(dir, func(name string, info os.FileInfo) bool {
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".json") {
			return false
		}
		return os.Remove(path.Join(dir, name)) == nil
	})

	return err
}
