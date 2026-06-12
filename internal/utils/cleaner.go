package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var cleanRegex = regexp.MustCompile(`\[.*?\]|\(.*?\)|{.*?}`)

// CleanQuery removes bracketed tags and file extensions from a filename to make it suitable for search.
func CleanQuery(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	
	cleaned := cleanRegex.ReplaceAllString(name, " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ") // remove extra spaces
	
	if cleaned == "" {
		return name
	}
	return cleaned
}
