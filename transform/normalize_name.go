package transform

import (
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

var (
	safeNameRE = regexp.MustCompile(`[^0-9A-Za-z]`)
)

func NormalizeName(name string) string {
	return strcase.ToSnake(safeNameRE.ReplaceAllLiteralString(name, "_"))
}

func NormalizeNameDesc(desc string) string {
	if strings.HasPrefix(desc, "/p.") {
		return "/p-" + desc[3:]
	}

	return desc
}

func NormalizeOriginDesc(desc string) string {
	return strings.ReplaceAll(desc, ".", "-")
}
