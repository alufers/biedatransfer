package main

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var listenersMutex = &sync.Mutex{}
var listeners = map[string][]chan interface{}{}

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
