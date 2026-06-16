package utils

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	cleanRegex = regexp.MustCompile(`\[.*?\]|\(.*?\)|{.*?}`)
	volRegex   = regexp.MustCompile(`(?i)(?:v|vol|volume|第|卷)\s*\.?\s*(\d+(?:\.\d+)?)`)
	numRegex   = regexp.MustCompile(`(\d+(?:\.\d+)?)`)
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// CamelToSnake converts a camelCase string to snake_case
func CamelToSnake(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// ParseVolumeNumber extracts a volume number (e.g. 1, 2.5) from a string.
// Returns -1 if no number is found.
func ParseVolumeNumber(input string) float64 {
	// Clean extensions
	name := input
	for _, ext := range []string{".zip", ".cbz", ".rar", ".7z", ".pdf", ".tar", ".cbr"} {
		if strings.HasSuffix(strings.ToLower(input), ext) {
			name = input[:len(input)-len(ext)]
			break
		}
	}

	// 1. Try explicit patterns
	match := volRegex.FindStringSubmatch(name)
	if len(match) > 1 {
		if val, err := strconv.ParseFloat(match[1], 64); err == nil {
			return val
		}
	}

	// 2. Try last number in string
	matches := numRegex.FindAllStringSubmatch(name, -1)
	if len(matches) > 0 {
		last := matches[len(matches)-1][1]
		if val, err := strconv.ParseFloat(last, 64); err == nil {
			return val
		}
	}

	return -1
}

// SimpleSimilarity computes a 0.0-1.0 similarity score between two strings
// using a basic character overlap (Sorensen-Dice coefficient style).
func SimpleSimilarity(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Tokenize
	tokensA := strings.Fields(a)
	tokensB := strings.Fields(b)
	
	common := 0
	for _, tA := range tokensA {
		for _, tB := range tokensB {
			if tA == tB {
				common++
				break
			}
		}
	}

	return 2.0 * float64(common) / float64(len(tokensA)+len(tokensB))
}

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
