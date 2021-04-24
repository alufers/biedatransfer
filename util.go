package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/logrusorgru/aurora"
)

// CleanPath makes a path safe for use with filepath.Join. This is done by not
// only cleaning the path, but also (if the path is relative) adding a leading
// '/' and cleaning it (then removing the leading '/'). This ensures that a
// path resulting from prepending another path will always resolve to lexically
// be a subdirectory of the prefixed path. This is all done lexically, so paths
// that include symlinks won't be safe as a result of using CleanPath.
func CleanPath(path string) string {
	// Deal with empty strings nicely.
	if path == "" {
		return ""
	}

	// Ensure that all paths are cleaned (especially problematic ones like
	// "/../../../../../" which can cause lots of issues).
	path = filepath.Clean(path)

	// If the path isn't absolute, we need to do more processing to fix paths
	// such as "../../../../<etc>/some/path". We also shouldn't convert absolute
	// paths to relative ones.
	if !filepath.IsAbs(path) {
		path = filepath.Clean(string(os.PathSeparator) + path)
		// This can't fail, as (by definition) all paths are relative to root.
		path, _ = filepath.Rel(string(os.PathSeparator), path)
	}

	// Clean the path again for good measure.
	return filepath.Clean(path)
}

func sendError(c *gin.Context, status int, message string) {
	sendWithFormat(c, status, map[string]interface{}{
		"status":  status,
		"message": message,
	})
}

func sendWithFormat(c *gin.Context, status int, data map[string]interface{}, extraTemplateData ...map[string]interface{}) {
	accept := strings.Split(strings.ToLower(c.GetHeader("Accept")), ",")[0]
	if _, forceJson := c.GetQuery("json"); forceJson {
		accept = "application/json"
	}
	switch accept {
	case "application/json":
		fallthrough
	case "text/json":
		fallthrough
	case "json":
		c.JSON(status, data)
		return
	case "text/html":
		fallthrough
	case "application/xhtml+xml":
		tplData := map[string]interface{}{
			"Data":     data,
			"Status":   status,
			"PageType": "unknown",
		}
		for _, e := range extraTemplateData {
			for k, v := range e {
				tplData[k] = v
			}
		}
		c.HTML(status, "with-format.html", tplData)
		return
	}

	if strings.HasPrefix(c.GetHeader("User-Agent"), "curl/") {
		colored := ""
		keys := make([]string, 0)
		data = collapseSizes(data)
		for k, _ := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := data[k]
			colored += aurora.Yellow(k).String()
			colored += aurora.Gray(10, ": ").String()
			switch typed := v.(type) {
			case string:
				colored += aurora.Green(typed).String()
			case byte:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			case int:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			case int64:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			case uint:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			case uint64:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			case float32:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			case float64:
				colored += aurora.Cyan(fmt.Sprintf("%v", typed)).String()
			default:
				marsh, _ := json.Marshal(v)
				colored += aurora.Reset(string(marsh)).String()
			}
			colored += "\n"

		}
		c.String(status, colored)
		return
	}

	c.JSON(status, data)
}

func addCommandsToResponse(data map[string]interface{}, cleanedPath string) (out map[string]interface{}) {
	basename := filepath.Base(cleanedPath)
	out = map[string]interface{}{
		"cmdDownload": fmt.Sprintf("curl %v -o %v", data["url"], basename),
	}
	for k, v := range data {
		out[k] = v
	}
	return
}

func collapseSizes(data map[string]interface{}) (out map[string]interface{}) {

	out = map[string]interface{}{}
	for k, v := range data {
		out[k] = v
	}
	if _, ok := out["size"]; ok {
		out["size"] = fmt.Sprintf("%v (%v B)", out["size"], out["sizeExact"])
		delete(out, "sizeExact")
	}
	return
}
