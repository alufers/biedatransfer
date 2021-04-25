package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
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

	if ipLocationDB, err = ip2location.OpenDB(viper.GetString("upload.locationDatabasePath")); err != nil {
		log.Printf("failed to load IP location database file: %v", err)
	}
	r := gin.Default()
	gin.SetMode(viper.GetString("http.mode"))
	setupRoutes(r)
	r.Run(viper.GetString("http.addr"))

}
