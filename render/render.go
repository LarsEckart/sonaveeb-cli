package render

import (
	"fmt"
	"strings"

	"github.com/lars/sonaveeb-cli/sonaveeb"
)

func RenderOutput(output sonaveeb.Output, quiet bool) string {
	var builder strings.Builder

	if !quiet && output.Header != "" {
		builder.WriteString(output.Header)
		builder.WriteByte('\n')
	}

	if !quiet && len(output.Translations) > 0 {
		_, _ = fmt.Fprintf(&builder, "  English: %s\n", strings.Join(output.Translations, ", "))
	}

	for _, line := range output.Lines {
		if quiet {
			_, _ = fmt.Fprintf(&builder, "%s\t%s\n", line.Code, line.Value)
			continue
		}
		_, _ = fmt.Fprintf(&builder, "  %-45s %s\n", line.Label+":", line.Value)
	}

	return builder.String()
}
