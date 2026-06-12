package scanner

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/openclaw/MangaMetaManager/internal/metadata"
)

func ReadComicInfo(pathStr string) (*metadata.ComicInfo, error) {
	r, err := zip.OpenReader(pathStr)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var targetFile *zip.File
	for _, f := range r.File {
		if path.Base(f.Name) == "ComicInfo.xml" {
			targetFile = f
			// Prefer root level if multiple exist
			if f.Name == "ComicInfo.xml" {
				break
			}
		}
	}

	if targetFile != nil {
		rc, err := targetFile.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}

		var info metadata.ComicInfo
		if err := xml.Unmarshal(data, &info); err != nil {
			return nil, err
		}
		return &info, nil
	}

	return nil, nil // No ComicInfo.xml found
}

func WriteComicInfo(pathStr string, info *metadata.ComicInfo) error {
	// Ensure Komga compatibility namespaces
	info.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	info.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "mmm-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	zw := zip.NewWriter(tmpFile)
	
	// Open original file
	r, err := zip.OpenReader(pathStr)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find the best path for ComicInfo.xml
	xmlPath := "ComicInfo.xml"
	existingXmlFound := false
	
	// First pass to find existing XML or determine image folder
	dirCounts := make(map[string]int)
	for _, f := range r.File {
		if path.Base(f.Name) == "ComicInfo.xml" {
			xmlPath = f.Name
			existingXmlFound = true
			if f.Name == "ComicInfo.xml" {
				// Root is preferred, stop looking
				break
			}
		}
		if isImage(f.Name) {
			dir := path.Dir(f.Name)
			if dir == "." {
				dir = ""
			}
			dirCounts[dir]++
		}
	}

	if !existingXmlFound {
		maxCount := 0
		bestDir := ""
		for dir, count := range dirCounts {
			if count > maxCount {
				maxCount = count
				bestDir = dir
			}
		}
		if bestDir != "" {
			xmlPath = path.Join(bestDir, "ComicInfo.xml")
		}
	}

	// Copy all files except the one at xmlPath
	for _, f := range r.File {
		if f.Name == xmlPath {
			continue
		}
		
		w, err := zw.Create(f.Name)
		if err != nil {
			return err
		}
		
		rc, err := f.Open()
		if err != nil {
			return err
		}
		
		_, err = io.Copy(w, rc)
		rc.Close()
		if err != nil {
			return err
		}
	}

	// Add/Update ComicInfo.xml at determined path
	w, err := zw.Create(xmlPath)
	if err != nil {
		return err
	}
	
	data, err := xml.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	
	// Add XML header
	header := []byte(xml.Header)
	if _, err := w.Write(header); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}
	
	// Replace original file
	tmpFile.Close()
	return os.Rename(tmpFile.Name(), pathStr)
}

func isImage(name string) bool {
	ext := strings.ToLower(path.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".avif", ".gif":
		return true
	}
	return false
}

func IsArchive(pathStr string) bool {
	ext := strings.ToLower(filepath.Ext(pathStr))
	return ext == ".zip" || ext == ".cbz"
}
