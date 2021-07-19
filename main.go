package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr           = flag.String("addr", "localhost:6060", "TCP address to listen to")
	useTls         = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert        = flag.String("cert", "", "Full certificate file path")
	tlsKey         = flag.String("key", "", "Full key file path")
	maxImgLength   = flag.Int("maximglength", 2000, "Maximum image height and width")
	maxBodySize    = flag.Int("maxbodysize", 100*1024*1024, "MaxRequestBodySize, defaults to 100MiB")
	browseImgColor = flag.String("browseimgcolor", "eaffe8", "Background color to use for browse page")
	debug          = flag.Bool("debug", false, "Enable debug logging")
	// TODO: Remove when auth added
	allowUpload = flag.Bool("allowupload", false, "Allow disabling upload (temporary)")

	images      = UpdateImageCache()
	rootPath    = []byte("/")
	cssPath     = []byte("/css/")
	svgPath     = []byte("/svg/")
	apiPath     = []byte("/api/")
	imgPath     = []byte("/images/")
	faviconPath = []byte("/favicon.ico")

	imgHandler = ImageHandler("www/content/images/", 1)

	rSrc = rand.NewSource(time.Now().Unix())
	rGen = rand.New(rSrc) // initialize local pseudorandom generator
)

func main() {
	flag.Parse()
	log.Print("- Loading srp-go")

	protocol := "http"
	if *useTls {
		protocol += "s"
	}

	checkMissingDirs()

	log.Printf("- Running srp-go on " + protocol + "://" + *addr)

	s := &fasthttp.Server{
		Handler:            requestHandler,
		Name:               "srp-go",
		MaxRequestBodySize: *maxBodySize,
	}

	if *useTls && len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		if err := s.ListenAndServeTLS(*addr, *tlsCert, *tlsKey); err != nil {
			log.Fatalf("- Error in ListenAndServeTLS: %s", err)
		}
	} else {
		if err := s.ListenAndServe(*addr); err != nil {
			log.Fatalf("- Error in ListenAndServe: %s", err)
		}
	}
}

// Main request handler
func requestHandler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	setCacheHeaders(ctx)
	handleDebug(ctx)

	switch {
	// Server 404 on /favicon.ico TODO: Add a default favicon
	case bytes.Equal(path, faviconPath):
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)

	// Handle css styles on /css/
	case bytes.HasPrefix(path, cssPath):
		content := GetCachedContent(ctx, cssMime)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}

	// Handle svg files on /svg/
	case bytes.HasPrefix(path, svgPath):
		content := GetCachedContent(ctx, svgMime)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}

	// Serve images on /images/
	case bytes.HasPrefix(path, imgPath):
		imgHandler(ctx)

	// Handle the api on /api/
	case bytes.HasPrefix(path, apiPath):
		HandleApi(ctx)

	// Default to serving html on all other paths
	default:
		content := GetCachedContent(ctx, htmlMime)

		if bytes.Equal(path, rootPath) {
			image := GetRandomImage()
			ctx.Response.Header.Set("X-Image-Hash", image)
			content = strings.Replace(content, "IMAGE_HASH", image, 2)
			content = strings.Replace(content, "#000000", GetColor(image), 1)
		}

		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}
	}
}

// setCacheHeaders sets the headers to avoid caching
func setCacheHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Pragma-Directive", "no-cache")
	ctx.Response.Header.Set("Cache-Directive", "no-cache")
	ctx.Response.Header.Set("Cache-Control", "no-store") // firefox ignores no-cache for this one
	ctx.Response.Header.Set("Pragma", "no-cache")
	ctx.Response.Header.Set("Expires", "0")
	ctx.Response.Header.Set("ETag", strconv.FormatInt(time.Now().UnixNano(), 10))
}

// handleDebug will print the debugging information on requests, including ctx.Path() and headers
func handleDebug(ctx *fasthttp.RequestCtx) {
	if *debug {
		fileCache = LoadAllCaches()       // re-read html caches for easier debugging, worse performance
		galleryCache = LoadGalleryCache() // re-create gallery html
		fmt.Println("")
		log.Printf("Path: %s", ctx.Path())
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			log.Printf("%v: %v", string(key), string(value))
		})
	}
}

// checkMissingDirs wil check for missing directories (side-effect from git, empty dirs don't get committed)
func checkMissingDirs() {
	// Check if tmp folder exists. Technically only needed for non-Docker testing
	if _, err := os.Stat("www/content/tmp/"); os.IsNotExist(err) {
		if err != nil {
			log.Printf("- Error checking for www/content/tmp/ folder: %v", err)
		}

		err = os.Mkdir("www/content/tmp/", os.FileMode(0700))
		if err != nil {
			log.Fatalf("- Error making www/content/tmp/ folder: %v", err)
			return
		}

		log.Printf("- Created www/content/tmp/ folder")
	}
}
