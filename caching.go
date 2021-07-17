package main

import (
	"github.com/valyala/fasthttp"
	"os"
	"path/filepath"
	"strings"
)

var (
	imageColors map[string]string // image hash, color hex
	htmlCache   map[string]string // file path, html
)

func GetColor(image string) string {
	hash := strings.TrimSuffix(image, filepath.Ext(image))
	color := imageColors[hash]

	if len(color) == 0 {
		newColor := MainImageColor("www/images/" + image)
		imageColors[hash] = newColor
		return newColor
	}

	return color
}

// GetHtml returns html found inside htmlCache, with the path reformatted to "/my_path.html"
func GetHtml(ctx *fasthttp.RequestCtx) string {
	ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")

	path := string(ctx.Path())
	if path == "/" {
		path = "/index"
	}
	path += ".html"

	content := htmlCache[path]
	if len(content) == 0 {
		ctx.Response.Header.Set("Content-Type", "text/plain")
		HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
		return ""
	}

	content = strings.ReplaceAll(content, "SERVER_NAME", string(ctx.Host()))
	return content
}

// InitHtml will read all the files in www/html/ and set the keys in htmlCache to match
func InitHtml() {
	files, err := os.ReadDir("www/html/")
	if err != nil {
		panic(err) // This can't fail, and if it does, something is wrong with the user's env
	}

	// Needed to make the initial map not-nil
	htmlCache = make(map[string]string, len(files))

	// Key format is "/file_name.html", to simplify handling the original ctx.Path()
	for _, f := range files {
		path := "www/html/" + f.Name()
		htmlCache["/"+f.Name()] = ReadFileUnsafe(path)
	}
}
