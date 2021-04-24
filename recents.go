package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

var recentsLock = &sync.Mutex{}
var cachedRecents interface{}

func addToRecents(data interface{}) {

	currentData, err := getRecents()
	if err != nil {
		log.Printf("failed to getRecents: %v", err)
		currentData = []interface{}{}
	}
	recentsLock.Lock()
	defer recentsLock.Unlock()
	recentsPath := filepath.Join(viper.GetString("upload.dataDir"), "./recents.json")
	truncData := (currentData.([]interface{}))
	if len(truncData) > viper.GetInt("upload.recentsSize") {
		truncData = truncData[:viper.GetInt("upload.recentsSize")]
	}
	currentData = append([]interface{}{data}, truncData...)
	cachedRecents = currentData
	marshaled, err := json.Marshal(currentData)
	if err != nil {
		log.Printf("failed to marshal recents.json: %v", err)
		return
	}
	err = os.WriteFile(recentsPath, marshaled, 0777)
	if err != nil {
		log.Printf("failed to write recents.json: %v", err)
		return
	}
	notifyFileListeners("/recents.json")
}

func getRecents() (interface{}, error) {
	recentsLock.Lock()
	defer recentsLock.Unlock()
	if cachedRecents == nil {

		recentsPath := filepath.Join(viper.GetString("upload.dataDir"), "./recents.json")
		data, err := os.ReadFile(recentsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open %v: %w", recentsPath, data)
		}
		var output interface{}
		if err := json.Unmarshal(data, &output); err != nil {
			return nil, fmt.Errorf("failed to parse %v: %w", recentsPath, data)
		}
		cachedRecents = output
	}
	return cachedRecents, nil

}
