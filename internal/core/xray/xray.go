package xray

import (
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bgscan/internal/core/fileutil"
)

// GenerateConfig builds a complete Xray configuration from a template
// and writes it to disk.
//
// The function performs the following steps:
//
//  1. Validates the provided IP address.
//  2. Loads the specified outbound template.
//  3. Replaces template placeholders (e.g. $ADDRESS).
//  4. Injects a scanner-generated inbound proxy.
//  5. Writes the final configuration file to disk.
//
// Each generated config contains:
//
//   - a local SOCKS inbound used by the scanner probes
//   - a single outbound derived from the selected template
//
// The returned value is the path to the generated configuration file,
// which can then be passed to an Xray process.
func GenerateConfig(outboundName, ip string, port uint16) (string, error) {
	// Validate IP
	if net.ParseIP(ip) == nil {
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
//
// Each target IP produces a dedicated configuration file stored
// inside the configPath directory.
func getNewXrayConfigName(ip string) string {
	filename := fmt.Sprintf("%s.json", ip)
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
