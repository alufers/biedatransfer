package main

import (
	"html/template"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func setupRoutes(r *gin.Engine) {
	r.LoadHTMLGlob("web/*.html")
	r.GET("/", func(c *gin.Context) {

		urlStr := viper.GetString("http.url")

		urlStr = strings.ReplaceAll(urlStr, "http://", `<span class="url-proto">http://</span>`)
		urlStr = strings.ReplaceAll(urlStr, "https://", `<span class="url-proto">https://</span>`)

		c.HTML(200, "index.html", map[string]interface{}{
			"URL":              template.HTML(urlStr),
			"TFTPPrefix":       template.HTML(viper.GetString("tftp.writePrefix")),
			"TFTPExamplesAddr": template.HTML(viper.GetString("tftp.examplesAddr")),
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
		if c.Request.Method == "PUT" || c.Request.Method == "POST" {
			handleUpload(c)
			return
		}
		if c.Request.Method == "GET" {
			handleDownload(c)
			return
		}
		sendError(c, 404, "Not found!")
	})
}
