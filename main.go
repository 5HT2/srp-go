package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr         = flag.String("addr", "localhost:6060", "TCP address to listen to")
	useTls       = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert      = flag.String("cert", "", "Full certificate file path")
	tlsKey       = flag.String("key", "", "Full key file path")
	maxImgLength = flag.Int("maximglength", 2000, "Maximum image height and width")
	maxBodySize  = flag.Int("maxbodysize", 100*1024*1024, "MaxRequestBodySize, defaults to 100MiB")
	debug        = flag.Bool("debug", false, "Enable debug logging")

	images      = UpdateImageCache()
	rootPath    = []byte("/")
	cssPath     = []byte("/css/")
	apiPath     = []byte("/api/")
	imgPath     = []byte("/images/")
	faviconPath = []byte("/favicon.ico")

	imgHandler = ImageHandler("www/images/", 1)

	rSrc = rand.NewSource(time.Now().Unix())
	rGen = rand.New(rSrc) // initialize local pseudorandom generator
)

func main() {
	flag.Parse()
	log.Print("- Loading srp-go")
	MkImageFolder()

	protocol := "http"
	if *useTls {
		protocol += "s"
	}

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

	// Serve images on /images/
	case bytes.HasPrefix(path, imgPath):
		imgHandler(ctx)

	// Handle css styles on /css/
	case bytes.HasPrefix(path, cssPath):
		content := GetCachedContent(ctx, "ctx", false)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}

	// Handle the api on /api/
	case bytes.HasPrefix(path, apiPath):
		HandleApi(ctx)

	// Default to serving html on all other paths
	default:
		content := GetCachedContent(ctx, "html", true)

		if bytes.Equal(path, rootPath) {
			image, hash := GetRandomImage()
			ctx.Response.Header.Set("X-Image-Hash", hash)
			content = strings.Replace(content, "IMAGE_LINK", image, 1)
			content = strings.Replace(content, "IMAGE_HASH", hash, 1)
			content = strings.Replace(content, "#000000", GetColor(image, hash), 1)
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
		fmt.Println("")
		log.Printf("Path: %s", ctx.Path())
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			log.Printf("%v: %v", string(key), string(value))
		})
	}
}
