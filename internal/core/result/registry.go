package result

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bgscan/internal/core/fileutil"
)

// Result directories used by different scan types.
// Each directory stores CSV files produced by a specific scan engine.
const (
	ICMPResultDir       = "result/icmp/"
	TCPResultDir        = "result/tcp/"
	HTTPResultDir       = "result/http/"
	XRAYResultDir       = "result/xray/"
	ResolveResultDir    = "result/resolve/"
	DNSTTResultDir      = "result/dnstt/"
	SlipStreamResultDir = "result/slipstream/"
)

// GetResultTypeFiles returns metadata for all result CSV files matching the given scan type.
//
// For performance reasons, this function only reads filesystem metadata
// and does NOT count the number of IPs stored inside each file.
// As a result, the IPCount field is always set to -1.
func GetResultTypeFiles(searchType ResultType) ([]ResultFile, error) {
	dirs := resolveResultDirs(searchType)
	if len(dirs) == 0 {
		return nil, nil
	}

	var results []ResultFile

	for _, d := range dirs {
		entries, err := os.ReadDir(d.dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasSuffix(strings.ToLower(name), ".csv") {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			results = append(results, ResultFile{
				Name:        fileutil.StripExt(name),
				SizeBytes:   info.Size(),
				CreatedTime: info.ModTime(),
				Type:        d.rType,
				Path:        filepath.Join(d.dir, name),
				IPCount:     -1,
			})
		}
	}

	return results, nil
}

// GetResultFiles returns all metadata files across every scan engine type.
// Serves as a backward-compatible shortcut for background processors.
func GetResultFiles() ([]ResultFile, error) {
	return GetResultTypeFiles(ResultAll)
}

// ReadResultFileIPs reads filesystem statistical info for a single file path
// and returns an initialized ResultFile schema.
//
// Like GetResultTypeFiles, it sets the IPCount to -1 to avoid scanning
// file lines synchronously. Use CountIPsInFile to compute the exact total.
func ReadResultFileIPs(path string) (ResultFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ResultFile{}, fmt.Errorf("cannot read result file: %w", err)
	}

	return ResultFile{
		Name:        fileutil.StripExt(info.Name()),
		SizeBytes:   info.Size(),
		CreatedTime: info.ModTime(),
		Type:        ResultTypeFromPath(path),
		Path:        path,
		IPCount:     -1,
	}, nil
}

// NormalizeResultFileName ensures a clean filename wrapper ends with `.csv`
// while safely stripping any internal absolute or relative path segments.
func NormalizeResultFileName(name string) string {
	baseName := filepath.Base(name)
	if !fileutil.HasExt(baseName, ".csv") {
		return baseName + ".csv"
	}
	return baseName
}

// ResultTypeFromPath infers the corresponding ResultType identifier dynamically from a filesystem path context.
//
// It sanitizes platform-specific slashes and evaluates input bounds against existing configurations,
// guaranteeing single-source-of-truth accuracy without relying on hardcoded strings.
func ResultTypeFromPath(path string) ResultType {
	cleanedPath := filepath.Clean(path)

	// Dynamically evaluate against registered engines
	for _, d := range resolveResultDirs(ResultAll) {
		if strings.Contains(cleanedPath, filepath.Clean(d.dir)) {
			return d.rType
		}
	}

	return ResultICMP
}

// CountIPsInFile calculates the total number of IP entries recorded inside a result file
// by invoking the core tracking package's file counting mechanics.
func CountIPsInFile(file ResultFile) (int64, error) {
	return Count(file.Path)
}

// BuildResultFilePath generates a clean, standardized, timestamp-suffixed path for a
// new scan output target inside an authorized output directory block.
func BuildResultFilePath(targetDir, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("prefix cannot be empty")
	}

	cleanTarget := filepath.Clean(targetDir)

	// Validate targetDir against mapped registry to maintain secure filesystem boundaries
	var isValid bool
	for _, d := range resolveResultDirs(ResultAll) {
		if filepath.Clean(d.dir) == cleanTarget {
			isValid = true
			break
		}
	}

	if !isValid {
		return "", fmt.Errorf("invalid or untracked result directory: %s", targetDir)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s%s.csv", prefix, timestamp)

	return filepath.Join(cleanTarget, filename), nil
}

// -----------------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------------

type resultDir struct {
	dir   string
	rType ResultType
}

func resolveResultDirs(searchType ResultType) []resultDir {
	all := []resultDir{
		{ICMPResultDir, ResultICMP},
		{TCPResultDir, ResultTCP},
		{HTTPResultDir, ResultHTTP},
		{XRAYResultDir, ResultXRAY},
		{DNSTTResultDir, ResultDNSTT},
		{SlipStreamResultDir, ResultSLIPSTREAM},
		{ResolveResultDir, ResultRESOLVE},
	}

	if searchType == ResultAll {
		return all
	}

	for _, d := range all {
		if d.rType == searchType {
			return []resultDir{d}
		}
	}

	return nil
}
