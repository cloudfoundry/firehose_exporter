package transform

import (
	"regexp"
	"strings"

	"github.com/fatih/camelcase"
)

var (
	safeNameRE = regexp.MustCompile(`[^0-9A-Za-z]*$`)
)

func NormalizeName(name string) string {
	var normalizedName []string

	words := camelcase.Split(name)
	for _, word := range words {
		if word == "__" {
			normalizedName = append(normalizedName, "")
			continue
		}
		safeWord := strings.Trim(safeNameRE.ReplaceAllLiteralString(strings.Trim(word, "_"), "_"), "_")
		lowerWord := strings.TrimSpace(strings.ToLower(safeWord))
		if lowerWord != "" {
			normalizedName = append(normalizedName, lowerWord)
		}
	}

	return strings.Join(normalizedName, "_")
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
