package articles

import (
	"sort"
	"time"
)

// GetAllArticles retrieves information about all articles based on a query.
func GetAllArticles(query map[string]string) []FileInfo {
	var articles []FileInfo

	for fileInfo := range FileCrawler("./web/_articles", nil) {
		category := fileInfo.Path[0]

		if query["category"] != "" && query["category"] != category {
			continue
		}

		articles = append(articles, FileInfo{
			Path:       fileInfo.Path,
			FileName:   fileInfo.FileName,
			CreatedAt:  fileInfo.CreatedAt,
			ModifiedAt: fileInfo.ModifiedAt,
		})
	}

	return articles
}

const DEFAULT_MAX_OUTPUT_COUNT = 5

func compareDesc(a, b time.Time) bool {
	return a.After(b)
}

// getLatestContent retrieves the latest content based on creation date.
func GetLatestContent(maxOutputCount int) ([]FileInfo, error) {
	query := map[string]string{} // Empty query for all articles

	allArticles := GetAllArticles(query)

	// Sort articles by creation time in descending order
	sort.Slice(allArticles, func(i, j int) bool {
		return compareDesc(allArticles[i].CreatedAt, allArticles[j].CreatedAt)
	})

	// Limit the output to the specified count or the default count
	outputCount := maxOutputCount
	if outputCount <= 0 || outputCount > len(allArticles) {
		outputCount = len(allArticles)
	}

	return allArticles[:outputCount], nil
}