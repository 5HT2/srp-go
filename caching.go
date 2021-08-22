package main

import (
	"github.com/mat/besticon/ico"
	"github.com/valyala/fasthttp"
	"image"
	"image/png"
	"log"
	"os"
	"strings"
)

var (
	imageColors  map[string]string // image hash, color hex
	fileCache    = LoadAllCaches() // file path, file content
	imageCache   = LoadImageCache()
	galleryCache = LoadGalleryCache()
	faviconCache = LoadFaviconCache(customFaviconPath)
	cssMime      = "text/css; charset=utf-8"
	htmlMime     = "text/html; charset=utf-8"
	jsonMime     = "application/json"
	svgMime      = "image/svg+xml"

	defaultFaviconPath = "www/ico/favicon.ico"
	customFaviconPath  = "config/favicon.ico"

	browseImgColor = "" // Set in main.go after env is parsed
)

func GetColor(image string) string {
	if len(imageColors) == 0 { // TODO: we need some better way to handle this
		imageColors = make(map[string]string, len(imageCache))
	}

	color := imageColors[image]

	if len(color) == 0 {
		newColor := MainImageColor("config/images/" + image)
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
		if *debug {
			log.Printf("- Error with finding cached content on path \"%s\"", path)
		}
		ctx.Response.Header.Set("Content-Type", "text/plain")
		HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
		return ""
	}

	// Unsure if I can get rid of this somehow... seems that you can change the window title with JS but that's it
	content = strings.ReplaceAll(content, "SERVER_NAME", string(ctx.Host()))
	// TODO: Find a way to replace this
	content = strings.Replace(content, "var(--color-placeholder)", browseImgColor, 1)
	content = strings.Replace(content, "ALL_GALLERY_ITEMS", galleryCache, 1)
	return content
}

// HandleCachedFavicon will return the favicon bytes to the client
func HandleCachedFavicon(ctx *fasthttp.RequestCtx) {
	if faviconCache == nil {
		HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
		return
	}

	ctx.Response.Header.Set(fasthttp.HeaderContentType, "image/x-icon")
	_ = png.Encode(ctx.Response.BodyWriter(), faviconCache)
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

// LoadGalleryCache will format the list of imageCache into a list of gallery html for each image we have
func LoadGalleryCache() string {
	template := "<a class=\"gallery-item\" data-src=\"\" data-sub-html=\"\"> <img class=\"img-responsive\" src=\"/images/IMAGE_HASH\" alt=\"IMAGE_HASH\"/> </a>"
	var galleryImages []string

	for _, img := range imageCache {
		content := strings.Replace(template, "IMAGE_HASH", img, 2)
		galleryImages = append(galleryImages, content)
	}

	return strings.Join(galleryImages[:], "\n    ")
}

// LoadImageCache opens the www/images directory and caches a list of FileInfos
func LoadImageCache() []string {
	dir := "config/images/"
	dirOpen, _ := os.Open(dir)
	tmpImages, err := dirOpen.Readdir(0)
	if err != nil {
		panic(err)
	}

	var imageNames []string
	for _, f := range tmpImages {
		imageNames = append(imageNames, f.Name())
	}

	return imageNames
}

// LoadFaviconCache tries to load config/favicon.ico and defaults to www/ico/favicon.ico
func LoadFaviconCache(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Error loading icon file: %s", err)
		if path != defaultFaviconPath { // Prevent infinite loop by calling itself. Less duplicated code
			return LoadFaviconCache(defaultFaviconPath)
		}
		return nil
	}
	defer f.Close()
	img, err := ico.Decode(f)
	if err != nil {
		log.Printf("Error loading icon: %s", err)
		return nil
	}

	return img
}
