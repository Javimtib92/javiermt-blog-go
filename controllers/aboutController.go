package controllers

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"coding-kittens.com/models"
	"coding-kittens.com/modules/articles"
	"github.com/adrg/frontmatter"
	"github.com/gin-gonic/gin"
)

func AboutController (c *gin.Context) map[string]interface{} {
	fromDate := time.Date(2015, time.May, 1, 0, 0, 0, 0, time.UTC)
	
	today := time.Now()

	difference := today.Sub(fromDate)

	latestContent, err := articles.GetLatestContent(5)

	if err != nil {
		latestContent = []articles.FileInfo{}
	}

	var latestArticles []models.Article
	for _, content := range latestContent {
		var category string = content.Path[0];

		file, err := os.ReadFile("./web/_articles/" + category + "/" + content.FileName)
		if err != nil {
			fmt.Println("Error reading file:", err)
		}

		var matter models.FrontMatter

		frontmatter.MustParse(bytes.NewReader(file), &matter)

		fmt.Printf("%+v\n", matter)

		article := models.Article{
			Slug:      content.FileName,
			Category:  content.Path[0],
			Data: matter,
			
		}
		latestArticles = append(latestArticles, article)
	}

	return map[string]interface{}{
		"yearDiff": int64(difference.Hours()/24/365),
		"LatestContent": latestArticles,
	}
}