package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	magic "github.com/hosom/gomagic"
	"github.com/spf13/viper"
)

func handleFileInfo(c *gin.Context) {
	cleanedPath := CleanPath(c.Request.URL.Path)
	writePath := filepath.Join(viper.GetString("upload.dataDir"), cleanedPath)

	for {
		if _, err := os.Stat(writePath + ".infolock"); err != nil {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}

	// try returning the cached stuff
	if rawData, err := os.ReadFile(writePath + "._infocache"); err == nil {
		var cached map[string]interface{}
		json.Unmarshal(rawData, &cached)
		if cached != nil {
			sendWithFormat(c, 200, cached, map[string]interface{}{
				"PageType": "info",
			})
			return
		}
	}
	f, err := os.Create(writePath + ".infolock")
	if err != nil {
		sendError(c, 500, fmt.Sprintf("failed to create lock: %v", err))
		return
	}
	defer f.Close()
	infoToOutput := map[string]interface{}{}
	var fileType = "Unknown"
	m, err := magic.Open(magic.MAGIC_NONE)
	if err != nil {
		fileType = fmt.Sprintf("error while opening magic database: %v", err)
	} else {
		fileType, err = m.File(writePath)
		if err != nil {
			fileType = fmt.Sprintf("error determining file type: %v", err)
		}
	}
	infoToOutput["url"] = viper.GetString("http.url") + cleanedPath
	if stat, err := os.Stat(writePath); err == nil {
		infoToOutput["size"] = datasize.ByteSize(stat.Size()).HR()
		infoToOutput["sizeExact"] = stat.Size()
		infoToOutput["uploadedAt"] = stat.ModTime()
	}
	infoToOutput["name"] = cleanedPath
	infoToOutput["type"] = fileType
	cmd := exec.Command("binwalk", writePath)
	binwalkOutput, _ := cmd.CombinedOutput()
	infoToOutput["binwalk"] = string(binwalkOutput)

	exiftoolCmd := exec.Command("exiftool", "-j", writePath)
	exiftoolOutput, _ := exiftoolCmd.CombinedOutput()

	var exiftoolOutputParsed interface{} = nil
	json.Unmarshal(exiftoolOutput, &exiftoolOutputParsed)

	infoToOutput["exiftool"] = exiftoolOutputParsed

	lddCmd := exec.Command("ldd", writePath)
	lddOutput, _ := lddCmd.CombinedOutput()
	infoToOutput["ldd"] = string(lddOutput)

	if marshalled, err := json.Marshal(infoToOutput); err == nil {
		os.WriteFile(writePath+"._infocache", marshalled, 0777) // ignore errors, this is only a cache lol
	}
	os.Remove(writePath + ".infolock")
	sendWithFormat(c, 200, infoToOutput, map[string]interface{}{
		"PageType": "info",
	})
}
