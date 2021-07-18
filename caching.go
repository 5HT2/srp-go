package main

import (
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

var (
	imageColors map[string]string // image hash, color hex
	fileCache   = ReadAllFiles()  // file path, file content
)

func GetColor(image string) string {
	if len(imageColors) == 0 { // TODO: we need some better way to handle this
		imageColors = make(map[string]string, len(images))
	}

	color := imageColors[image]

	if len(color) == 0 {
		newColor := MainImageColor("www/images/" + image)
		imageColors[image] = newColor
		return newColor
	}

	return color
}

// GetCachedContent returns content found inside the appropriate cache
func GetCachedContent(ctx *fasthttp.RequestCtx, mime string, html bool) string {
	ctx.Response.Header.Set("Content-Type", "text/"+mime+"; charset=utf-8")

	path := string(ctx.Path())
	if path == "/" {
		path = "www/html/index.html"
	} else if html {
		path = "www/html" + path + ".html"
	} else { // Other paths include their folder and file extension by default
		path = "www" + path
	}

	content := fileCache[path]
	if len(content) == 0 {
		log.Println(path)
		ctx.Response.Header.Set("Content-Type", "text/plain")
		HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
		return ""
	}

	content = strings.ReplaceAll(content, "SERVER_NAME", string(ctx.Host()))
	content = strings.Replace(content, "AAAAAA", *browseImgColor, 1)
	return content
}

// ReadAllFiles will read all the files in dir and return the map of path:content.
// The dir variable must have a slash suffix
func ReadAllFiles() map[string]string {
	filesHtml := ReadDirUnsafe("www/html/")
	filesCss := ReadDirUnsafe("www/css/")

	// Needed to make the initial map not-nil
	cache := make(map[string]string, len(filesHtml)+len(filesCss))

	// Key format is "www/dir/file_name.ext", to simplify things
	for _, f := range filesHtml {
		path := "www/html/" + f.Name()
		cache[path] = ReadFileUnsafe(path)
	}
	for _, f := range filesCss {
		path := "www/css/" + f.Name()
		cache[path] = ReadFileUnsafe(path)
	}

	return cache
}
