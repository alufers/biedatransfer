package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func handleDownload(c *gin.Context) {
	cleanedPath := CleanPath(c.Request.URL.Path)
	if _, wait := c.GetQuery("wait"); wait {
		waitForFileChange(c, cleanedPath)
	}
	fullPath := filepath.Join(viper.GetString("upload.dataDir"), cleanedPath)
	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		sendError(c, 404, fmt.Sprintf("File %v not found!", c.Request.URL.Path))
		return
	}
	if _, wait := c.GetQuery("info"); wait {
		handleFileInfo(c)
		return
	}

	c.File(fullPath)
}
