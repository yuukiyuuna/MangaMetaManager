package scanner

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
)

var (
	fileLocks = make(map[string]*sync.Mutex)
	locksMu   sync.Mutex
)

func getFileLock(path string) *sync.Mutex {
	locksMu.Lock()
	defer locksMu.Unlock()
	if l, ok := fileLocks[path]; ok {
		return l
	}
	l := &sync.Mutex{}
	fileLocks[path] = l
	return l
}

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

func WriteComicInfo(pathStr string, info *metadata.ComicInfo, backup bool) error {
	lock := getFileLock(pathStr)
	lock.Lock()
	defer lock.Unlock()

	if backup {
		if err := backupFile(pathStr); err != nil {
			// Log error but continue? Or fail? Let's log it.
			fmt.Printf("Warning: failed to create backup for %s: %v\n", pathStr, err)
		}
	}

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

	// Hardcode root path for ComicInfo.xml to ensure Komga compatibility
	xmlPath := "ComicInfo.xml"
	
	// Collision tracking
	usedNames := make(map[string]bool)
	usedNames[xmlPath] = true // Reserve the XML name

	// Copy all files from original archive, flattening the structure
	for _, f := range r.File {
		// Skip directories and existing ComicInfo.xml
		if f.FileInfo().IsDir() || strings.ToLower(path.Base(f.Name)) == "comicinfo.xml" {
			continue
		}
		
		baseName := path.Base(f.Name)
		newName := baseName

		// Handle collisions if root already has this filename
		if usedNames[newName] {
			// Try prepending the immediate parent directory name (e.g., vol1_01.jpg)
			dirName := path.Base(path.Dir(f.Name))
			if dirName != "." && dirName != "" && dirName != "/" {
				newName = fmt.Sprintf("%s_%s", dirName, baseName)
			}
			
			// If still colliding, append a counter
			counter := 1
			for usedNames[newName] {
				ext := path.Ext(baseName)
				nameWithoutExt := strings.TrimSuffix(baseName, ext)
				newName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
				counter++
			}
		}

		usedNames[newName] = true

		w, err := zw.Create(newName)
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

	// Add/Update ComicInfo.xml at the root
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
	r.Close() // Close before rename for Windows compatibility (though on Linux it's less critical)
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

func backupFile(src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	dst := src + ".bak"
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}
