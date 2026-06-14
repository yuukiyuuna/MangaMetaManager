package utils

import (
	"regexp"
	"strings"
)

var cleanRegex = regexp.MustCompile(`\[.*?\]|\(.*?\)|{.*?}`)

// BuildBookSearchQuery intelligently combines series title and book filename for scraping.
func BuildBookSearchQuery(seriesTitle, filename string) string {
	// 1. Strip common archive extensions
	name := filename
	lower := strings.ToLower(filename)
	for _, ext := range []string{".zip", ".cbz", ".rar", ".7z", ".pdf", ".tar", ".cbr"} {
		if strings.HasSuffix(lower, ext) {
			name = filename[:len(filename)-len(ext)]
			break
		}
	}

	// 2. Remove bracketed tags
	name = cleanRegex.ReplaceAllString(name, " ")

	// 3. If the name contains the seriesTitle (case-insensitive), replace it with a space to prevent duplication.
	lowerName := strings.ToLower(name)
	lowerTitle := strings.ToLower(seriesTitle)
	if seriesTitle != "" && strings.Contains(lowerName, lowerTitle) {
		start := strings.Index(lowerName, lowerTitle)
		name = name[:start] + " " + name[start+len(lowerTitle):]
	}

	// 4. Clean up separators and extra spaces
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune("-_.,", r) {
			return ' '
		}
		return r
	}, name)
	
	cleanedSub := strings.Join(strings.Fields(name), " ")
	
	if seriesTitle == "" {
		return cleanedSub
	}
	
	if cleanedSub == "" {
		return seriesTitle
	}
	
	return seriesTitle + " " + cleanedSub
}
