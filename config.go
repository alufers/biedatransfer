package main

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetDefault("http.addr", ":8080")
	viper.SetDefault("http.url", "http://localhost:8080")
	viper.SetDefault("http.mode", "debug")
	viper.SetDefault("http.trustedProxies", []string{"127.0.0.0/24"})
	viper.SetDefault("upload.dataDir", "./data")
	viper.SetDefault("upload.locationDatabasePath", "./IP2LOCATION-LITE-DB5.BIN")
	viper.SetDefault("upload.recentsSize", 40)
	viper.SetDefault("upload.forbiddenNames", []string{
		"index.html",
		"index.htm",
		"robots.txt",
		"humans.txt",
		"favicon.ico",
		"wp-admin.php",
		"xmlrpc.php",
		".env",
		".git",
		".config",
		"recents.json",
		".",
		"/",
		"./",
		"style.css",
	})
	viper.SetDefault("upload.forbiddenPrefixes", []string{
		".well-known", // letsencrypt site verification
		".git",
		".htaccess",
		".htpasswd",
		"google", // google website verifification
	})
	viper.SetEnvPrefix("biedatransfer")
	viper.SetConfigName("biedatransfer-config") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // optionally look for config in the working directory
	viper.AddConfigPath("/etc/biedatransfer")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Printf("failed to read config file: %v", err)
	}
	if writeConfig {
		log.Print(viper.SafeWriteConfig())
		os.Exit(0)
	}
}
