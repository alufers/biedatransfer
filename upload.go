package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	magic "github.com/hosom/gomagic"
	"github.com/spf13/viper"
)

func handleUpload(c *gin.Context) {
	cleanedPath := CleanPath(c.Request.URL.Path)
	forbiddenNames := viper.GetStringSlice("upload.forbiddenNames")
	forbiddenPrefixes := viper.GetStringSlice("upload.forbiddenPrefixes")
	extension := filepath.Ext(strings.ToLower(cleanedPath))
	if extension == "._infocache" || extension == "._infolock" {
		sendError(c, 400, "Forbidden filename extension (._infocache)!")
		return
	}
	for _, n := range forbiddenNames {
		lowercase := strings.ToLower(cleanedPath)
		if strings.ToLower(n) == lowercase || strings.ToLower("/"+n) == lowercase {
			sendError(c, 400, "Forbidden filename!")
			return
		}
	}
	for _, n := range forbiddenPrefixes {
		lowercase := strings.ToLower(cleanedPath)
		if strings.HasPrefix(lowercase, strings.ToLower(n)) || strings.HasPrefix(lowercase, strings.ToLower("/"+n)) {
			sendError(c, 400, fmt.Sprintf("Forbidden filename prefix: %v!", n))
			return
		}
	}
	writePath := filepath.Join(viper.GetString("upload.dataDir"), cleanedPath)
	dirPath := filepath.Dir(writePath)
	os.MkdirAll(dirPath, 0777)
	os.Remove(writePath + "._infocache")
	os.Remove(writePath + "._infolock")
	f, err := os.Create(writePath)
	if err != nil {
		sendError(c, 500, fmt.Sprintf("failed to create file %v: %v", cleanedPath, err))
	}
	defer f.Close()
	fileSize, err := io.Copy(f, c.Request.Body)
	if err != nil {
		sendError(c, 500, fmt.Sprintf("failed to copy` file %v: %v", cleanedPath, err))
	}
	fileType := ""
	m, err := magic.Open(magic.MAGIC_NONE)
	if err != nil {
		fileType = fmt.Sprintf("error while opening magic database: %v", err)
	}
	fileType, err = m.File(writePath)
	if err != nil {
		fileType = fmt.Sprintf("error determining file type: %v", err)
	}
	data := map[string]interface{}{
		"url":       viper.GetString("http.url") + cleanedPath,
		"sizeExact": fileSize,
		"size":      datasize.ByteSize(fileSize).HR(),
		"type":      fileType,
		"message":   fmt.Sprintf("File %v uploaded!", cleanedPath),
	}
	sendWithFormat(c, 201, addCommandsToResponse(data, cleanedPath))

	// notify all waiting listeners
	notifyFileListeners(cleanedPath)
	data["uploadedAt"] = time.Now()
	data["uploaderLocation"] = "Unknown"
	data["name"] = cleanedPath
	func() {
		if ipLocationDB != nil {

			record, err := ipLocationDB.Get_all(c.ClientIP())
			if err != nil {
				log.Printf("failed to lookup IP location: %v", err)
				return
			}
			data["uploaderLocation"] = strings.Join([]string{record.City, record.Region, record.Country_short}, ", ")
		}
	}()

	addToRecents(data)
}
