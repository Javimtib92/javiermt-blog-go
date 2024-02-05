package articles

import (
	"os"
	"path/filepath"
	"time"
)

// FileInfo represents information about a file.
type FileInfo struct {
	Path       []string
	FileName   string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

// FileCrawler is a generator function to crawl files in a directory.
func FileCrawler(dir string, path []string) <-chan FileInfo {
	fileInfoChan := make(chan FileInfo)

	go func() {
		defer close(fileInfoChan)

		files, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				subPath := append(path, file.Name())
				subDir := filepath.Join(dir, file.Name())
				fileInfoChan <- <-FileCrawler(subDir, subPath)
			} else {
				filePath := filepath.Join(dir, file.Name())
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					continue
				}

				fileInfoChan <- FileInfo{
					Path:       path,
					FileName:   file.Name(),
					CreatedAt:  fileInfo.ModTime(),
					ModifiedAt: fileInfo.ModTime(),
				}
			}
		}
	}()

	return fileInfoChan
}