package internal

import (
	"cmp"
	"strings"
)

func prepArgs(template []string, flags []string, source string, artifact string, flag_allowlist []string) []string {
	args := make([]string, 0, len(template)+len(flags))

	if len(flag_allowlist) > 0 { // sanitize if we have an allowlist else assuming all valid
		allowedFlags := make(map[string]bool)
		for _, allowedFlag := range flag_allowlist {
			allowedFlags[allowedFlag] = true
		}

		filteredFlags := make([]string, 0)
		for _, flag := range flags {
			if allowedFlags[flag] {
				filteredFlags = append(filteredFlags, flag)
			}
		}
		flags = filteredFlags
	}

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

func sanitizeLimits(defaultLimits Resource, requestedLimits Resource) Resource {
	limits := Resource{
		WallTime:   clampResource(defaultLimits.WallTime, requestedLimits.WallTime), // cool way to write instead of if else
		Memory:     clampResource(defaultLimits.Memory, requestedLimits.Memory),
		MaxProcess: clampResource(defaultLimits.MaxProcess, requestedLimits.MaxProcess),
	}
	return limits

}
