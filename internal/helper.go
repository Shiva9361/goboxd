package internal

import (
	"bytes"
	"cmp"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func buildAllowlist(allowlist []string) FlagLookup {
	exactMatches := make(map[string]struct{})
	var patterns []string

	for _, rule := range allowlist {
		if strings.ContainsAny(rule, "*?[") {
			patterns = append(patterns, rule)
		} else {
			exactMatches[rule] = struct{}{} // this is apparently efficient set in GO
		}
	}
	return FlagLookup{ExactMatches: exactMatches, patterns: patterns}
}

func sanitizeFlags(flags []string, lookup FlagLookup) []string {
	filteredFlags := make([]string, 0, len(flags))

	for _, flag := range flags {
		if _, exists := lookup.ExactMatches[flag]; exists {
			filteredFlags = append(filteredFlags, flag)
			continue
		}
		isAllowed := false
		for _, pattern := range lookup.patterns {
			if matched, _ := path.Match(pattern, flag); matched {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			filteredFlags = append(filteredFlags, flag)
		}
	}

	return filteredFlags
}

func prepArgs(template []string, flags []string, source string, artifact string) []string {
	args := make([]string, 0, len(template)+len(flags))

	for _, arg := range template {
		if arg == "{{flags}}" {
			args = append(args, flags...)
			continue
		}
		arg = strings.ReplaceAll(arg, "{{source}}", source)
		arg = strings.ReplaceAll(arg, "{{artifact}}", artifact)
		args = append(args, arg)
	}
	return args
}

func clampResource[T cmp.Ordered](defaultValue T, requestedValue T) T {
	var zero T
	if requestedValue > zero && requestedValue < defaultValue {
		return requestedValue
	}
	return defaultValue
}

func sanitizeFilename(filename string) string {
	return filepath.Base(filename)
}

func sanitizeLimits(defaultLimits Resource, requestedLimits Resource) Resource {
	limits := Resource{
		WallTime:   clampResource(defaultLimits.WallTime, requestedLimits.WallTime), // cool way to write instead of if else
		Memory:     clampResource(defaultLimits.Memory, requestedLimits.Memory),
		MaxProcess: clampResource(defaultLimits.MaxProcess, requestedLimits.MaxProcess),
	}
	return limits

}

type CappedWriter struct {
	buf       bytes.Buffer
	limit     int
	written   int
	truncated bool
}

func NewCappedWriter(limit int) *CappedWriter {
	return &CappedWriter{limit: limit}
}

func (cw *CappedWriter) Write(p []byte) (n int, err error) {
	if cw.truncated {
		return len(p), nil // pretend so that segpipe is stopped
	}

	willWrite := len(p)
	if cw.written+willWrite > cw.limit {
		allowed := cw.limit - cw.written
		cw.buf.Write(p[:allowed])
		cw.buf.WriteString("[Output Truncated: Exceeded Buffer Limit]")
		cw.truncated = true

	} else {

		cw.buf.Write(p)
		cw.written += willWrite
	}

	return len(p), nil
}

func (cw *CappedWriter) String() string {
	return cw.buf.String()
}

func GetPeakMemory(cgroupPath string) int64 {
	peakFile := filepath.Join(cgroupPath, "memory.peak")
	data, err := os.ReadFile(peakFile)
	if err != nil {
		log.Printf("Warning: Could not read memory.peak at %s: %v", peakFile, err)
		return 0
	}

	bytes, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0
	}

	return bytes / 1024
}
