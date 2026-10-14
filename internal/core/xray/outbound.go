package xray

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/logger"
	libxray "github.com/xtls/libxray/share"
)

// isProxyProtocol reports whether the protocol dials a real remote server.
func isProxyProtocol(p Protocol) bool {
	switch p {
	case "", ProtocolFreedom, ProtocolBlackhole, ProtocolDNS, ProtocolLoopback:
		return false
	default:
		return true
	}
}

type addressLocation struct {
	path    []string
	portKey string
}

// outboundAddressLocations returns every remote server address inside a proxy outbound.
func outboundAddressLocations(outbound map[string]any) ([]addressLocation, error) {
	protocol, _ := outbound["protocol"].(string)
	if !isProxyProtocol(Protocol(protocol)) {
		return nil, fmt.Errorf("outbound %q is not a proxy", protocol)
	}

	settings, ok := outbound["settings"].(map[string]any)
	if !ok {
		return nil, errors.New("outbound has no settings")
	}

	if Protocol(protocol) == ProtocolWireGuard {
		peers, _ := settings["peers"].([]any)
		locs := make([]addressLocation, 0, len(peers))

		for i, peer := range peers {
			node, ok := peer.(map[string]any)
			if !ok {
				continue
			}
			if _, ok := node["endpoint"].(string); ok {
				locs = append(locs, addressLocation{
					path: []string{"settings", "peers", strconv.Itoa(i), "endpoint"},
				})
			}
		}

		if len(locs) == 0 {
			return nil, errors.New("wireguard outbound has no peer endpoint")
		}
		return locs, nil
	}

	var locs []addressLocation

	if _, ok := settings["address"].(string); ok {
		locs = append(locs, addressLocation{
			path:    []string{"settings", "address"},
			portKey: "port",
		})
	}

	for _, key := range []string{"vnext", "servers"} {
		entries, _ := settings[key].([]any)
		for i, entry := range entries {
			node, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if _, ok := node["address"].(string); ok {
				locs = append(locs, addressLocation{
					path:    []string{"settings", key, strconv.Itoa(i), "address"},
					portKey: "port",
				})
			}
		}
	}

	if len(locs) == 0 {
		return nil, errors.New("outbound has no replaceable remote address")
	}

	return locs, nil
}

func requireReplaceableAddress(outbound map[string]any) error {
	_, err := outboundAddressLocations(outbound)
	return err
}

func setOutboundAddress(outbound map[string]any, ip netip.Addr, port uint16) error {
	if !ip.IsValid() {
		return fmt.Errorf("invalid IP target address: %s", ip)
	}

	locs, err := outboundAddressLocations(outbound)
	if err != nil {
		return err
	}

	for _, loc := range locs {
		if err := setAddressAt(outbound, loc, ip, port); err != nil {
			return err
		}
	}

	return nil
}

// setAddressAt applies a scanned IP to one address location
func setAddressAt(outbound map[string]any, loc addressLocation, ip netip.Addr, port uint16) error {
	node := any(outbound)

	for _, key := range loc.path[:len(loc.path)-1] {
		next, ok := descend(node, key)
		if !ok {
			return errors.New("malformed outbound address")
		}
		node = next
	}

	parent, ok := node.(map[string]any)
	if !ok {
		return errors.New("malformed outbound address")
	}

	last := loc.path[len(loc.path)-1]
	current, ok := parent[last].(string)
	if !ok {
		return errors.New("malformed outbound address")
	}

	if loc.portKey == "" {
		// WireGuard peers dial "host:port" endpoints.
		if port != 0 {
			parent[last] = net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
			return nil
		}
		if _, embeddedPort, err := net.SplitHostPort(current); err == nil {
			parent[last] = net.JoinHostPort(ip.String(), embeddedPort)
			return nil
		}
		parent[last] = ip.String()
		return nil
	}

	parent[last] = ip.String()
	if port != 0 {
		parent[loc.portKey] = port
	}

	return nil
}

