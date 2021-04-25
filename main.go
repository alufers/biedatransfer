package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	magic "github.com/hosom/gomagic"
	"github.com/ip2location/ip2location-go"
	"github.com/spf13/viper"
)

var ipLocationDB *ip2location.DB

var writeConfig = false

func main() {
	flag.BoolVar(&writeConfig, "write-config", false, "writes the default config in the current PWD")
	flag.Parse()
	initConfig()
	var err error
	ipLocationDB, err = ip2location.OpenDB(viper.GetString("upload.locationDatabasePath"))
	if err != nil {
		log.Printf("failed to load IP location database file: %v", err)
	}
	r := gin.Default()
	gin.SetMode(viper.GetString("http.mode"))
	r.LoadHTMLGlob("web/*.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", map[string]interface{}{
			"URL": viper.GetString("http.url"),
		})
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("web/favicon.ico")
	})
	r.GET("/script.js", func(c *gin.Context) {
		c.File("web/script.js")
	})
	r.GET("/my-ip", func(c *gin.Context) {
		remoteIP, trusted := c.RemoteIP()
		sendWithFormat(c, 200, map[string]interface{}{
			"clientIP": c.ClientIP(),
			"remoteIP": remoteIP,
			"trusted":  trusted,
		})
	})
	r.GET("/recents.json", func(c *gin.Context) {
		if _, wait := c.GetQuery("wait"); wait {
			waitForFileChange(c, "/recents.json")
		}
		recents, err := getRecents()
		if err != nil {
			sendError(c, 500, err.Error())
			return
		}
		c.JSON(200, recents)
	})
	r.GET("/style.css", func(c *gin.Context) {
		c.File("web/style.css")
	})
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "PUT" {
			handleUpload(c)
			return
		}
		if c.Request.Method == "GET" {
			handleDownload(c)
			return
		}
		sendError(c, 404, "Not found!")
	})
	r.Run(viper.GetString("http.addr"))

}

var listenersMutex = &sync.Mutex{}
var listeners = map[string][]chan interface{}{}

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

func handleUpload(c *gin.Context) {
	cleanedPath := CleanPath(c.Request.URL.Path)
	forbiddenNames := viper.GetStringSlice("upload.forbiddenNames")
	extension := filepath.Ext(strings.ToLower(cleanedPath))
	if extension == "._infocache" || extension == "._infolock" {
		sendError(c, 400, "Forbidden filename extension (._infocache)!")
		return
	}
	for _, n := range forbiddenNames {
		lowercase := strings.ToLower(cleanedPath)
		if n == lowercase || "/"+n == lowercase {
			sendError(c, 400, "Forbidden filename!")
			return
		}
	}
	writePath := filepath.Join(viper.GetString("upload.dataDir"), cleanedPath)
	dirPath := filepath.Dir(writePath)
	os.MkdirAll(dirPath, 0777)
	os.Remove(writePath + "._infocache")
	os.Remove(writePath + "._infolock")
	f, err := os.Create(writePath)
	defer f.Close()
	if err != nil {
		sendError(c, 500, fmt.Sprintf("failed to create file %v: %v", cleanedPath, err))
	}
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

func notifyFileListeners(cleanedPath string) {
	listenersMutex.Lock()
	defer listenersMutex.Unlock()
	if listenersForThisFile, ok := listeners[cleanedPath]; ok {
		for _, l := range listenersForThisFile {
			select {
			case l <- nil:
			default:
			}
		}
	}
}

func waitForFileChange(c *gin.Context, cleanedPath string) {
	listenerChan := make(chan interface{})
	func() {
		listenersMutex.Lock()
		defer listenersMutex.Unlock()
		if _, ok := listeners[cleanedPath]; !ok {
			listeners[cleanedPath] = []chan interface{}{}
		}
		listeners[cleanedPath] = append(listeners[cleanedPath], listenerChan)
	}()
	select {
	case <-listenerChan:
	case <-c.Request.Context().Done():
		func() {
			listenersMutex.Lock()
			defer listenersMutex.Unlock()
			if listOfListeners, ok := listeners[cleanedPath]; ok {
				newList := make([]chan interface{}, 0, len(listOfListeners)-1)
				for _, l := range listOfListeners {
					if l != listenerChan {
						newList = append(newList, l)
					}
				}
				if len(newList) == 0 {
					delete(listeners, cleanedPath)
				} else {
					listeners[cleanedPath] = newList
				}
			}
		}()

		c.String(200, "cancelled")
		return
	}
}
