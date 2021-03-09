package transform

import (
	"github.com/iancoleman/strcase"
	"regexp"
	"strings"
)

var (
	safeNameRE = regexp.MustCompile(`[^0-9A-Za-z]`)
)

func NormalizeName(name string) string {

	return strcase.ToSnake(safeNameRE.ReplaceAllLiteralString(name, "_"))
}

func NormalizeNameDesc(desc string) string {
	if strings.HasPrefix(desc, "/p.") {
		return "/p-" + desc[3:len(desc)]
	}

	return desc
}

func NormalizeOriginDesc(desc string) string {
	return strings.Replace(desc, ".", "-", -1)
}
