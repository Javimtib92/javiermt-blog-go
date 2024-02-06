package image

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/h2non/bimg"
)

var ImageTypeFromString = reverseMap(bimg.ImageTypes)

func ProcessImage(c *gin.Context) {
    // Get the image URL from the query parameter
	imagePath := filepath.Join("./web/", c.Query("url"))

    // Fetch the image from the URL
    imageBytes, err := bimg.Read(imagePath)
    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to read the image"})
        return
    }

	// Extract original image width and height
	size, err := bimg.Size(imageBytes)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to get original image dimensions"})
		return
	}

    // Extract width and height from query parameters
    widthStr := c.DefaultQuery("w", strconv.Itoa(size.Width))
    heightStr := c.Query("h")

    width := atoi(widthStr)
    height := size.Height // Default to original height if height is not provided

    if heightStr != "" {
        height = atoi(heightStr)
    } else {
        // If height is not provided, calculate the corresponding height to preserve aspect ratio
        height = (size.Height * width) / size.Width
    }
	quality := c.DefaultQuery("q", "80")

    // Resize the image
    resizedImage, err := bimg.Resize(imageBytes, bimg.Options{
        Width:  width,
        Height: height,
		Quality: atoi(quality),
    })
    if err != nil {
        c.JSON(400, gin.H{"error": "Failed to resize the image"})
        return
    }

	// Get supported MIME types
    options := []string{"image/avif", "image/webp", "image/jpeg", "image/png", "image/gif"}
    mimeType := getSupportedMimeType(options, c)
    if mimeType == "" {
        c.JSON(400, gin.H{"error": "No supported MIME type found"})
        return
    }

    if !strings.Contains(mimeType, "image/") {
        img, err := bimg.NewImage(imageBytes).Convert(getImageTypeFromString(mimeType))
        if err != nil {
            c.JSON(400, gin.H{"error": "Failed to convert the image to the proper MIME type"})
            return
        }
        imageBytes = img
    }

    // Return the resized image
    c.Data(200, mimeType, resizedImage)
}

func reverseMap(originalMap map[bimg.ImageType]string) map[string]bimg.ImageType {
    reversedMap := make(map[string]bimg.ImageType)

    for key, value := range originalMap {
        reversedMap[value] = key
    }

    return reversedMap
}

func getImageTypeFromString(mimeType string) bimg.ImageType {
    // Lookup the ImageType in the map
    if imageType, ok := ImageTypeFromString[mimeType]; ok {
        return imageType
    }
    // If the MIME type is not recognized, return a default value or handle the error
    return bimg.JPEG
}

func getSupportedMimeType(options []string, c *gin.Context) string {
	// Get accept header from Gin context
	accept := c.GetHeader("Accept")

	// Find the best match between the supported mime types and the accept header
	mimeType := mediaType(accept, options)

	// Check if the mimeType is included in the accept header
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

func atoi(s string) int {
    i, _ := strconv.Atoi(s)
    return i
}

func itoa(i int) string {
	s := strconv.Itoa(i)
	return s
}