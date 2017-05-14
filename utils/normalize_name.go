package utils

import (
	"regexp"
	"strings"

	"github.com/fatih/camelcase"
)

var (
	safeNameRE = regexp.MustCompile(`[^a-zA-Z0-9_]*$`)
)

func NormalizeName(name string) string {
	var normalizedName []string

	words := camelcase.Split(name)
	for _, word := range words {
		safeWord := strings.Trim(safeNameRE.ReplaceAllLiteralString(word, "_"), "_")
		lowerWord := strings.TrimSpace(strings.ToLower(safeWord))
		if lowerWord != "" {
			normalizedName = append(normalizedName, lowerWord)
		}
	}

	return strings.Join(normalizedName, "_")
}

func NormalizeDesc(desc string) string {
	return strings.Replace(desc, ".", "-", -1)
}
