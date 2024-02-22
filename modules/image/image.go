package image

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"coding-kittens.com/modules/cache"
	"github.com/gin-gonic/gin"
	"github.com/h2non/bimg"
)

// Define the cache directory
const cacheDir = "./cache/images"

// Define the cache instance
var imageCache = cache.NewFilesystemCache(cacheDir)

var urlPattern = regexp.MustCompile(`^web/static/assets/[^\.]+\.(jpeg|jpg|png|gif|webp|avif)$`)

var StaticAssets embed.FS

func GenerateCacheKey(imageURL string, width, height int, mimeType string) string {
    parts := strings.Split(mimeType, "/")

    return fmt.Sprintf("%s_%d_%d.%s", imageURL, width, height, parts[1])
}

func ProcessImage(c *gin.Context) {
	imagePath := filepath.Join("./web/", c.Query("url"))

    if !urlPattern.MatchString(imagePath) {
        c.JSON(400, gin.H{"error": "Invalid image URL"})
        return
    }
    
    imageBytes, err := StaticAssets.ReadFile(imagePath)
    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to read the image"})
        return
    }

	size, err := bimg.Size(imageBytes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to get original image dimensions"})
		return
	}

    widthStr := c.DefaultQuery("w", strconv.Itoa(size.Width))
    heightStr := c.Query("h")

    width, err := strconv.Atoi(widthStr)
    if err != nil || width <= 0 {
        c.JSON(400, gin.H{"error": "Invalid width"})
        return
    }

    var height int
    if heightStr != "" {
        height, err = strconv.Atoi(heightStr)
        if err != nil || height <= 0 {
            c.JSON(400, gin.H{"error": "Invalid height"})
            return
        }
    } else {
        // Calculate height to preserve aspect ratio
        height = (size.Height * width) / size.Width
    }

	qualityStr := c.DefaultQuery("q", "80")
    quality, err := strconv.Atoi(qualityStr)
    if err != nil || quality <= 0 || quality > 100 {
        c.JSON(400, gin.H{"error": "Invalid quality"})
        return
    }
    // TODO: add image/avif when it's supported
    options := []string{"image/webp", "image/jpeg", "image/png", "image/gif"}
    mimeType := getSupportedMimeType(options, c)

    if mimeType == "" {
        c.JSON(400, gin.H{"error": "No supported MIME type found"})
        return
    }

    cacheKey := GenerateCacheKey(imagePath, width, height, mimeType)

    if cachedImage, mimeType, err := imageCache.Get(cacheKey); err == nil {
        c.Data(200, mimeType, cachedImage)
        return
    }

    image, err := bimg.NewImage(imageBytes).Convert(getImageTypeFromString(mimeType))
    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to convert the image to the proper MIME type"})
        return
    }

    if bimg.NewImage(image).Type() == "png" {
        fmt.Fprintln(os.Stderr, "The image was converted into png")
      }

    resizedImage, err := bimg.Resize(image, bimg.Options{
        Width:  width,
        Height: height,
		Quality: quality,
    })

    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to resize the image"})
        return
    }

    imageCache.Set(cacheKey, resizedImage, mimeType, 7 * 24 * time.Hour)

    c.Data(200, mimeType, resizedImage)
}

func getImageTypeFromString(mimeType string) bimg.ImageType {
    mimeType = strings.TrimPrefix(mimeType, "image/")

    for imageType, typeName := range bimg.ImageTypes {
		if typeName == mimeType {
			return imageType
		}
	}
	// If the MIME type is not recognized, return a default value or handle the error
	return bimg.JPEG
}

func getSupportedMimeType(options []string, c *gin.Context) string {
	accept := c.GetHeader("Accept")
    mimeTypeQueryParam := c.Query("mimeType")

    if mimeTypeQueryParam != "" {
        for _, option := range options {
            if strings.Contains(mimeTypeQueryParam, option) {
                return option
            }
        }
    } 

	mimeType := mediaType(accept, options)

	if strings.Contains(accept, mimeType) {
		return mimeType
	}

	return ""
}

func mediaType(accept string, options []string) string {
	if accept == "" {
		return options[0]
	}

	accept = strings.ToLower(accept)

	for _, option := range options {
		if strings.Contains(accept, option) {
			return option
		}
	}

	return options[0]
}