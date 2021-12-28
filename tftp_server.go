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
	magic "github.com/hosom/gomagic"
	"github.com/pin/tftp"
	"github.com/spf13/viper"
)

func tftpReadHandler(filename string, rf io.ReaderFrom) error {
	cleanedPath := CleanPath(filename)

	fullPath := filepath.Join(viper.GetString("upload.dataDir"), cleanedPath)

	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("TFTP: failed to open file for reading: %v", err)
		return err
	}
	defer file.Close()

	n, err := rf.ReadFrom(file)
	if err != nil {
		log.Printf("TFTP: failed to read file: %v", err)
		return err
	}
	log.Printf("TFTP: %d bytes sent of %v", n, filename)
	return nil
}

func tftpWriteHandler(filename string, wf io.WriterTo) error {
	if !strings.HasPrefix(filename, viper.GetString("tftp.writePrefix")) {
		log.Printf("TFTP: forbidden filename: %v", filename)
		return fmt.Errorf("forbidden filename, must start with %v", viper.GetString("tftp.writePrefix"))
	}
	cleanedPath := CleanPath(filename)
	fullPath := filepath.Join(viper.GetString("upload.dataDir"), cleanedPath)
	if err := validateUploadFilename(cleanedPath); err != nil {
		log.Printf("TFTP: invalid filename %v: %v", cleanedPath, err)
		return err
	}
	dirPath := filepath.Dir(fullPath)
	os.MkdirAll(dirPath, 0777)
	os.Remove(fullPath + "._infocache")
	os.Remove(fullPath + "._infolock")
	f, err := os.Create(fullPath)
	if err != nil {
		log.Printf("TFTP: failed to create file %v: %v", cleanedPath, err)
		return err
	}
	defer f.Close()
	n, err := wf.WriteTo(f)
	if err != nil {
		log.Printf("TFTP: failed to write file %v: %v", cleanedPath, err)
		return err
	}
	log.Printf("TFTP: %d bytes received of %v", n, filename)

	fileType := ""
	m, err := magic.Open(magic.MAGIC_NONE)
	if err != nil {
		fileType = fmt.Sprintf("error while opening magic database: %v", err)
	} else {

		fileType, err = m.File(fullPath)
		if err != nil {
			fileType = fmt.Sprintf("error determining file type: %v", err)
		}
	}

	data := map[string]interface{}{
		"url":       viper.GetString("http.url") + "/" + cleanedPath,
		"sizeExact": n,
		"size":      datasize.ByteSize(n).HR(),
		"type":      fileType,
		"message":   fmt.Sprintf("File %v uploaded!", cleanedPath),
	}

	// notify all waiting listeners
	notifyFileListeners(cleanedPath)
	data["uploadedAt"] = time.Now()
	data["uploaderLocation"] = "TFTP"
	data["name"] = cleanedPath

	addToRecents(data)

	return nil
}

func runTftpServer() {
	if !viper.GetBool("tftp.enabled") {
		return
	}
	s := tftp.NewServer(tftpReadHandler, tftpWriteHandler)
	s.SetTimeout(5 * time.Second) // optional
	log.Printf("TFTP: started on port 69")
	err := s.ListenAndServe(":69") // blocks until s.Shutdown() is called
	if err != nil {
		log.Fatalf("TFTP: failed to start server: %v", err)

	}
}
