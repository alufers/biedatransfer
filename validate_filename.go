package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func validateUploadFilename(cleanedPath string) error {
	forbiddenNames := viper.GetStringSlice("upload.forbiddenNames")
	forbiddenPrefixes := viper.GetStringSlice("upload.forbiddenPrefixes")
	extension := filepath.Ext(strings.ToLower(cleanedPath))
	if extension == "._infocache" || extension == "._infolock" {
		return fmt.Errorf("forbidden filename extension (._infocache)")
	}
	for _, n := range forbiddenNames {
		lowercase := strings.ToLower(cleanedPath)
		if strings.ToLower(n) == lowercase || strings.ToLower("/"+n) == lowercase {
			return fmt.Errorf("forbidden filename")
		}
	}
	for _, n := range forbiddenPrefixes {
		lowercase := strings.ToLower(cleanedPath)
		if strings.HasPrefix(lowercase, strings.ToLower(n)) || strings.HasPrefix(lowercase, strings.ToLower("/"+n)) {

			return fmt.Errorf("forbidden filename prefix: %v", n)
		}
	}
	return nil
}
