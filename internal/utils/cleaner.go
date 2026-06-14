package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var cleanRegex = regexp.MustCompile(`\[.*?\]|\(.*?\)|{.*?}`)

// BuildBookSearchQuery intelligently combines series title and book filename for scraping.
func BuildBookSearchQuery(seriesTitle, filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// 1. Remove bracketed tags
	name = cleanRegex.ReplaceAllString(name, " ")

	// 2. If the name contains the seriesTitle (case-insensitive), replace it with a space to prevent duplication.
	lowerName := strings.ToLower(name)
	lowerTitle := strings.ToLower(seriesTitle)
	if strings.Contains(lowerName, lowerTitle) {
		// Use regex for case-insensitive replacement if needed, 
		// but simple strings.Replace with indices or just stripping is fine for a query.
		start := strings.Index(lowerName, lowerTitle)
		name = name[:start] + " " + name[start+len(lowerTitle):]
	}

	// 3. Clean up separators and extra spaces
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune("-_.,", r) {
			return ' '
		}
		return r
	}, name)
	
	cleanedSub := strings.Join(strings.Fields(name), " ")
	
	if cleanedSub == "" {
		return seriesTitle
	}
	
	return seriesTitle + " " + cleanedSub
}
