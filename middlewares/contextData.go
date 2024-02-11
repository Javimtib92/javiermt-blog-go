package middlewares

import (
	"sync"

	"coding-kittens.com/modules/color"
	"coding-kittens.com/modules/utils"
	"github.com/gin-gonic/gin"
)

// ContextData holds additional data to be passed down the request handling chain
type ContextData struct {
	LiveReloadEnabled bool
	AccentBaseHSL     color.HSL
	// Add more fields as needed
}

var accentBaseOnce sync.Once
var accentBaseHSL color.HSL

// SetContextDataMiddleware is a Gin middleware that sets custom context data
func SetContextDataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure accentBaseHSL is calculated only once
		accentBaseOnce.Do(func() {
			accentBaseHSL, _ = color.HextoHSL(utils.GetAccentBaseValue())
		})

		// Create a custom context data struct
		contextData := ContextData{
			LiveReloadEnabled: gin.IsDebugging(),
			AccentBaseHSL:     accentBaseHSL,
		}

		// Set the custom context data
		c.Set("ContextData", contextData)
		c.Next()
	}
}