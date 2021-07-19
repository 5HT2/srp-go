package main

import (
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

var (
	imageColors  map[string]string // image hash, color hex
	fileCache    = LoadAllCaches() // file path, file content
	galleryCache = LoadGalleryCache()
	cssMime      = "text/css; charset=utf-8"
	htmlMime     = "text/html; charset=utf-8"
	svgMime      = "image/svg+xml"
)

func GetColor(image string) string {
	if len(imageColors) == 0 { // TODO: we need some better way to handle this
		imageColors = make(map[string]string, len(images))
	}

	color := imageColors[image]

	if len(color) == 0 {
		newColor := MainImageColor("www/content/images/" + image)
		imageColors[image] = newColor
		return newColor
	}

	return color
}

// GetCachedContent returns content found inside the appropriate cache
func GetCachedContent(ctx *fasthttp.RequestCtx, mime string) string {
	ctx.Response.Header.Set("Content-Type", mime)

	path := string(ctx.Path())
	if path == "/" {
		path = "www/html/index.html"
	} else if mime == htmlMime {
		path = "www/html" + path + ".html"
	} else { // Other paths include their folder and file extension by default
		path = "www" + path
	}

	content := fileCache[path]
	if len(content) == 0 {
		log.Printf("- Error with finding cached content on path \"%s\"", path)
		ctx.Response.Header.Set("Content-Type", "text/plain")
		HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
		return ""
	}

	content = strings.ReplaceAll(content, "SERVER_NAME", string(ctx.Host()))
	content = strings.Replace(content, "var(--color-placeholder)", "#"+*browseImgColor, 1)
	content = strings.Replace(content, "ALL_GALLERY_ITEMS", galleryCache, 1)
	return content
}

// LoadAllCaches will read all the files in dir and return the map of path:content.
// The dir variable must have a slash suffix
func LoadAllCaches() map[string]string {
	allDirEntries := ReadDirsUnsafe("www/html/", "www/css/", "www/svg/")

	// Needed to make the initial map not-nil
	cache := make(map[string]string, 0)

	// Key format is "www/dir/file_name.ext", to simplify things
	for dir, entries := range allDirEntries {
		for _, file := range entries {
			path := dir + file.Name()
			cache[path] = ReadFileUnsafe(path)
		}
	}

	return cache
}

// LoadGalleryCache will format the list of images into a list of gallery html for each image we have
func LoadGalleryCache() string {
	template := "<a class=\"gallery-item\" data-src=\"\" data-sub-html=\"\"> <img class=\"img-responsive\" src=\"/images/IMAGE_HASH\" alt=\"IMAGE_HASH\"/> </a>"
	var galleryImages []string

	for _, img := range images {
		content := strings.Replace(template, "IMAGE_HASH", img, 2)
		galleryImages = append(galleryImages, content)
	}

	return strings.Join(galleryImages[:], "\n    ")
}
