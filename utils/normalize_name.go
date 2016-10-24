package utils

import (
	"strings"

	"github.com/fatih/camelcase"
)

func NormalizeName(name string) string {
	var normalizedName []string

	words := camelcase.Split(name)
	for _, word := range words {
		if word != "." && word != "_" && word != "-" {
			lowerWord := strings.ToLower(word)
			normalizedName = append(normalizedName, strings.Replace(lowerWord, "/", ":", -1))
		}
	}
	return strings.Join(normalizedName, "_")
}
