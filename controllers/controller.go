package controllers

import (
	"embed"

	"github.com/gin-gonic/gin"
)

type ControllerFunc func(c *gin.Context) map[string]interface{}

var ArticlesFS embed.FS