func descend(node any, key string) (any, bool) {
	switch n := node.(type) {
	case map[string]any:
		v, ok := n[key]
		return v, ok
	case []any:
		i, err := strconv.Atoi(key)
		if err != nil || i < 0 || i >= len(n) {
			return nil, false
		}
		return n[i], true
	default:
		return nil, false
	}
}

func applyOutboundTemplate(templatePath string, ip netip.Addr, port uint16) (any, error) {
	if !ip.IsValid() {
		return nil, fmt.Errorf("invalid IP target address: %s", ip)
	}

	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("read template file: %w", err)
	}

	var outbound map[string]any
	if err := json.Unmarshal(raw, &outbound); err != nil {
		return nil, fmt.Errorf("parse outbound template JSON: %w", err)
	}

	if err := setOutboundAddress(outbound, ip, port); err != nil {
		return nil, err
	}

	return outbound, nil
}

func SaveOutboundFromFile(log *logger.Logger, src, name string) (*OutboundFile, error) {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source outbound file does not exist: %s", src)
		}
		return nil, fmt.Errorf("cannot access source file %s: %w", src, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("source path is a directory, expected file: %s", src)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("cannot read source file %s: %w", src, err)
	}

	data, err = cleanJSONC(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON in outbound template: %w", err)
	}

	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON in outbound template: %w", err)
	}

	outbound, err := extractMainOutbound(jsonData)
	if err != nil {
		return nil, err
	}

	return saveOutboundTemplate(log, name, outbound)
}

// ParseLink converts a sharing URL to the main outbound it describes.
func ParseLink(link string) (*Outbound, error) {
	parsed, err := libxray.ConvertShareLinksToXrayJson(link, "")
	if err != nil {
		return nil, fmt.Errorf("parse sharing link: %w", err)
	}

	var jsonData any
	if err := json.Unmarshal(parsed, &jsonData); err != nil {
		return nil, fmt.Errorf("decode outbound JSON: %w", err)
	}

	outbound, err := extractMainOutbound(jsonData)
	if err != nil {
		return nil, fmt.Errorf("select main outbound: %w", err)
	}

	return outbound, nil
}

func SaveOutboundFromLink(log *logger.Logger, link, name string) (*OutboundFile, error) {
	outbound, err := ParseLink(link)
	if err != nil {
		return nil, err
	}
	return saveOutboundTemplate(log, name, outbound)
}

func saveOutboundTemplate(log *logger.Logger, name string, outbound *Outbound) (*OutboundFile, error) {
	if outbound == nil {
		return nil, errors.New("outbound is nil")
	}

	name = normalizeTemplateName(name)
	dst := filepath.Join(templateDir(), name)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("outbound template %q already exists", name)
	}

	outbound.normalizeWSALPN()

	data, err := json.MarshalIndent(outbound, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode outbound template: %w", err)
	}

	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		return nil, fmt.Errorf("decode outbound template: %w", err)
	}

	if err := requireReplaceableAddress(check); err != nil {
		return nil, err
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("save outbound template: %w", err)
	}

	if err := ValidateOutbound(name); err != nil {
		if removeErr := os.Remove(dst); removeErr != nil && log != nil {
			log.Error("failed to remove invalid outbound template: %v", removeErr)
		}
		return nil, fmt.Errorf("outbound validation failed: %w", err)
	}

	return loadOutboundFileMetadata(dst)
}

func GetOutboundTemplateByName(name string) (*OutboundFile, error) {
	path := filepath.Join(templateDir(), normalizeTemplateName(name))
	return loadOutboundFileMetadata(path)
}

func ListOutboundTemplates(log *logger.Logger) ([]OutboundFile, error) {
	files, err := fileutil.ListFiles(templateDir(), func(name string, info os.FileInfo) bool {
		return !info.IsDir() && filepath.Ext(name) == ".json"
	})
	if err != nil {
		return nil, err
	}

	templates := make([]OutboundFile, 0, len(files))
	for _, f := range files {
		meta, err := loadOutboundFileMetadata(f.Path)
		if err != nil {
			if log != nil {
				log.Error("failed to parse outbound metadata: %v", err)
			}
			continue
		}
		templates = append(templates, *meta)
	}

	return templates, nil
}

