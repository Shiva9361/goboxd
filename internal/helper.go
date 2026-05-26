package internal

import "strings"

func prepArgs(template []string, flags []string, source string, artifact string) []string {
	args := make([]string, 0, len(template)+len(flags))

	for _, arg := range args {
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
