package articles

import (
	"path/filepath"
	"time"
)

type FileInfo struct {
	Path       []string
	FileName   string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func FileCrawler(dir string, path []string) <-chan FileInfo {
	fileInfoChan := make(chan FileInfo)

	go func() {
		defer close(fileInfoChan)

		files, err := ArticlesFS.ReadDir(dir)
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
				
				f, err := ArticlesFS.Open(filePath)
				if err != nil {
					continue
				}

				defer f.Close()

				fileInfo, err := f.Stat()
				
				if err != nil {
					continue
				}
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