func RenameOutboundTemplate(oldName, newName string) (*OutboundFile, error) {
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
		return nil, fmt.Errorf("rename outbound template: %w", err)
	}

	return loadOutboundFileMetadata(dst)
}

// ValidateOutbound checks a saved template by building a throwaway loopback config.
func ValidateOutbound(outbound string) error {
	svc := NewXrayService()
	loopback := netip.AddrFrom4([4]byte{127, 0, 0, 1})

	cfg, err := svc.GenerateConfig(outbound, loopback, 40443, 0)
	if err != nil {
		return err
	}

	return svc.ValidateConfig(context.Background(), cfg)
}

func loadOutboundFileMetadata(path string) (*OutboundFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var outbound map[string]any
	if err := json.Unmarshal(data, &outbound); err != nil {
		return nil, err
	}

	meta := &OutboundFile{
		Name:        fileutil.StripExt(filepath.Base(path)),
		Path:        path,
		CreatedTime: info.ModTime(),
	}

	if protocol, ok := outbound["protocol"].(string); ok {
		meta.Protocol = Protocol(protocol)
	}

	streamSettings, _ := outbound["streamSettings"].(map[string]any)
	if streamSettings == nil {
		return meta, nil
	}

	if network, ok := streamSettings["network"].(string); ok {
		meta.Network = Network(network)
	}
	if security, ok := streamSettings["security"].(string); ok {
		meta.Security = Security(security)
	}

	return meta, nil
}

// normalizeTemplateName guarantees a plain file name with a lowercase .json extension.
func normalizeTemplateName(name string) string {
	name = filepath.Base(name)
	return strings.TrimSuffix(name, filepath.Ext(name)) + ".json"
}

// extractMainOutbound supports:
// - a bare outbound object
// - an array of outbounds
// - a full config object with an "outbounds" array
func extractMainOutbound(data any) (*Outbound, error) {
	switch v := data.(type) {
	case map[string]any:
		if raw, ok := v["outbounds"]; ok {
			arr, ok := raw.([]any)
			if !ok {
				return nil, errors.New(`"outbounds" is not an array`)
			}
			outbounds, err := decodeOutbounds(arr)
			if err != nil {
				return nil, err
			}
			return pickMain(outbounds)
		}

		if _, ok := v["protocol"]; ok {
			return decodeOutbound(v)
		}

		return nil, errors.New("no outbound found in JSON")

	case []any:
		outbounds, err := decodeOutbounds(v)
		if err != nil {
			return nil, err
		}
		return pickMain(outbounds)

	default:
		return nil, errors.New("unsupported JSON shape")
	}
}

func decodeOutbound(v any) (*Outbound, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var outbound Outbound
	if err := json.Unmarshal(raw, &outbound); err != nil {
		return nil, err
	}
	if outbound.Protocol == "" {
		return nil, errors.New("outbound has no protocol")
	}

	return &outbound, nil
}

func decodeOutbounds(v any) ([]Outbound, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var outbounds []Outbound
	if err := json.Unmarshal(raw, &outbounds); err != nil {
		return nil, err
	}

	return outbounds, nil
}

// pickMain selects the first proxy outbound, otherwise one tagged "proxy".
func pickMain(outbounds []Outbound) (*Outbound, error) {
	for i := range outbounds {
		if outbounds[i].IsProxy() {
			return &outbounds[i], nil
		}
	}

	for i := range outbounds {
		if outbounds[i].Tag == "proxy" {
			return &outbounds[i], nil
		}
	}

	if len(outbounds) == 0 {
		return nil, errors.New("outbounds is empty")
	}

	return nil, errors.New("no proxy outbound found")
}
