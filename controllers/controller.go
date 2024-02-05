package controllers

import "github.com/gin-gonic/gin"

type ControllerFunc func(c *gin.Context) map[string]interface{}